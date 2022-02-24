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

// NewDynamicLoader instantiates a new DynamicLoader.
func NewDynamicLoader() (*DynamicLoader, error) {

	return &DynamicLoader{
		mux: newFSProvider(),
	}, nil
}

// LoadConfig loads an already created LoaderConfig into the DynamicLoader.
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
	// Set Defaults
	if cfg.IntegrationsFilter == "" {
		cfg.IntegrationsFilter = "integrations-*.yml"
	}
	if cfg.AgentFilter == "" {
		cfg.AgentFilter = "agent-*.yml"
	}
	if cfg.ServerFilter == "" {
		cfg.ServerFilter = "server-*.yml"
	}
	if cfg.MetricsFilter == "" {
		cfg.MetricsFilter = "metrics-*.yml"
	}
	if cfg.MetricsInstanceFilter == "" {
		cfg.MetricsInstanceFilter = "metrics_instances-*.yml"
	}
	if cfg.LogsFilter == "" {
		cfg.LogsFilter = "logs-*.yml"
	}
	if cfg.TracesFilter == "" {
		cfg.TracesFilter = "traces-*.yml"
	}
	cl := loader.NewConfigLoader(context.Background(), sources)
	c.loader = cl
	c.cfg = &cfg
	return nil
}

// LoadConfigByPath creates a config based on a path.
func (c *DynamicLoader) LoadConfigByPath(path string) error {
	var buf []byte
	var err error
	switch {
	case strings.HasPrefix(path, "file://"):
		stripPath := strings.ReplaceAll(path, "file://", "")
		buf, err = ioutil.ReadFile(stripPath)
		if err != nil {
			return err
		}
	case strings.HasPrefix(path, "s3://"):
		blobURL, err := url.Parse(path)
		if err != nil {
			return err
		}
		buf, err = data.ReadBlob(*blobURL)
		if err != nil {
			return err
		}
	default:
		return fmt.Errorf("config path must start with file:// or s3://, not %s", path)
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

	err := yaml.UnmarshalStrict([]byte{}, cfg)
	returnErr = errorAppend(returnErr, err)

	err = c.processAgent(cfg)
	returnErr = errorAppend(returnErr, err)

	serverConfig, err := c.processServer()
	returnErr = errorAppend(returnErr, err)
	if serverConfig != nil {
		cfg.Server = *serverConfig
	}

	metricConfig, err := c.processMetric()
	returnErr = errorAppend(returnErr, err)
	if metricConfig != nil {
		cfg.Metrics = *metricConfig
	}

	instancesConfigs, err := c.processMetricInstances()
	returnErr = errorAppend(returnErr, err)
	cfg.Metrics.Configs = append(cfg.Metrics.Configs, instancesConfigs...)

	logsCfg, err := c.processLogs()
	returnErr = errorAppend(returnErr, err)
	if logsCfg != nil {
		cfg.Logs = logsCfg
	}

	traceConfigs, err := c.processTraces()
	returnErr = errorAppend(returnErr, err)
	if traceConfigs != nil {
		cfg.Traces = *traceConfigs
	}

	integrations, err := c.processIntegrations()
	returnErr = errorAppend(returnErr, err)
	// If integrations havent already been defined then we need to do
	// some setup
	if cfg.Integrations.configV2 == nil {
		cfg.Integrations = VersionedIntegrations{}
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
	returnErr = errorAppend(returnErr, err)
	return returnErr
}

func (c *DynamicLoader) processAgent(cfg *Config) error {
	for _, path := range c.cfg.TemplatePaths {
		result, err := c.generateConfigsFromPath(path, c.cfg.AgentFilter, func() interface{} {
			return cfg
		}, c.handleAgentMatch)
		if err != nil {
			return err
		}
		if len(result) > 1 {
			return fmt.Errorf("found %d agent templates; expected 0 or 1", len(result))
		}

	}
	return nil
}

func (c *DynamicLoader) processServer() (*server.Config, error) {
	for _, path := range c.cfg.TemplatePaths {
		result, err := c.generateConfigsFromPath(path, c.cfg.ServerFilter, func() interface{} {
			return &server.Config{}
		}, c.handleMatch)
		if err != nil {
			return nil, err
		}
		if len(result) > 1 {
			return nil, fmt.Errorf("found %d server templates; expected 0 or 1", len(result))
		}
		if len(result) == 1 {
			return result[0].(*server.Config), nil
		}
	}
	return nil, nil
}

func (c *DynamicLoader) processMetric() (*metrics.Config, error) {
	for _, path := range c.cfg.TemplatePaths {
		result, err := c.generateConfigsFromPath(path, c.cfg.MetricsFilter, func() interface{} {
			return &metrics.Config{}
		}, c.handleMatch)
		if err != nil {
			return nil, err
		}
		if len(result) > 1 {
			return nil, fmt.Errorf("found %d metrics templates; expected 0 or 1", len(result))
		}
		if len(result) == 1 {
			return result[0].(*metrics.Config), nil
		}
	}
	return nil, nil
}

func (c *DynamicLoader) processMetricInstances() ([]instance.Config, error) {
	var retError error
	configs := make([]instance.Config, 0)
	for _, path := range c.cfg.TemplatePaths {
		result, err := c.generateConfigsFromPath(path, c.cfg.MetricsInstanceFilter, func() interface{} {
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
	var returnError error
	configs := make([]v2.Config, 0)
	for _, path := range c.cfg.TemplatePaths {
		result, err := c.generateConfigsFromPath(path, c.cfg.IntegrationsFilter, func() interface{} {
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
	return configs, returnError
}

func (c *DynamicLoader) processLogs() (*logs.Config, error) {
	for _, path := range c.cfg.TemplatePaths {
		result, err := c.generateConfigsFromPath(path, c.cfg.LogsFilter, func() interface{} {
			return &logs.Config{}
		}, c.handleMatch)
		if err != nil {
			return nil, err
		}
		if len(result) > 1 {
			return nil, fmt.Errorf("found %d logs templates; expected 0 or 1", len(result))
		}
		if len(result) == 1 {
			return result[0].(*logs.Config), nil
		}

	}
	return nil, nil
}

func (c *DynamicLoader) processTraces() (*traces.Config, error) {
	for _, path := range c.cfg.TemplatePaths {
		result, err := c.generateConfigsFromPath(path, c.cfg.TracesFilter, func() interface{} {
			return &traces.Config{}
		}, c.handleMatch)
		if err != nil {
			return nil, err
		}
		if len(result) > 1 {
			return nil, fmt.Errorf("found %d traces templates; expected 0 or 1", len(result))
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
	fBytes, err := fs.ReadFile(handler, f.Name())
	if err != nil {
		return nil, err
	}
	fString := string(fBytes)
	// Parse the template
	processedConfigString := ""

	processedConfigString, err = c.loader.GenerateTemplate("", fString)
	if err != nil {
		return nil, err
	}

	cfg := configMake()

	// Expand Vars is false since gomplate already allows expanding vars
	err = LoadBytes([]byte(processedConfigString), false, cfg.(*Config))
	if err != nil {
		return nil, err
	}
	// setVersion actually does the unmarshalling for integrations
	err = cfg.(*Config).Integrations.setVersion(integrationsVersion2)
	return []interface{}{cfg}, err
}

func (c *DynamicLoader) handleMatch(handler fs.FS, f fs.DirEntry, configMake func() interface{}) ([]interface{}, error) {
	fBytes, err := fs.ReadFile(handler, f.Name())
	if err != nil {
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
	fBytes, err := fs.ReadFile(handler, f.Name())
	if err != nil {
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

// errorAppend is a wrapper around multierror.Append that is needed since multierror will create a new error. In this case
// we only want to create a new error if newErr is not nil
func errorAppend(root error, newErr error) error {
	if newErr == nil {
		return root
	}
	return multierror.Append(root, newErr)
}
