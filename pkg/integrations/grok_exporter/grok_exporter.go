// Package grok_exporter embeds https://github.com/fstab/grok_exporter
package grok_exporter

import (
	"context"
	"fmt"
	v3 "github.com/fstab/grok_exporter/config/v3"
	"github.com/fstab/grok_exporter/exporter"
	"github.com/fstab/grok_exporter/oniguruma"
	"github.com/fstab/grok_exporter/tailer"
	"github.com/fstab/grok_exporter/tailer/fswatcher"
	"github.com/go-kit/kit/log"
	"github.com/go-kit/kit/log/level"
	"github.com/grafana/agent/pkg/integrations"
	"github.com/grafana/agent/pkg/integrations/config"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/sirupsen/logrus"
	"gopkg.in/yaml.v2"
	"net/http"
	"os"
	"time"
)

var (
	logfile = "logfile"
	extra   = "extra"
)

var additionalFieldDefinitions = map[string]string{
	logfile: "full path of the logger file",
	extra:   "full json logger object",
}

const (
	numberOfLinesMatchedLabel = "matched"
	numberOfLinesIgnoredLabel = "ignored"
	inputTypeWebhook          = "webhook"
)

// Config controls the grok_exporter integration.
type Config struct {
	Common config.Common `yaml:",inline"`

	GrokConfig v3.Config `yaml:",inline"`

	IncludeExporterMetrics bool `yaml:"include_exporter_metrics"`
}

// UnmarshalYAML implements yaml.Unmarshaler for Config.
func (c *Config) UnmarshalYAML(unmarshal func(interface{}) error) error {
	type plain Config
	return unmarshal((*plain)(c))
}

// Name returns the name of the integration that this config represents.
func (c *Config) Name() string {
	return "grok_exporter"
}

// CommonConfig returns the common settings shared across all integrations.
func (c *Config) CommonConfig() config.Common {
	return c.Common
}

// NewIntegration converts this config into an instance of an integration.
func (c *Config) NewIntegration(l log.Logger) (integrations.Integration, error) {
	return New(l, c)
}

func init() {
	integrations.RegisterIntegration(&Config{})
}

// Exporter defines the grok_exporter integration.
type Exporter struct {
	config          *Config
	registry        *prometheus.Registry
	metrics         []exporter.Metric
	selfMetrics     selfMetrics
	logTailer       fswatcher.FileTailer
	retentionTicker *time.Ticker
	logger          log.Logger
}

type selfMetrics struct {
	nLinesTotal                  *prometheus.CounterVec
	nMatchesByMetric             *prometheus.CounterVec
	procTimeMicrosecondsByMetric *prometheus.CounterVec
	nErrorsByMetric              *prometheus.CounterVec
}

// New creates a new grok_exporter integration
func New(logger log.Logger, config *Config) (integrations.Integration, error) {
	configBytes, err := yaml.Marshal(config.GrokConfig)
	if err != nil {
		return nil, fmt.Errorf("error marshalling config - %v", err)
	}

	v3Config, err := v3.Unmarshal(configBytes)
	if err != nil {
		return nil, fmt.Errorf("error unmarshalling config - %v", err)
	}
	config.GrokConfig = *v3Config

	registry := prometheus.NewRegistry()

	patterns, err := initPatterns(&config.GrokConfig)
	if err != nil {
		return nil, err
	}

	metrics, err := createMetrics(&config.GrokConfig, patterns)
	if err != nil {
		return nil, err
	}

	for _, m := range metrics {
		registry.MustRegister(m.Collector())
	}

	nLinesTotal, nMatchesByMetric, procTimeMicrosecondsByMetric, nErrorsByMetric := initSelfMonitoring(metrics, registry)

	if config.IncludeExporterMetrics {
		registry.MustRegister(prometheus.NewProcessCollector(prometheus.ProcessCollectorOpts{}))
		registry.MustRegister(prometheus.NewGoCollector())
	}

	logTailer, err := startTailer(&config.GrokConfig, registry)
	if err != nil {
		return nil, err
	}

	return &Exporter{
		config:   config,
		registry: registry,
		metrics:  metrics,
		selfMetrics: selfMetrics{
			nLinesTotal:                  nLinesTotal,
			nMatchesByMetric:             nMatchesByMetric,
			procTimeMicrosecondsByMetric: procTimeMicrosecondsByMetric,
			nErrorsByMetric:              nErrorsByMetric,
		},
		logTailer:       logTailer,
		retentionTicker: time.NewTicker(config.GrokConfig.Global.RetentionCheckInterval),
		logger:          logger,
	}, nil
}

// MetricsHandler returns the HTTP handler for the integration.
func (e *Exporter) MetricsHandler() (http.Handler, error) {
	metricsHandler := promhttp.HandlerFor(e.registry, promhttp.HandlerOpts{
		ErrorHandling: promhttp.ContinueOnError,
	})
	if e.config.IncludeExporterMetrics {
		metricsHandler = promhttp.InstrumentMetricHandler(e.registry, metricsHandler)
	}
	return metricsHandler, nil
}

