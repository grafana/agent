package config

import (
	"context"
	"fmt"
	"io/fs"
	"net/url"
	"os"
	"path/filepath"
	"strings"

	"github.com/grafana/agent/pkg/config/instrumentation"
	v2 "github.com/grafana/agent/pkg/integrations/v2"
	"github.com/grafana/agent/pkg/logs"
	"github.com/grafana/agent/pkg/metrics"
	"github.com/grafana/agent/pkg/metrics/instance"
	"github.com/grafana/agent/pkg/server"
	"github.com/grafana/agent/pkg/traces"
	"github.com/hairyhenderson/go-fsimpl"
	"github.com/hairyhenderson/go-fsimpl/blobfs"
	"github.com/hairyhenderson/go-fsimpl/filefs"
	"github.com/hairyhenderson/gomplate/v3/data"
	"github.com/hairyhenderson/gomplate/v3/loader"
	"github.com/hashicorp/go-multierror"
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
		// It takes some work arounds to parse all windows paths as url so treating it differently now is easier
		// otherwise we could parse path and then pivot
		stripPath := strings.ReplaceAll(path, "file://", "")
		buf, err = os.ReadFile(stripPath)
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

	instrumentation.InstrumentConfig(buf)

	cl := &LoaderConfig{}
	err = yaml.Unmarshal(buf, cl)
	if err != nil {
		return err
	}
	return c.LoadConfig(*cl)
}

// ProcessConfigs loads the configurations in a predetermined order to handle functioning correctly.
func (c *DynamicLoader) ProcessConfigs(cfg *Config) error {
	if c.cfg == nil {
		return fmt.Errorf("LoadConfig or LoadConfigByPath must be called")
	}
	var returnErr error

	err := c.processAgent(cfg)
	returnErr = errorAppend(returnErr, err)

	serverConfig, err := c.processServer()
	returnErr = errorAppend(returnErr, err)
	if serverConfig != nil {
		cfg.Server = serverConfig
	}

	metricConfig, err := c.processMetrics()
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

	cfg.Integrations.ExtraIntegrations = append(cfg.Integrations.ExtraIntegrations, integrations...)

	return returnErr
}

func (c *DynamicLoader) processAgent(cfg *Config) error {
	var returnError error
	found := 0
	for _, path := range c.cfg.TemplatePaths {
		filesContents, err := c.retrieveMatchingFileContents(path, c.cfg.AgentFilter, "agent")
		returnError = errorAppend(returnError, err)
		found = len(filesContents) + found
		if len(filesContents) == 1 {
			err = LoadBytes([]byte(filesContents[0]), false, cfg)
			returnError = errorAppend(returnError, err)
		}
	}
	if found > 1 {
		returnError = errorAppend(returnError, fmt.Errorf("found %d agent templates; expected 0 or 1", found))
	}
	// If we didnt find anything we still want to unmarshal the cfg to get defaults
	if found == 0 {
		_ = LoadBytes([]byte("{}"), false, cfg)
	}
	return returnError
}

func (c *DynamicLoader) processServer() (*server.Config, error) {
	var returnError error
	found := 0
	var cfg *server.Config
	for _, path := range c.cfg.TemplatePaths {
		filesContents, err := c.retrieveMatchingFileContents(path, c.cfg.ServerFilter, "server")
		returnError = errorAppend(returnError, err)
		found = len(filesContents) + found
		if len(filesContents) == 1 {
			cfg = &server.Config{}
			err = yaml.Unmarshal([]byte(filesContents[0]), cfg)
			returnError = errorAppend(returnError, err)
		}
	}
	if found > 1 {
		returnError = errorAppend(returnError, fmt.Errorf("found %d server templates; expected 0 or 1", found))
	}
	return cfg, returnError
}

func (c *DynamicLoader) processMetrics() (*metrics.Config, error) {
	var returnError error
	found := 0
	var cfg *metrics.Config
	for _, path := range c.cfg.TemplatePaths {
		filesContents, err := c.retrieveMatchingFileContents(path, c.cfg.MetricsFilter, "metrics")
		returnError = errorAppend(returnError, err)
		found = len(filesContents) + found
		if len(filesContents) == 1 {
			cfg = &metrics.Config{}
			err = yaml.Unmarshal([]byte(filesContents[0]), cfg)
			returnError = errorAppend(returnError, err)
		}
	}
	if found > 1 {
		returnError = errorAppend(returnError, fmt.Errorf("found %d metrics templates; expected 0 or 1", found))
	}
	return cfg, returnError
}

