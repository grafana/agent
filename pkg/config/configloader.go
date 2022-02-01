package config

import (
	"context"
	"errors"
	"flag"
	"fmt"
	v2 "github.com/grafana/agent/pkg/integrations/v2"
	"github.com/grafana/agent/pkg/logs"
	"github.com/grafana/agent/pkg/metrics"
	"github.com/grafana/agent/pkg/metrics/instance"
	"github.com/grafana/agent/pkg/traces"
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

// ConfigLoader is used to load configs from a variety of sources and squash them together.
// This is used by the dynamic configuration feature to load configurations from a set of templates and then run them through
// gomplate producing an end result.
type ConfigLoader struct {
	loader *loader.ConfigLoader
	mux    fsimpl.FSMux
	cfg    LoaderConfig
}

// NewConfigLoader instantiates a new ConfigLoader
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
func (c *ConfigLoader) ProcessConfigs(cfg *Config, fs *flag.FlagSet) error {

	err := c.processAgent(cfg)
	if err != nil {
		return err
	}
	serverConfig, err := c.processServer()
	if err != nil {
		return err
	}
	if serverConfig != nil {
		cfg.Server = *serverConfig
	}

	metricConfig, err := c.processMetric()
	if err != nil {
		return err
	}
	if metricConfig != nil {
		cfg.Metrics = *metricConfig
	}

	instancesConfigs, err := c.processMetricInstances()
	if err != nil {
		return err
	}
	for _, i := range instancesConfigs {
		cfg.Metrics.Configs = append(cfg.Metrics.Configs, i)
	}

	integrations, err := c.processIntegrations()
	if err != nil {
		return err
	}

	// If integrations havent already been defined then we need to do
	// some setup
	if cfg.Integrations.configV2 == nil {
		cfg.Integrations = DefaultVersionedIntegrations
		cfg.Integrations.setVersion(integrationsVersion2)
		defaultV2 := v2.DefaultSubsystemOptions
		cfg.Integrations.configV2 = &defaultV2
	}

	if len(integrations) > 0 {
		cfg.Integrations.configV2.Configs = integrations
	}

	logs, err := c.processLogs()
	if err != nil {
		return err
	}
	if logs != nil {
		cfg.Logs = logs
	}

	traceConfigs, err := c.processTraces()
	if err != nil {
		return err
	}
	if traceConfigs != nil {
		cfg.Traces = *traceConfigs
	}
	err = cfg.Validate(fs)
	if err != nil {
		return err
	}
	return nil
}

// processAgent will return the first agent configuration found, following the pattern `agent-*.yml`, sections
// of this config is overloaded by subsequent process* functions
func (c *ConfigLoader) processAgent(cfg *Config) error {
	for _, path := range c.cfg.TemplatePaths {
		result, err := c.generateConfigsFromPath(path, "agent-*.yml", func() interface{} {
			return cfg
		}, c.handleAgentMatch)
		if err != nil {
			return err
		}
		if len(result) > 1 {
			return errors.New("multiple agent configurations found")
		}

	}
	return nil
}

// processServer will return the first server configuration found, following the pattern `server-*.yml`
func (c *ConfigLoader) processServer() (*server.Config, error) {
	for _, path := range c.cfg.TemplatePaths {
		result, err := c.generateConfigsFromPath(path, "server-*.yml", func() interface{} {
			return &server.Config{}
		}, c.handleMatch)
		if err != nil {
			return nil, err
		}
		if len(result) > 1 {
			return nil, errors.New("multiple server configurations found")
		}
		if result != nil && len(result) == 1 {
			for _, cfg := range result {
				return cfg.(*server.Config), nil
			}
		}
	}
	return nil, nil
}

// processMetric will return the first metric configuration found, following the pattern `metrics-*.yml`
func (c *ConfigLoader) processMetric() (*metrics.Config, error) {
	for _, path := range c.cfg.TemplatePaths {
		result, err := c.generateConfigsFromPath(path, "metrics-*.yml", func() interface{} {
			return &metrics.Config{}
		}, c.handleMatch)
		if err != nil {
			return nil, err
		}
		if len(result) > 1 {
			return nil, errors.New("multiple metric configurations found")
		}
		if result != nil && len(result) == 1 {
			for _, cfg := range result {
				return cfg.(*metrics.Config), nil
			}
		}
	}
	return nil, nil
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

// processIntegrations will return the exporter configurations, following the pattern `integrations-*.yml`
func (c *ConfigLoader) processIntegrations() ([]v2.Config, error) {
	builder := strings.Builder{}
	configs := make([]v2.Config, 0)
	for _, path := range c.cfg.TemplatePaths {
		result, err := c.generateConfigsFromPath(path, "integrations-*.yml", func() interface{} {
			// This can return nil because we do a fancy lookup by the exporter name for the lookup
			return nil
		}, c.handleExporterMatch)
		if err != nil {
			builder.WriteString(err.Error())
		}
		for _, v := range result {
			cfg := v.(v2.Config)
			configs = append(configs, cfg)
		}

	}
	// Check to ensure no singleton exist
	singletonCheck := make(map[string]int, 0)
	for k, v := range v2.TypeRegistry() {
		if v == v2.TypeSingleton {
			singletonCheck[k] = 0
		}
	}
	for _, cfg := range configs {
		if _, ok := singletonCheck[cfg.Name()]; ok {
			singletonCheck[cfg.Name()]++
			if singletonCheck[cfg.Name()] > 1 {
				builder.WriteString(fmt.Sprintf("found multiple instances of singleton integration %s", cfg.Name()))
			}
		}
	}
	var combinedError error
	if builder.Len() > 0 {
		combinedError = errors.New(builder.String())
	}
	return configs, combinedError
}

// processLogs will return the logs configuration following the pattern `logs-*.yml`
func (c *ConfigLoader) processLogs() (*logs.Config, error) {
	for _, path := range c.cfg.TemplatePaths {
		result, err := c.generateConfigsFromPath(path, "logs-*.yml", func() interface{} {
			return &logs.Config{}
		}, c.handleMatch)
		if err != nil {
			return nil, err
		}
		if result != nil && len(result) > 1 {
			return nil, errors.New("multiple log templates found")
		}
		if result != nil && len(result) == 1 {
			return result[0].(*logs.Config), nil
		}

	}
	return nil, nil
}

// processTraces will return the traces configuration following the pattern `traces-*.yml`
func (c *ConfigLoader) processTraces() (*traces.Config, error) {
	for _, path := range c.cfg.TemplatePaths {
		result, err := c.generateConfigsFromPath(path, "traces-*.yml", func() interface{} {
			return &traces.Config{}
		}, c.handleMatch)
		if err != nil {
			return nil, err
		}
		if result != nil && len(result) > 1 {
			return nil, errors.New("multiple traces templates found")
		}
		if result != nil && len(result) == 1 {
			return result[0].(*traces.Config), nil
		}

	}
	return nil, nil
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

func (c *ConfigLoader) handleAgentMatch(handler fs.FS, f fs.DirEntry, configMake func() interface{}) ([]interface{}, error) {
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

	err = LoadBytes([]byte(processedConfigString), false, cfg.(*Config))
	if err != nil && err != io.EOF {
		return nil, err
	}
	// setVersion actually does the unmarshalling for integrations
	cfg.(*Config).Integrations.setVersion(integrationsVersion2)
	return []interface{}{cfg}, nil
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
	cfg, err := v2.UnmarshalYamlToExporters(processedConfigString)
	if cfg == nil || err != nil {
		return nil, err
	}
	// TODO (mattdurham) there has to be a better way to handle this conversion
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
	// naming conventions. They can be S3/gcp, or file based resources. The directory structure is NOT walked.
	TemplatePaths []string `yaml:"template_paths"`
}

// Datasource is used for gomplate and can be used for a variety of resources.
type Datasource struct {
	Name string `yaml:"name"`
	URL  string `yaml:"url"`
}
