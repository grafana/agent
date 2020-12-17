// Package jmx_exporter runs https://github.com/prometheus/jmx_exporter as a child process.
package jmx_exporter //nolint:golint

import (
	"context"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httputil"
	"net/url"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/go-kit/kit/log"
	"github.com/go-kit/kit/log/level"
	"github.com/gorilla/mux"
	"github.com/grafana/agent/pkg/integrations/config"
	"github.com/grafana/agent/pkg/integrations/downloader"
	"github.com/grafana/agent/pkg/util"
	"gopkg.in/yaml.v2"
)

// URL and SHA256 sum of the JMX exporter JAR.
var (
	JMXExporterVersion  = "0.14.0"
	JMXExporterURL      = fmt.Sprintf("https://repo1.maven.org/maven2/io/prometheus/jmx/jmx_prometheus_httpserver/%[1]s/jmx_prometheus_httpserver-%[1]s-jar-with-dependencies.jar", JMXExporterVersion)
	JMXExporterFilename = fmt.Sprintf("jmx_prometheus_httpserver-%s-jar-with-dependencies.jar", JMXExporterVersion)
	JMXExporterSHA256   = "c52a8f0556aa5d6a2a2180f3766abb45b7d09fff5f2ad01f8b7dba67e4ce8291"

	// JMXCacheSubpath is the subpath within the configured CacheDirectory to store the downloaded JAR.
	JMXCacheSubpath = "grafana-agent/jmx_exporter"
)

// DefaultConfig holds default values for the config.
var DefaultConfig = Config{
	ListenAddress:  "localhost:5556",
	ExporterConfig: yaml.MapSlice{},
}

type Config struct {
	CommonConfig config.Common `yaml:",inline"`

	// Enabled enables the jmx_exporter integration.
	Enabled bool `yaml:"enabled"`

	// CacheDirectory holds the directory to hold the downloaded JAR.
	// If empty, defaults to one of the following based on host OS:
	//
	// - Linux: $XDG_CACHE_HOME or $HOME/.cache
	// - macOS: $HOME/Library/Caches
	// - Windows: %LocalAppData%
	//
	// The JAR will be stored inside the subpath grafana-agent/jmx_exporter/jmx_prometheus_javaagent-<version>.jar.
	CacheDirectory string `yaml:"jar_cache_directory"`

	// Path to the Java binary. If empty, looks in the PATH environment variable
	// for Java. Java must be installed for the integration to be used.
	JavaPath string `yaml:"java_path"`

	// JVMOptions like "-Dcom.sun.management.jmxremote.ssl=false"
	JVMOptions []string `yaml:"jvm_options"`

	// ListenAddress is the address to listen for http traffic on for exposing
	// metrics. Should be in host:port form.
	ListenAddress string `yaml:"listen_address"`

	// Config of the exporter. Passed to the JAR as the config file.
	ExporterConfig yaml.MapSlice `yaml:"exporter_config"`
}

// UnmarshalYAML implements yaml.Unmarshaler for Config.
func (c *Config) UnmarshalYAML(unmarshal func(interface{}) error) error {
	*c = DefaultConfig

	type plain Config
	return unmarshal((*plain)(c))
}

// Integration is the jmx_exporter integration.
type Integration struct {
	c Config
	l log.Logger
}

func New(log log.Logger, c Config) (*Integration, error) {
	if c.CacheDirectory == "" {
		var err error
		c.CacheDirectory, err = os.UserCacheDir()
		if err != nil {
			return nil, fmt.Errorf("failed to determine cache directory: %w", err)
		}
	}
	// Try to create the final cache directory to bail early if something is wrong.
	if err := os.MkdirAll(filepath.Join(c.CacheDirectory, JMXCacheSubpath), 0775); err != nil {
		return nil, fmt.Errorf("failed to create cache directory: %w", err)
	}

	if c.JavaPath == "" {
		c.JavaPath = "java"
	}
	if _, err := exec.LookPath(c.JavaPath); err != nil {
		return nil, fmt.Errorf("could not validate Java is installed: %w", err)
	}

	return &Integration{l: log, c: c}, nil
}

// CommonConfig satisfies Integration.CommonConfig.
func (i *Integration) CommonConfig() config.Common { return i.c.CommonConfig }

// Name satisfies Integration.Name.
func (i *Integration) Name() string { return "jmx_exporter" }

// RegisterRoutes satisfies Integration.RegisterRoutes.
func (i *Integration) RegisterRoutes(r *mux.Router) error {
	proxyURL, err := url.Parse(fmt.Sprintf("http://%s/metrics", i.c.ListenAddress))
	if err != nil {
		return err
	}

	var reverseProxy httputil.ReverseProxy
	reverseProxy.Director = func(r *http.Request) {
		r.URL = proxyURL
	}
	r.Handle("/metrics", &reverseProxy)

	return nil
}

// ScrapeConfigs satisfies Integration.ScrapeConfigs.
func (i *Integration) ScrapeConfigs() []config.ScrapeConfig {
	return []config.ScrapeConfig{{
		JobName:     i.Name(),
		MetricsPath: "/metrics",
	}}
}

// Run satisfies Integration.Run.
func (i *Integration) Run(ctx context.Context) error {
	// Download the JAR for running.
	jarPath := filepath.Join(i.c.CacheDirectory, JMXCacheSubpath, JMXExporterFilename)
	err := downloader.Global.Download(ctx, JMXExporterURL, jarPath, JMXExporterSHA256)
	if err != nil {
		return fmt.Errorf("failed to download jmx_exporter jar: %w", err)
	}

	// Create config file on disk so the JAR can read it.
	configFile, err := ioutil.TempFile(os.TempDir(), "*-jmx_exporter-config.yml")
	if err != nil {
		return fmt.Errorf("couldn't create config file for jmx_exporter: %w", err)
	}
	defer func() {
		os.Remove(configFile.Name())
	}()

	err = yaml.NewEncoder(configFile).Encode(i.c.ExporterConfig)
	configFile.Close()
	if err != nil {
		return fmt.Errorf("failed to write config file for jmx_exporter: %w", err)
	}

	// Run the JAR.
	args := append(
		i.c.JVMOptions,
		"-jar", jarPath,
		i.c.ListenAddress,
		configFile.Name(),
	)

	cmd := exec.CommandContext(ctx, i.c.JavaPath, args...)
	cmd.Stdout = &util.LogWriter{Log: level.Debug(i.l)}
	cmd.Stderr = &util.LogWriter{Log: level.Warn(i.l)}

	err = cmd.Run()
	if ctx.Err() != nil {
		return ctx.Err()
	}
	return err
}
