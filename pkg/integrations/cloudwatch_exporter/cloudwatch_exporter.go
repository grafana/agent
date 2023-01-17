package cloudwatch_exporter

import (
	"context"
	"fmt"
	"github.com/grafana/agent/pkg/integrations"
	integrations_v2 "github.com/grafana/agent/pkg/integrations/v2"
	"github.com/grafana/agent/pkg/integrations/v2/metricsutils"
	"time"

	"github.com/go-kit/log"
	yace "github.com/nerdswords/yet-another-cloudwatch-exporter/pkg"
	yaceConf "github.com/nerdswords/yet-another-cloudwatch-exporter/pkg/config"
	yaceLog "github.com/nerdswords/yet-another-cloudwatch-exporter/pkg/logger"
	yaceModel "github.com/nerdswords/yet-another-cloudwatch-exporter/pkg/model"
	yaceSess "github.com/nerdswords/yet-another-cloudwatch-exporter/pkg/session"
	"github.com/prometheus/client_golang/prometheus"
)

func init() {
	integrations.RegisterIntegration(&Config{})
	integrations_v2.RegisterLegacy(&Config{}, integrations_v2.TypeMultiplex, metricsutils.NewNamedShim("cloudwatch_exporter"))
}

type Config struct {
	STSRegion string          `yaml:"stsRegion"`
	Discovery DiscoveryConfig `yaml:"discovery"`
	Static    []StaticJob     `yaml:"static"`
}

// Discovery Jobs

type DiscoveryConfig struct {
	ExportedTags TagsPerNamespace `yaml:"exportedTags"`
	Jobs         []*DiscoveryJob  `yaml:"jobs"`
}

// TagsPerNamespace represents for each namespace, a list of tags that will be exported as labels in each metric.
type TagsPerNamespace map[string][]string

type DiscoveryJob struct {
	InlineRegionAndRoles `yaml:",inline"`
	InlineCustomTags     `yaml:",inline"`
	Type                 string   `yaml:"type"`
	Metrics              []Metric `yaml:"metrics"`
}

// Static Jobs

type StaticJob struct {
	InlineRegionAndRoles `yaml:",inline"`
	InlineCustomTags     `yaml:",inline"`
	Name                 string      `yaml:"name"`
	Namespace            string      `yaml:"namespace"`
	Dimensions           []Dimension `yaml:"dimensions"`
	Metrics              []Metric    `yaml:"metrics"`
}

// Supporting types

// InlineRegionAndRoles exposes for each supported job, the AWS regions and IAM roles in which the agent should perform the
// scrape.
type InlineRegionAndRoles struct {
	Regions []string `yaml:"regions"`
	Roles   []Role   `yaml:"roles"`
}

type InlineCustomTags struct {
	CustomTags []Tag `yaml:"customTags"`
}

type Role struct {
	RoleArn    string `yaml:"roleArn"`
	ExternalID string `yaml:"externalID"`
}

type Dimension struct {
	Name  string `yaml:"name"`
	Value string `yaml:"value"`
}

type Tag struct {
	Key   string `yaml:"key"`
	Value string `yaml:"value"`
}

type Metric struct {
	Name       string        `yaml:"name"`
	Statistics []string      `yaml:"statistics"`
	Period     time.Duration `yaml:"period"`
}

func (c *Config) Name() string {
	return "cloudwatch_exporter"
}

// todo: is this used when there's more than one integration instnace in an agent config
func (c *Config) InstanceKey(agentKey string) (string, error) {
	return c.Name(), nil
}

func (c *Config) NewIntegration(l log.Logger) (integrations.Integration, error) {
	exporterConfig, err := ToYACEConfig(c)
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
