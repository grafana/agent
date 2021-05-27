package grok_exporter //nolint:golint

import (
	"fmt"
	v3 "github.com/fstab/grok_exporter/config/v3"
	"github.com/fstab/grok_exporter/exporter"
	"github.com/fstab/grok_exporter/oniguruma"
	"github.com/go-kit/kit/log"
	"github.com/grafana/agent/pkg/integrations"
	"github.com/grafana/agent/pkg/integrations/config"
	"gopkg.in/yaml.v3"
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
	p := (*plain)(c)
	err := unmarshal(p)
	if err != nil {
		return err
	}

	// Marshal and Unmarshal config again to add Defaults
	v3ConfigBytes, err := yaml.Marshal(p.GrokConfig)
	if err != nil {
		return fmt.Errorf("error marshalling grok config - %v", err)
	}

	v3Config, err := v3.Unmarshal(v3ConfigBytes)
	if err != nil {
		fmt.Errorf("error unmarshalling grok config - %v", err)
	}

	p.GrokConfig = *v3Config
	return nil
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

func (c *Config) CreateMetrics() ([]exporter.Metric, error) {
	patterns, err := c.initPatterns()
	if err != nil {
		return nil, err
	}

	result := make([]exporter.Metric, 0, len(c.GrokConfig.AllMetrics))
	for _, m := range c.GrokConfig.AllMetrics {
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
		case metricTypeCounter:
			result = append(result, exporter.NewCounterMetric(&m, regex, deleteRegex))
		case metricTypeGauge:
			result = append(result, exporter.NewGaugeMetric(&m, regex, deleteRegex))
		case metricTypeHistogram:
			result = append(result, exporter.NewHistogramMetric(&m, regex, deleteRegex))
		case metricTypeSummary:
			result = append(result, exporter.NewSummaryMetric(&m, regex, deleteRegex))
		default:
			return nil, fmt.Errorf("failed to initialize metrics: Metric type %v is not supported", m.Type)
		}
	}
	return result, nil
}

func (c *Config) initPatterns() (*exporter.Patterns, error) {
	patterns := exporter.InitPatterns()
	for _, importedPatterns := range c.GrokConfig.Imports {
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
	for _, pattern := range c.GrokConfig.GrokPatterns {
		err := patterns.AddPattern(pattern)
		if err != nil {
			return nil, err
		}
	}
	return patterns, nil
}
