package eventhandler

import (
	"fmt"
	"path/filepath"

	"github.com/go-kit/log"
	"github.com/grafana/agent/pkg/integrations/v2"
	"k8s.io/client-go/util/homedir"
)

// Config defaults
var DefaultConfig = Config{
	SendTimeout:    60,
	ClusterName:    "cloud",
	CachePath:      "./.eventcache/eventhandler.cache",
	LogsInstance:   "default",
	InCluster:      false,
	InformerResync: 120,
}

// Config configures the eventhandler integration
type Config struct {
	// eventhandler hands events off to promtail using a promtail
	// client channel. this configures how long to wait on the channel
	// before abandoning and moving on
	SendTimeout int `yaml:"send_timeout,omitempty"` // seconds
	// configures a cluster= label for log lines
	ClusterName string `yaml:"cluster_name,omitempty"`
	// path to kubeconfig. if omitted will look in user's home dir.
	// this isn't used if InCluster is set to true
	KubeconfigPath string `yaml:"kubeconfig_path,omitempty"`
	// path to a cache file that will store a log of timestamps and events
	// shipped for those timestamps. used to prevent double-shipping on informer
	// restart / relist
	CachePath string `yaml:"cache_path,omitempty"`
	// name of logs subsystem instance to hand events off to
	LogsInstance string `yaml:"logs_instance,omitempty"`
	// are we running in cluster?
	InCluster bool `yaml:"in_cluster,omitempty"`
	// informer resync interval. out of scope to describe this here.
	InformerResync int `yaml:"informer_resync,omitempty"` // seconds
}

// UnmarshalYAML implements yaml.Unmarshaler for Config
func (c *Config) UnmarshalYAML(unmarshal func(interface{}) error) error {
	*c = DefaultConfig

	type plain Config
	return unmarshal((*plain)(c))
}

// Name returns the name of the integration that this config represents
func (c *Config) Name() string { return "eventhandler" }

// ApplyDefaults applies runtime-specific defaults to c
func (c *Config) ApplyDefaults(globals integrations.Globals) error {
	// if not running in cluster and KC path not set,
	// use default kubeconfig path in user's home dir
	if !c.InCluster && c.KubeconfigPath == "" {
		if home := homedir.HomeDir(); home != "" {
			c.KubeconfigPath = filepath.Join(home, ".kube", "config")
		} else {
			// unable to return home dir
			return fmt.Errorf("could not set home dir and find a kubeconfig. please set kubeconfig_path in your agent config or set in_cluster to true")
		}
	}
	return nil
}

// Identifier uniquely identifies this instance of Config
func (c *Config) Identifier(globals integrations.Globals) (string, error) {
	return globals.AgentIdentifier, nil
}

// NewIntegration converts this config into an instance of an integration.
func (c *Config) NewIntegration(l log.Logger, globals integrations.Globals) (integrations.Integration, error) {
	return newEventHandler(l, globals, c)
}

func init() {
	integrations.Register(&Config{}, integrations.TypeSingleton)
}
