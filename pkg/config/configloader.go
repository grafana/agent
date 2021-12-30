package config

import (
	"context"
	"errors"
	"github.com/grafana/agent/pkg/integrations"
	"github.com/grafana/agent/pkg/metrics"
	"github.com/grafana/agent/pkg/metrics/instance"
	"github.com/hairyhenderson/go-fsimpl"
	"github.com/hairyhenderson/go-fsimpl/blobfs"
	"github.com/hairyhenderson/go-fsimpl/filefs"
	"github.com/hairyhenderson/gomplate/v3/data"
	"github.com/hairyhenderson/gomplate/v3/loader"
	"github.com/weaveworks/common/server"
	"gopkg.in/yaml.v2"
	"io"
	"io/fs"
	"net/url"
	"path/filepath"
	"strings"
)

// ConfigLoader is used to load configs from a variety of sources and squash them together. Supports templates via gomplate
type ConfigLoader struct {
	loader *loader.ConfigLoader
	mux    fsimpl.FSMux
	cfg    LoaderConfig
}

func NewConfigLoader(cfg LoaderConfig) (*ConfigLoader, error) {
	sources := make(map[string]*data.Source)
	for _, v := range cfg.Sources {
		url, err := url.Parse(v.URL)
		if err != nil {
			return nil, err
		}
		sources[v.Name] = &data.Source{
			URL:   url,
			Alias: v.Name,
		}
	}
	cl := loader.NewConfigLoader(context.Background(), sources)
	return &ConfigLoader{
		loader: cl,
		mux:    newFSProvider(),
		cfg:    cfg,
	}, nil
}

// ProcessConfigs loads the configurations in a predetermined order to handle functioning correctly. The only section
// not loaded is Server which is loaded from the passed in configuration. That is considered non-changing.
func (c *ConfigLoader) ProcessConfigs() (Config, error) {
	mainCfg := Config{}
	metricConfig, err := c.processMetric()
	if err != nil {
		return mainCfg, err
	}
	instancesConfigs, err := c.processMetricInstances()
	if err != nil {
		return mainCfg, err
	}
	metricConfig.Configs = instancesConfigs
	mainCfg.Metrics = metricConfig

	// The configuration for server fields MUST come from the confd style settings
	mainCfg.Server = c.cfg.Server
	return mainCfg, nil
}

// processMetric will return the first metric configuration found, following pattern `metrics-*.yml`
func (c *ConfigLoader) processMetric() (metrics.Config, error) {
	for _, path := range c.cfg.TemplatePaths {
		result, err := c.generateConfigsFromPath(path, "metrics-*.yml", func() interface{} {
			return &metrics.Config{}
		}, c.handleMatch)
		if err != nil {
			return metrics.Config{}, err
		}
		if len(result) > 1 {
			return metrics.Config{}, errors.New("multiple metrics configuration found")
		}
		if result != nil && len(result) == 1 {
			for _, cfg := range result {
				return *(cfg.(*metrics.Config)), nil
			}
		}
	}
	return metrics.Config{}, errors.New("no metrics configurations found")
}

// processMetricInstances will return the instance configurations used in the metrics section,
// following pattern `metrics_instances-*.yml`
func (c *ConfigLoader) processMetricInstances() ([]instance.Config, error) {
	builder := strings.Builder{}
	configs := make([]instance.Config, 0)
	for _, path := range c.cfg.TemplatePaths {
		result, err := c.generateConfigsFromPath(path, "metrics_instances-*.yml", func() interface{} {
			return &instance.Config{}
		}, c.handleMatch)
		if err != nil {
			builder.WriteString(err.Error())
		}
		for _, v := range result {
			pv := v.(*instance.Config)
			configs = append(configs, *pv)
		}

	}
	var combinedError error
	if builder.Len() > 0 {
		combinedError = errors.New(builder.String())
	}
	return configs, combinedError
}

