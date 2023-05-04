package vsphere

import (
	"flag"
	"net/url"
	"time"

	"github.com/vmware/govmomi/vim25/soap"
)

// Config represents needed config structure for vSphere exporter
type Config struct {
	ListenAddr              string
	TelemetryPath           string
	TLSConfigPath           string
	ChunkSize               int
	CollectConcurrency      int
	VSphereURL              *url.URL
	ObjectDiscoveryInterval time.Duration
	EnableExporterMetrics   bool
}

var defaultConfig = &Config{
	ListenAddr:              ":9237",
	TelemetryPath:           "/metrics",
	TLSConfigPath:           "",
	ChunkSize:               5,
	CollectConcurrency:      8,
	ObjectDiscoveryInterval: 0,
	EnableExporterMetrics:   false,
}

type soapURLFlag struct {
	u *url.URL
}

func (v soapURLFlag) String() string {
	if v.u != nil {
		return v.u.String()
	}
	return ""
}

func (v soapURLFlag) Set(s string) error {
	u, err := soap.ParseURL(s)
	if err != nil {
		return err
	}
	*v.u = *u
	return nil
}

// RegisterFlags register the flag params for the exporter
func (c *Config) RegisterFlags(fs *flag.FlagSet) {
	// Exporter web configs
	{
		fs.StringVar(&c.ListenAddr, "web.listen-address", defaultConfig.ListenAddr,
			"Address on which to expose metrics and web interface.")
		fs.StringVar(&c.TelemetryPath, "web.telemetry-path", defaultConfig.TelemetryPath,
			"Path under which to expose metrics.")
		fs.StringVar(&c.TLSConfigPath, "web.config", defaultConfig.TLSConfigPath,
			"Path to config yaml file that can enable TLS or authentication.")
	}

	// Vsphere client configs
	{
		u := &url.URL{}
		fs.Var(&soapURLFlag{u}, "vsphere.url", "vSphere SDK URL.")
		c.VSphereURL = u
		fs.DurationVar(&c.ObjectDiscoveryInterval, "vsphere.discovery-interval",
			defaultConfig.ObjectDiscoveryInterval,
			"Object discovery duration interval. Discovery will occur per scrape if set to 0.")
		fs.IntVar(&c.ChunkSize, "vsphere.mo-chunk-size", defaultConfig.ChunkSize,
			"Managed object reference chunk size to use when fetching from vSphere.")
		fs.IntVar(&c.CollectConcurrency, "vsphere.concurrent-requests",
			defaultConfig.CollectConcurrency,
			"The number of concurrent requests to make while fetching metrics from vSphere.")
	}

	// Misc configs
	{
		fs.BoolVar(&c.EnableExporterMetrics, "exporter.metrics.enable",
			defaultConfig.EnableExporterMetrics,
			"Enable metrics to observe exporter behavior.")
	}
}