func (c *DynamicLoader) processMetricInstances() ([]instance.Config, error) {
	var returnError error
	var configs []instance.Config
	for _, path := range c.cfg.TemplatePaths {
		filesContents, err := c.retrieveMatchingFileContents(path, c.cfg.MetricsInstanceFilter, "metrics instances")
		returnError = errorAppend(returnError, err)
		for _, c := range filesContents {
			cfg := &instance.Config{}
			err = yaml.Unmarshal([]byte(c), cfg)
			returnError = errorAppend(returnError, err)
			configs = append(configs, *cfg)
		}
	}

	return configs, returnError
}

func (c *DynamicLoader) processIntegrations() ([]v2.Config, error) {
	var returnError error
	var configs []v2.Config
	for _, path := range c.cfg.TemplatePaths {
		filesContents, err := c.retrieveMatchingFileContents(path, c.cfg.IntegrationsFilter, "integrations")
		returnError = errorAppend(returnError, err)
		for _, c := range filesContents {
			intConfigs, err := unmarshalYamlToExporters(c)
			if err != nil {
				returnError = errorAppend(returnError, err)
				continue
			}
			configs = append(configs, intConfigs...)
		}
	}
	return configs, returnError
}

func (c *DynamicLoader) processLogs() (*logs.Config, error) {
	var returnError error
	found := 0
	var cfg *logs.Config
	for _, path := range c.cfg.TemplatePaths {
		filesContents, err := c.retrieveMatchingFileContents(path, c.cfg.LogsFilter, "logs")
		returnError = errorAppend(returnError, err)
		found = len(filesContents) + found
		if len(filesContents) == 1 {
			cfg = &logs.Config{}
			err = yaml.Unmarshal([]byte(filesContents[0]), cfg)
			returnError = errorAppend(returnError, err)
		}
	}
	if found > 1 {
		returnError = errorAppend(returnError, fmt.Errorf("found %d logs templates; expected 0 or 1", found))
	}
	return cfg, returnError
}

func (c *DynamicLoader) processTraces() (*traces.Config, error) {
	var returnError error
	found := 0
	var cfg *traces.Config
	for _, path := range c.cfg.TemplatePaths {
		filesContents, err := c.retrieveMatchingFileContents(path, c.cfg.TracesFilter, "traces")
		returnError = errorAppend(returnError, err)
		found = len(filesContents) + found
		if len(filesContents) == 1 {
			cfg = &traces.Config{}
			err = yaml.Unmarshal([]byte(filesContents[0]), cfg)
			returnError = errorAppend(returnError, err)
		}
	}
	if found > 1 {
		returnError = errorAppend(returnError, fmt.Errorf("found %d traces templates; expected 0 or 1", found))
	}
	return cfg, returnError
}

// retrieveMatchingFileContents retrieves the contents of files based on path and pattern
// the pattern is the same as used by filepath.Match.
func (c *DynamicLoader) retrieveMatchingFileContents(path, pattern, name string) ([]string, error) {
	var filesContents []string
	handler, err := c.mux.Lookup(path)
	if err != nil {
		return nil, err
	}
	files, err := fs.ReadDir(handler, ".")
	if err != nil {
		return nil, err
	}
	for _, f := range files {
		// We don't recurse into directories, mainly due to not wanting to deal with symlinks and other oddities
		// its likely we will revisit
		if f.IsDir() {
			continue
		}
		matched, err := filepath.Match(pattern, f.Name())
		if err != nil {
			return nil, err
		}
		if matched {
			contents, err := fs.ReadFile(handler, f.Name())
			if err != nil {
				return nil, err
			}
			processedConfigString, err := c.loader.GenerateTemplate(name, string(contents))
			if err != nil {
				return nil, err
			}
			filesContents = append(filesContents, processedConfigString)
		}
	}
	return filesContents, nil
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
