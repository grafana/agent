package config

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"io"
	"io/fs"
	"io/ioutil"
	"net/url"
	"path/filepath"
	"strings"

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
	"github.com/hashicorp/go-multierror"
	"github.com/weaveworks/common/server"
	"gopkg.in/yaml.v2"
)

// DynamicLoader is used to load configs from a variety of sources and squash them together.
// This is used by the dynamic configuration feature to load configurations from a set of templates and then run them through
// gomplate producing an end result.
type DynamicLoader struct {
	loader *loader.ConfigLoader
	mux    fsimpl.FSMux
	cfg    *LoaderConfig
}

// NewDynamicLoader instantiates a new DynamicLoader
func NewDynamicLoader() (*DynamicLoader, error) {

	return &DynamicLoader{
		loader: nil,
		mux:    newFSProvider(),
		cfg:    nil,
	}, nil
}

// LoadConfig loads an already created LoaderConfig into the DynamicLoader
func (c *DynamicLoader) LoadConfig(cfg LoaderConfig) error {
	sources := make(map[string]*data.Source)
	for _, v := range cfg.Sources {
		sourceURL, err := url.Parse(v.URL)
		if err != nil {
			return err
		}
		sources[v.Name] = &data.Source{
			URL:   sourceURL,
			Alias: v.Name,
		}
	}
	cl := loader.NewConfigLoader(context.Background(), sources)
	c.loader = cl
	c.cfg = &cfg
	return nil
}

// LoadConfigByPath creates a config based on a path
func (c *DynamicLoader) LoadConfigByPath(path string) error {
	var buf []byte
	var err error
	if strings.HasPrefix(path, "file://") {
		stripPath := strings.ReplaceAll(path, "file://", "")
		buf, err = ioutil.ReadFile(stripPath)
		if err != nil {
			return err
		}
	} else if strings.HasPrefix(path, "s3://") {
		blobURL, err := url.Parse(path)
		if err != nil {
			return err
		}
		buf, err = data.ReadBlob(*blobURL)
		if err != nil {
			return err
		}
	}
	cl := &LoaderConfig{}
	err = yaml.Unmarshal(buf, cl)
	if err != nil {
		return err
	}
	return c.LoadConfig(*cl)
}

// ProcessConfigs loads the configurations in a predetermined order to handle functioning correctly.
func (c *DynamicLoader) ProcessConfigs(cfg *Config, fs *flag.FlagSet) error {
	if c.cfg == nil {
		return errors.New("LoadConfig or LoadConfigByPath must be called")
	}
	var returnErr error
	err := c.processAgent(cfg)
	if err != nil {
		returnErr = multierror.Append(returnErr, err)
	}

	serverConfig, err := c.processServer()
	if err != nil {
		returnErr = multierror.Append(returnErr, err)
	}
	if serverConfig != nil {
		cfg.Server = *serverConfig
	}

	metricConfig, err := c.processMetric()
	if err != nil {
		returnErr = multierror.Append(returnErr, err)
	}
	if metricConfig != nil {
		cfg.Metrics = *metricConfig
	}

	instancesConfigs, err := c.processMetricInstances()
	if err != nil {
		returnErr = multierror.Append(returnErr, err)
	}
	cfg.Metrics.Configs = append(cfg.Metrics.Configs, instancesConfigs...)

	logsCfg, err := c.processLogs()
	if err != nil {
		returnErr = multierror.Append(returnErr, err)
	}
	if logsCfg != nil {
		cfg.Logs = logsCfg
	}

	traceConfigs, err := c.processTraces()
	if err != nil {
		returnErr = multierror.Append(returnErr, err)
	}
	if traceConfigs != nil {
		cfg.Traces = *traceConfigs
	}

	integrations, err := c.processIntegrations()
	if err != nil {
		returnErr = multierror.Append(returnErr, err)
	}
	// If integrations havent already been defined then we need to do
	// some setup
	if cfg.Integrations.configV2 == nil {
		cfg.Integrations = DefaultVersionedIntegrations
		err = cfg.Integrations.setVersion(integrationsVersion2)
		if err != nil {
			returnErr = multierror.Append(returnErr, err)
		}
		defaultV2 := v2.DefaultSubsystemOptions
		cfg.Integrations.configV2 = &defaultV2
	}
	if len(integrations) > 0 {
		cfg.Integrations.configV2.Configs = append(cfg.Integrations.configV2.Configs, integrations...)
	}

	err = cfg.Validate(fs)
	if err != nil {
		returnErr = multierror.Append(returnErr, err)
	}
	return returnErr
}

