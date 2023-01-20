package cloudwatch_exporter

import (
	"context"
	"github.com/go-kit/log"
	yace "github.com/nerdswords/yet-another-cloudwatch-exporter/pkg"
	yaceConf "github.com/nerdswords/yet-another-cloudwatch-exporter/pkg/config"
	yaceLog "github.com/nerdswords/yet-another-cloudwatch-exporter/pkg/logger"
	yaceModel "github.com/nerdswords/yet-another-cloudwatch-exporter/pkg/model"
	yaceSess "github.com/nerdswords/yet-another-cloudwatch-exporter/pkg/session"
	"github.com/prometheus/client_golang/prometheus"
)

// cloudwatchCollector wraps YACE entrypoint around a Collector implementation
type cloudwatchCollector struct {
	logger       yaceLoggerWrapper
	sessionCache yaceSess.SessionCache
	scrapeConf   yaceConf.ScrapeConf
}

// newCloudwatchCollector creates a new YACE wrapper
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
	c.logger.Debug("Running describe in cloudwatch_exporter")
}

func (c *cloudwatchCollector) Collect(ch chan<- prometheus.Metric) {
	c.logger.Debug("Running collect in cloudwatch_exporter")

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

	// close concurrency channels
	close(cwSemaphore)
	close(tagSemaphore)

	reg.Collect(ch)
}

// yaceLoggerWrapper is wrapper implementation of yaceLog.Logger, based out of a log.Logger.
type yaceLoggerWrapper struct {
	log log.Logger

	// debug is just used for development purposes
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
		log: withLog,
	}
}

func (l yaceLoggerWrapper) IsDebugEnabled() bool {
	return l.debug
}