// ScrapeConfigs satisfies Integration.ScrapeConfigs.
func (e *Exporter) ScrapeConfigs() []config.ScrapeConfig {
	return []config.ScrapeConfig{{
		JobName:     e.config.Name(),
		MetricsPath: "/metrics",
	}}
}

// Run satisfies Run.
func (e *Exporter) Run(ctx context.Context) error {
	for {
		select {
		case err := <-e.logTailer.Errors():
			if err.Type() == fswatcher.FileNotFound || os.IsNotExist(err.Cause()) {
				return fmt.Errorf("error reading log lines: %v: use 'fail_on_missing_logfile: false' in the input configuration if you want grok_exporter to start even though the logfile is missing", err)
			} else {
				return fmt.Errorf("error reading log lines: %v", err.Error())
			}
		case line := <-e.logTailer.Lines():
			matched := false
			for _, metric := range e.metrics {
				start := time.Now()
				if !metric.PathMatches(line.File) {
					continue
				}
				match, err := metric.ProcessMatch(line.Line, makeAdditionalFields(line))
				if err != nil {
					level.Warn(e.logger).Log("WARNING: skipping log line - ", line.Line, "err - ", err.Error())
					e.selfMetrics.nErrorsByMetric.WithLabelValues(metric.Name()).Inc()
				} else if match != nil {
					e.selfMetrics.nMatchesByMetric.WithLabelValues(metric.Name()).Inc()
					e.selfMetrics.procTimeMicrosecondsByMetric.WithLabelValues(metric.Name()).Add(float64(time.Since(start).Nanoseconds() / int64(1000)))
					matched = true
				}
				_, err = metric.ProcessDeleteMatch(line.Line, makeAdditionalFields(line))
				if err != nil {
					level.Warn(e.logger).Log("WARNING: skipping log line - ", line.Line, "err - ", err.Error())
					e.selfMetrics.nErrorsByMetric.WithLabelValues(metric.Name()).Inc()
				}
			}
			if matched {
				e.selfMetrics.nLinesTotal.WithLabelValues(numberOfLinesMatchedLabel).Inc()
			} else {
				e.selfMetrics.nLinesTotal.WithLabelValues(numberOfLinesIgnoredLabel).Inc()
			}
		case <-e.retentionTicker.C:
			for _, metric := range e.metrics {
				err := metric.ProcessRetention()
				if err != nil {
					level.Warn(e.logger).Log("WARNING: error while processing retention on metric - ", metric.Name(), "err - ", err.Error())
					e.selfMetrics.nErrorsByMetric.WithLabelValues(metric.Name()).Inc()
				}
			}
		}
	}
	return nil
}

// CustomHandlers returns extra handlers for the integration.
func (e *Exporter) CustomHandlers() map[string]http.Handler {
	handlers := make(map[string]http.Handler)
	if e.config.GrokConfig.Input.Type == inputTypeWebhook {
		handlers[fmt.Sprintf("/integrations/grok_exporter%s", e.config.GrokConfig.Input.WebhookPath)] = tailer.WebhookHandler()
	}

	return handlers
}

func initPatterns(cfg *v3.Config) (*exporter.Patterns, error) {
	patterns := exporter.InitPatterns()
	for _, importedPatterns := range cfg.Imports {
		if importedPatterns.Type == "grok_patterns" {
			if len(importedPatterns.Dir) > 0 {
				err := patterns.AddDir(importedPatterns.Dir)
				if err != nil {
					return nil, fmt.Errorf("failed to initialize patterns: %w", err)
				}
			} else if len(importedPatterns.File) > 0 {
				err := patterns.AddGlob(importedPatterns.File)
				if err != nil {
					return nil, fmt.Errorf("failed to initialize patterns: %w", err)
				}
			}
		}
	}
	for _, pattern := range cfg.GrokPatterns {
		err := patterns.AddPattern(pattern)
		if err != nil {
			return nil, err
		}
	}
	return patterns, nil
}

func createMetrics(cfg *v3.Config, patterns *exporter.Patterns) ([]exporter.Metric, error) {
	result := make([]exporter.Metric, 0, len(cfg.AllMetrics))
	for _, m := range cfg.AllMetrics {
		var (
			regex, deleteRegex *oniguruma.Regex
			err                error
		)
		regex, err = exporter.Compile(m.Match, patterns)
		if err != nil {
			return nil, fmt.Errorf("failed to initialize metric %v: %v", m.Name, err.Error())
		}
		if len(m.DeleteMatch) > 0 {
			deleteRegex, err = exporter.Compile(m.DeleteMatch, patterns)
			if err != nil {
				return nil, fmt.Errorf("failed to initialize metric %v: %v", m.Name, err.Error())
			}
		}
		err = exporter.VerifyFieldNames(&m, regex, deleteRegex, additionalFieldDefinitions)
		if err != nil {
			return nil, fmt.Errorf("failed to initialize metric %v: %v", m.Name, err.Error())
		}
		switch m.Type {
		case "counter":
			result = append(result, exporter.NewCounterMetric(&m, regex, deleteRegex))
		case "gauge":
			result = append(result, exporter.NewGaugeMetric(&m, regex, deleteRegex))
		case "histogram":
			result = append(result, exporter.NewHistogramMetric(&m, regex, deleteRegex))
		case "summary":
			result = append(result, exporter.NewSummaryMetric(&m, regex, deleteRegex))
		default:
			return nil, fmt.Errorf("failed to initialize metrics: Metric type %v is not supported", m.Type)
		}
	}
	return result, nil
}