// processExporters will return the exporter configurations, following pattern `exporters-.yml`
func (c *ConfigLoader) processExporters() ([]integrations.UnmarshaledConfig, error) {
	builder := strings.Builder{}
	configs := make([]integrations.UnmarshaledConfig, 0)
	for _, path := range c.cfg.TemplatePaths {
		result, err := c.generateConfigsFromPath(path, "exporters-*.yml", func() interface{} {
			return &integrations.UnmarshaledConfig{}
		}, c.handleExporterMatch)
		if err != nil {
			builder.WriteString(err.Error())
		}
		for _, v := range result {
			cfg := v.(*integrations.UnmarshaledConfig)
			configs = append(configs, *cfg)
		}

	}
	var combinedError error
	if builder.Len() > 0 {
		combinedError = errors.New(builder.String())
	}
	return configs, combinedError
}

// generateConfigsFromPath creates a series of yaml configs based on the path and a given string pattern.
// the pattern is the same as used by filepath.Match.
func (c *ConfigLoader) generateConfigsFromPath(path string, pattern string, configMake func() interface{}, matchHandler func(fs.FS, fs.DirEntry, func() interface{}) ([]interface{}, error)) ([]interface{}, error) {
	handler, err := c.mux.Lookup(path)
	if err != nil {
		return nil, err
	}
	files, err := fs.ReadDir(handler, ".")
	if err != nil {
		return nil, err
	}
	var configs []interface{}
	for _, f := range files {
		// We don't recurse into directories
		if f.IsDir() {
			continue
		}
		matched, _ := filepath.Match(pattern, f.Name())
		if matched {
			matchedConfigs, err := matchHandler(handler, f, configMake)
			if err != nil {
				return nil, err
			}
			for _, mc := range matchedConfigs {
				configs = append(configs, mc)
			}
		}
	}
	return configs, nil
}

func (c *ConfigLoader) handleMatch(handler fs.FS, f fs.DirEntry, configMake func() interface{}) ([]interface{}, error) {
	file, err := handler.Open(f.Name())
	stats, err := f.Info()
	if err != nil {
		return nil, err
	}
	fBytes := make([]byte, stats.Size())
	bytesRead, err := file.Read(fBytes)
	if bytesRead == 0 {
		return nil, errors.New("no bytes read")
	}

	if err != nil && err != io.EOF {
		return nil, err
	}
	fString := string(fBytes)
	// Parse the template
	processedConfigString, err := c.loader.GenerateTemplate("", fString)
	if err != nil {
		return nil, err
	}
	cfg := configMake()
	err = yaml.Unmarshal([]byte(processedConfigString), cfg)
	if err != nil && err != io.EOF {
		return nil, err
	}
	return []interface{}{cfg}, nil
}

// handleExporterMatch is more complex since those can be of any different types and not a single concrete type, like
// most of the configurations.
func (c *ConfigLoader) handleExporterMatch(handler fs.FS, f fs.DirEntry, _ func() interface{}) ([]interface{}, error) {
	file, err := handler.Open(f.Name())
	stats, err := f.Info()
	if err != nil {
		return nil, err
	}
	fBytes := make([]byte, stats.Size())
	bytesRead, err := file.Read(fBytes)
	if bytesRead == 0 {
		return nil, errors.New("no bytes read")
	}

	if err != nil && err != io.EOF {
		return nil, err
	}
	fString := string(fBytes)
	// Parse the template
	processedConfigString, err := c.loader.GenerateTemplate("", fString)
	if err != nil {
		return nil, err
	}
	cfg := integrations.TryUnmarshal(processedConfigString)
	if cfg == nil {
		return nil, err
	}
	// TODO there has to be a better way to handle this conversion
	var intConfigs []interface{}
	for _, i := range cfg {
		intConfigs = append(intConfigs, i)
	}
	return intConfigs, nil
}

func newFSProvider() fsimpl.FSMux {
	mux := fsimpl.NewMux()
	mux.Add(filefs.FS)
	mux.Add(blobfs.FS)
	return mux
}

type LoaderConfig struct {

	// Sources is used to define sources for variables using gomplate
	Sources []Datasource `yaml:"sources"`

	// TemplatePaths is the "directory" to look for templates in, they will be found and matched to configs but various
	// naming conventions. They can be S3/gcp, or file based resources.
	TemplatePaths []string `yaml:"template_paths"`

	// The server settings MUST be set
	Server server.Config `yaml:"server"`
}

type Datasource struct {
	Name string `yaml:"name"`
	URL  string `yaml:"url"`
}
