package eventhandler

import (
	"fmt"
	"path/filepath"

	"github.com/go-kit/log"
	"github.com/grafana/agent/pkg/integrations/v2"
	"k8s.io/client-go/util/homedir"
)

var DefaultConfig = Config{
	SendTimeout:  500,
	ClusterName:  "cloud",
	CachePath:    "./cache/eventhandler.cache",
	LogsInstance: "default",
	InCluster:    false,
}

// Config controls the EventHandler integration.
type Config struct {
	// todo: what should this be set to? check it out when investigating queueing, etc.
	SendTimeout    int    `yaml:"send_timeout,omitempty"` // milliseconds
	ClusterName    string `yaml:"cluster_name,omitempty"`
	KubeconfigPath string `yaml:"kubeconfig_path,omitempty"`
	CachePath      string `yaml:"cache_path,omitempty"`
	LogsInstance   string `yaml:"logs_instance,omitempty"`
	InCluster      bool   `yaml:"in_cluster,omitempty"`
}

// UnmarshalYAML implements yaml.Unmarshaler for Config
// TODO: understand & test
func (c *Config) UnmarshalYAML(unmarshal func(interface{}) error) error {
	*c = DefaultConfig

	type plain Config
	return unmarshal((*plain)(c))
}

// Name returns the name of the integration that this config represents.
func (c *Config) Name() string { return "eventhandler" }

// ApplyDefaults applies runtime-specific defaults to c.
func (c *Config) ApplyDefaults(globals integrations.Globals) error {
	// not in cluster and KC path not set,
	// try to use default kubeconfig path
	if !c.InCluster && c.KubeconfigPath == "" {
		if home := homedir.HomeDir(); home != "" {
			c.KubeconfigPath = filepath.Join(home, ".kube", "config")
		} else {
			// unable to find KC
			return fmt.Errorf("could not locate a kubeconfig. please set the kubeconfig_path in your agent config")
		}
	}
	return nil
}

// Identifier uniquely identifies this instance of Config.
func (c *Config) Identifier(globals integrations.Globals) (string, error) {
	return globals.AgentIdentifier, nil
}

// NewIntegration converts this config into an instance of an integration.
func (c *Config) NewIntegration(l log.Logger, globals integrations.Globals) (integrations.Integration, error) {
	return newEventHandler(
		l,
		globals,
		c.LogsInstance,
		c.CachePath,
		c.KubeconfigPath,
		c.InCluster,
		c.ClusterName,
		c.SendTimeout,
	)
}

func init() {
	integrations.Register(&Config{}, integrations.TypeSingleton)
}