func initSelfMonitoring(metrics []exporter.Metric, registry prometheus.Registerer) (*prometheus.CounterVec, *prometheus.CounterVec, *prometheus.CounterVec, *prometheus.CounterVec) {
	buildInfo := prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Name: "grok_exporter_build_info",
		Help: "A metric with a constant '1' value labeled by version, builddate, branch, revision, goversion, and platform on which grok_exporter was built.",
	}, []string{"version", "builddate", "branch", "revision", "goversion", "platform"})
	nLinesTotal := prometheus.NewCounterVec(prometheus.CounterOpts{
		Name: "grok_exporter_lines_total",
		Help: "Total number of logger lines processed by grok_exporter.",
	}, []string{"status"})
	nMatchesByMetric := prometheus.NewCounterVec(prometheus.CounterOpts{
		Name: "grok_exporter_lines_matching_total",
		Help: "Number of lines matched for each metric. Note that one line can be matched by multiple metrics.",
	}, []string{"metric"})
	procTimeMicrosecondsByMetric := prometheus.NewCounterVec(prometheus.CounterOpts{
		Name: "grok_exporter_lines_processing_time_microseconds_total",
		Help: "Processing time in microseconds for each metric. Divide by grok_exporter_lines_matching_total to get the averge processing time for one logger line.",
	}, []string{"metric"})
	nErrorsByMetric := prometheus.NewCounterVec(prometheus.CounterOpts{
		Name: "grok_exporter_line_processing_errors_total",
		Help: "Number of errors for each metric. If this is > 0 there is an error in the configuration file. Check grok_exporter's console output.",
	}, []string{"metric"})

	registry.MustRegister(buildInfo)
	registry.MustRegister(nLinesTotal)
	registry.MustRegister(nMatchesByMetric)
	registry.MustRegister(procTimeMicrosecondsByMetric)
	registry.MustRegister(nErrorsByMetric)

	buildInfo.WithLabelValues(exporter.Version, exporter.BuildDate, exporter.Branch, exporter.Revision, exporter.GoVersion, exporter.Platform).Set(1)
	// Initializing a value with zero makes the label appear. Otherwise the label is not shown until the first value is observed.
	nLinesTotal.WithLabelValues(numberOfLinesMatchedLabel).Add(0)
	nLinesTotal.WithLabelValues(numberOfLinesIgnoredLabel).Add(0)
	for _, metric := range metrics {
		nMatchesByMetric.WithLabelValues(metric.Name()).Add(0)
		procTimeMicrosecondsByMetric.WithLabelValues(metric.Name()).Add(0)
		nErrorsByMetric.WithLabelValues(metric.Name()).Add(0)
	}
	return nLinesTotal, nMatchesByMetric, procTimeMicrosecondsByMetric, nErrorsByMetric
}

func startTailer(cfg *v3.Config, registry prometheus.Registerer) (fswatcher.FileTailer, error) {
	var (
		tail fswatcher.FileTailer
		err  error
	)
	logger := logrus.New()
	logger.SetLevel(logrus.WarnLevel)
	switch {
	case cfg.Input.Type == "file":
		if cfg.Input.PollInterval == 0 {
			tail, err = fswatcher.RunFileTailer(cfg.Input.Globs, cfg.Input.Readall, cfg.Input.FailOnMissingLogfile, logger)
			if err != nil {
				return nil, fmt.Errorf("failed to run file tailer: %v", err)
			}
		} else {
			tail, err = fswatcher.RunPollingFileTailer(cfg.Input.Globs, cfg.Input.Readall, cfg.Input.FailOnMissingLogfile, cfg.Input.PollInterval, logger)
			if err != nil {
				return nil, fmt.Errorf("failed to run polling file tailer: %v", err)
			}
		}
	case cfg.Input.Type == "stdin":
		tail = tailer.RunStdinTailer()
	case cfg.Input.Type == "webhook":
		tail = tailer.InitWebhookTailer(&cfg.Input)
	case cfg.Input.Type == "kafka":
		tail = tailer.RunKafkaTailer(&cfg.Input)
	default:
		return nil, fmt.Errorf("config error: Input type '%v' unknown", cfg.Input.Type)
	}
	bufferLoadMetric := exporter.NewBufferLoadMetric(logger, cfg.Input.MaxLinesInBuffer > 0, registry)
	return tailer.BufferedTailerWithMetrics(tail, bufferLoadMetric, logger, cfg.Input.MaxLinesInBuffer), nil
}

func makeAdditionalFields(line *fswatcher.Line) map[string]interface{} {
	return map[string]interface{}{
		logfile: line.File,
		extra:   line.Extra,
	}
}
