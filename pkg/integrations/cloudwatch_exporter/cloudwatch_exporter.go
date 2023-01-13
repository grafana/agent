package cloudwatch_exporter

import (
	"context"
	"fmt"
	"github.com/grafana/agent/pkg/integrations"

	"github.com/go-kit/log"
	yace "github.com/nerdswords/yet-another-cloudwatch-exporter/pkg"
	yaceConf "github.com/nerdswords/yet-another-cloudwatch-exporter/pkg/config"
	yaceLog "github.com/nerdswords/yet-another-cloudwatch-exporter/pkg/logger"
	yaceModel "github.com/nerdswords/yet-another-cloudwatch-exporter/pkg/model"
	yaceSvc "github.com/nerdswords/yet-another-cloudwatch-exporter/pkg/services"
	yaceSess "github.com/nerdswords/yet-another-cloudwatch-exporter/pkg/session"
	"github.com/prometheus/client_golang/prometheus"
)

const (
	metricsPerQuery       = 500
	cloudWatchConcurrency = 5
	tagConcurrency        = 5
	labelsSnakeCase       = false
)

func init() {
	integrations.RegisterIntegration(&Config{})
}

func (c Config) ToYACEConfig() (yaceConf.ScrapeConf, error) {
	var nilToZero = true
	discoveryJobs := []*yaceConf.Job{}
	for _, job := range c.Jobs {
		lengthSeconds := int64(job.ScrapeInterval.Seconds())
		periodSeconds := lengthSeconds
		roundingPeriod := lengthSeconds
		roles := []yaceConf.Role{}
		for _, role := range job.Roles {
			roles = append(roles, yaceConf.Role{
				RoleArn:    role.RoleArn,
				ExternalID: role.ExternalID,
			})
		}
		metrics := []*yaceConf.Metric{}
		for _, metric := range job.Metrics {
			metrics = append(metrics, &yaceConf.Metric{
				Name:       metric.Name,
				Statistics: metric.Statistics,
				Period:     periodSeconds,
				Length:     lengthSeconds,
			})
		}
		discoveryJobs = append(discoveryJobs, &yaceConf.Job{
			Regions:        job.Regions,
			Type:           job.Type,
			Roles:          roles,
			NilToZero:      &nilToZero,
			RoundingPeriod: &roundingPeriod,
			Metrics:        metrics,
		})
	}
	conf := yaceConf.ScrapeConf{
		StsRegion: c.STSRegion,
		Discovery: yaceConf.Discovery{
			ExportedTagsOnMetrics: nil,
			Jobs:                  discoveryJobs,
		},
	}
	return conf, conf.Validate(yaceSvc.CheckServiceName)
}

func (c *Config) Name() string {
	return "cloudwatch_exporter"
}

// todo: is this used when there's more than one integration instnace in an agent config
func (c *Config) InstanceKey(agentKey string) (string, error) {
	return c.Name(), nil
}

func (c *Config) NewIntegration(l log.Logger) (integrations.Integration, error) {
	exporterConfig, err := c.ToYACEConfig()
	if err != nil {
		return nil, fmt.Errorf("invalid cloudwatch exporter configuration: %w", err)
	}
	collector := newCloudwatchCollector(l, exporterConfig)

	l.Log("msg", "creating new cloudwatch integration")

	return integrations.NewCollectorIntegration(
		c.Name(),
		integrations.WithCollectors(collector),
	), nil
}

// cloudwatchCollector
type cloudwatchCollector struct {
	logger       yaceLoggerWrapper
	sessionCache yaceSess.SessionCache
	scrapeConf   yaceConf.ScrapeConf
}

func newCloudwatchCollector(logger log.Logger, conf yaceConf.ScrapeConf) *cloudwatchCollector {
	loggerWrapper := yaceLoggerWrapper{
		debug: false,
		log:   logger,
	}
	return &cloudwatchCollector{
		logger:       loggerWrapper,
		sessionCache: yaceSess.NewSessionCache(conf, true, loggerWrapper),
		scrapeConf:   conf,
	}
}

func (c *cloudwatchCollector) Describe(desc chan<- *prometheus.Desc) {
	c.logger.Info("running describe in cloudwatch_exporter")
}

func (c *cloudwatchCollector) Collect(ch chan<- prometheus.Metric) {
	c.logger.Info("running collect in cloudwatch_exporter")

	reg := prometheus.NewRegistry()
	cwSemaphore := make(chan struct{}, cloudWatchConcurrency)
	tagSemaphore := make(chan struct{}, tagConcurrency)
	observedMetricLabels := map[string]yaceModel.LabelSet{}
	yace.UpdateMetrics(
		context.Background(),
		c.scrapeConf,
		reg,
		metricsPerQuery,
		labelsSnakeCase,
		cwSemaphore,
		tagSemaphore,
		c.sessionCache,
		observedMetricLabels,
		c.logger,
	)

	// Close concurrency channels
	close(cwSemaphore)
	close(tagSemaphore)

	reg.Collect(ch)
}

// yaceLoggerWrapper is wrapper implementation of yaceLog.Logger, based out of a log.Logger.
type yaceLoggerWrapper struct {
	log   log.Logger
	debug bool
}

func (l yaceLoggerWrapper) Info(message string, keyvals ...interface{}) {
	l.log.Log(append([]interface{}{"level", "info", "msg", message}, keyvals...)...)
}

func (l yaceLoggerWrapper) Debug(message string, keyvals ...interface{}) {
	l.log.Log(append([]interface{}{"level", "debug", "msg", message}, keyvals...)...)
}

func (l yaceLoggerWrapper) Error(err error, message string, keyvals ...interface{}) {
	l.log.Log(append([]interface{}{"level", "error", "msg", message, "err", err}, keyvals...)...)
}

func (l yaceLoggerWrapper) Warn(message string, keyvals ...interface{}) {
	l.log.Log(append([]interface{}{"level", "warn", "msg", message}, keyvals...)...)
}

func (l yaceLoggerWrapper) With(keyvals ...interface{}) yaceLog.Logger {
	withLog := log.With(l.log, keyvals)
	return yaceLoggerWrapper{
		log:   withLog,
		debug: l.debug,
	}
}

func (l yaceLoggerWrapper) IsDebugEnabled() bool {
	return l.debug
}