func (c *DynamicLoader) processAgent(cfg *Config) error {
	filter := "agent-*.yml"
	if c.cfg.AgentFilter != "" {
		filter = c.cfg.AgentFilter
	}

	for _, path := range c.cfg.TemplatePaths {
		result, err := c.generateConfigsFromPath(path, filter, func() interface{} {
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

func (c *DynamicLoader) processServer() (*server.Config, error) {
	filter := "server-*.yml"
	if c.cfg.ServerFilter != "" {
		filter = c.cfg.ServerFilter
	}

	for _, path := range c.cfg.TemplatePaths {
		result, err := c.generateConfigsFromPath(path, filter, func() interface{} {
			return &server.Config{}
		}, c.handleMatch)
		if err != nil {
			return nil, err
		}
		if len(result) > 1 {
			return nil, errors.New("multiple server configurations found")
		}
		if len(result) == 1 {
			return result[0].(*server.Config), nil
		}
	}
	return nil, nil
}

func (c *DynamicLoader) processMetric() (*metrics.Config, error) {
	filter := "metrics-*.yml"
	if c.cfg.MetricsFilter != "" {
		filter = c.cfg.MetricsFilter
	}

	for _, path := range c.cfg.TemplatePaths {
		result, err := c.generateConfigsFromPath(path, filter, func() interface{} {
			return &metrics.Config{}
		}, c.handleMatch)
		if err != nil {
			return nil, err
		}
		if len(result) > 1 {
			return nil, errors.New("multiple metrics configurations found")
		}
		if len(result) == 1 {
			return result[0].(*metrics.Config), nil
		}
	}
	return nil, nil
}

func (c *DynamicLoader) processMetricInstances() ([]instance.Config, error) {
	filter := "metrics_instances-*.yml"
	if c.cfg.MetricsInstanceFilter != "" {
		filter = c.cfg.MetricsInstanceFilter
	}

	var retError error
	configs := make([]instance.Config, 0)
	for _, path := range c.cfg.TemplatePaths {
		result, err := c.generateConfigsFromPath(path, filter, func() interface{} {
			return &instance.Config{}
		}, c.handleMatch)
		if err != nil {
			retError = multierror.Append(retError, err)
		}
		for _, v := range result {
			pv := v.(*instance.Config)
			configs = append(configs, *pv)
		}

	}
	return configs, retError
}

func (c *DynamicLoader) processIntegrations() ([]v2.Config, error) {
	filter := "integrations-*.yml"
	if c.cfg.IntegrationsFilter != "" {
		filter = c.cfg.IntegrationsFilter
	}

	var returnError error
	configs := make([]v2.Config, 0)
	for _, path := range c.cfg.TemplatePaths {
		result, err := c.generateConfigsFromPath(path, filter, func() interface{} {
			// This can return nil because we do a fancy lookup by the exporter name for the lookup
			return nil
		}, c.handleExporterMatch)
		if err != nil {
			returnError = multierror.Append(returnError, err)
		}
		for _, v := range result {
			cfg := v.(v2.Config)
			configs = append(configs, cfg)
		}
	}
	// Check to ensure no singleton exist
	singletonCheck := make(map[string]interface{})
	for _, cfg := range configs {
		if t, ok := v2.RegisteredType(cfg.Name()); ok && t == v2.TypeSingleton {
			if _, ok := singletonCheck[cfg.Name()]; ok {
				returnError = multierror.Append(returnError, fmt.Errorf("found multiple instances of singleton integration %s", cfg.Name()))
			} else {
				singletonCheck[cfg.Name()] = nil
			}
		}
	}
	return configs, returnError
}

func (c *DynamicLoader) processLogs() (*logs.Config, error) {
	filter := "logs-*.yml"
	if c.cfg.LogsFilter != "" {
		filter = c.cfg.LogsFilter
	}

	for _, path := range c.cfg.TemplatePaths {
		result, err := c.generateConfigsFromPath(path, filter, func() interface{} {
			return &logs.Config{}
		}, c.handleMatch)
		if err != nil {
			return nil, err
		}
		if len(result) > 1 {
			return nil, errors.New("multiple logs templates found")
		}
		if len(result) == 1 {
			return result[0].(*logs.Config), nil
		}

	}
	return nil, nil
}

// processTraces will return the traces configuration following the pattern `traces-*.yml`
func (c *DynamicLoader) processTraces() (*traces.Config, error) {
	filter := "traces-*.yml"
	if c.cfg.TracesFilter != "" {
		filter = c.cfg.TracesFilter
	}

	for _, path := range c.cfg.TemplatePaths {
		result, err := c.generateConfigsFromPath(path, filter, func() interface{} {
			return &traces.Config{}
		}, c.handleMatch)
		if err != nil {
			return nil, err
		}
		if len(result) > 1 {
			return nil, errors.New("multiple traces templates found")
		}
		if len(result) == 1 {
			return result[0].(*traces.Config), nil
		}

	}
	return nil, nil
}

// generateConfigsFromPath creates a series of yaml configs based on the path and a given string pattern.
// the pattern is the same as used by filepath.Match.
func (c *DynamicLoader) generateConfigsFromPath(path string, pattern string, configMake func() interface{}, matchHandler func(fs.FS, fs.DirEntry, func() interface{}) ([]interface{}, error)) ([]interface{}, error) {
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
		matched, err := filepath.Match(pattern, f.Name())
		if err != nil {
			return nil, err
		}
		if matched {
			matchedConfigs, err := matchHandler(handler, f, configMake)
			if err != nil {
				return nil, err
			}
			configs = append(configs, matchedConfigs...)
		}
	}
	return configs, nil
}

func (c *DynamicLoader) handleAgentMatch(handler fs.FS, f fs.DirEntry, configMake func() interface{}) ([]interface{}, error) {
	file, err := handler.Open(f.Name())
	if err != nil {
		return nil, err
	}
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

	// Expand Vars is false since gomplate already allows expanding vars
	err = LoadBytes([]byte(processedConfigString), false, cfg.(*Config))
	if err != nil && err != io.EOF {
		return nil, err
	}
	// setVersion actually does the unmarshalling for integrations
	err = cfg.(*Config).Integrations.setVersion(integrationsVersion2)
	return []interface{}{cfg}, err
}

func (c *DynamicLoader) handleMatch(handler fs.FS, f fs.DirEntry, configMake func() interface{}) ([]interface{}, error) {
	file, err := handler.Open(f.Name())
	if err != nil {
		return nil, err
	}
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
func (c *DynamicLoader) handleExporterMatch(handler fs.FS, f fs.DirEntry, _ func() interface{}) ([]interface{}, error) {
	file, err := handler.Open(f.Name())
	if err != nil {
		return nil, err
	}
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
	cfg, err := unmarshalYamlToExporters(processedConfigString)
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

// unmarshalYamlToExporters attempts to convert the contents of yaml string into a set of exporters and then return
// those configurations.
func unmarshalYamlToExporters(contents string) ([]v2.Config, error) {
	o := &v2.SubsystemOptions{}
	err := yaml.Unmarshal([]byte(contents), o)
	if err != nil {
		return nil, err
	}
	return o.Configs, nil
}

func newFSProvider() fsimpl.FSMux {
	mux := fsimpl.NewMux()
	mux.Add(filefs.FS)
	mux.Add(blobfs.FS)
	return mux
}
