package cluster

import (
	"flag"
	"strings"
	"time"

	util_log "github.com/cortexproject/cortex/pkg/util/log"
	"github.com/grafana/agent/pkg/metrics/cluster/client"
	flagutil "github.com/grafana/agent/pkg/util"
	"github.com/grafana/dskit/kv"
	"github.com/grafana/dskit/ring"
)

// DefaultConfig provides default values for the config
var DefaultConfig = *flagutil.DefaultConfigFromFlags(&Config{}).(*Config)

// Config describes how to instantiate a scraping service Server instance.
type Config struct {
	Enabled                    bool                  `yaml:"enabled"`
	ReshardInterval            time.Duration         `yaml:"reshard_interval"`
	ReshardTimeout             time.Duration         `yaml:"reshard_timeout"`
	ClusterReshardEventTimeout time.Duration         `yaml:"cluster_reshard_event_timeout"`
	KVStore                    kv.Config             `yaml:"kvstore"`
	Lifecycler                 ring.LifecyclerConfig `yaml:"lifecycler"`

	DangerousAllowReadingFiles bool `yaml:"dangerous_allow_reading_files"`

	// TODO(rfratto): deprecate scraping_service_client in Agent and replace with this.
	Client                    client.Config `yaml:"-"`
	APIEnableGetConfiguration bool          `yaml:"-"`
}

// UnmarshalYAML implements yaml.Unmarshaler.
func (c *Config) UnmarshalYAML(unmarshal func(interface{}) error) error {
	*c = DefaultConfig

	type plain Config
	err := unmarshal((*plain)(c))
	if err != nil {
		return err
	}
	c.Lifecycler.RingConfig.ReplicationFactor = 1
	return nil
}

// RegisterFlags adds the flags required to config the Server to the given
// FlagSet.
func (c *Config) RegisterFlags(f *flag.FlagSet) {
	c.RegisterFlagsWithPrefix("", f)
}

// RegisterFlagsWithPrefix adds the flags required to config this to the given
// FlagSet with a specified prefix.
func (c *Config) RegisterFlagsWithPrefix(prefix string, f *flag.FlagSet) {
	f.BoolVar(&c.Enabled, prefix+"enabled", false, "enables the scraping service mode")
	f.DurationVar(&c.ReshardInterval, prefix+"reshard-interval", time.Minute*1, "how often to manually refresh configuration")
	f.DurationVar(&c.ReshardTimeout, prefix+"reshard-timeout", time.Second*30, "timeout for refreshing the configuration. Timeout of 0s disables timeout.")
	f.DurationVar(&c.ClusterReshardEventTimeout, prefix+"cluster-reshard-event-timeout", time.Second*30, "timeout for the cluster reshard. Timeout of 0s disables timeout.")
	c.KVStore.RegisterFlagsWithPrefix(prefix+"config-store.", "configurations/", f)
	c.Lifecycler.RegisterFlagsWithPrefix(prefix, f, util_log.Logger)

	// GRPCClientConfig.RegisterFlags expects that prefix does not end in a ".",
	// unlike all other flags.
	noDotPrefix := strings.TrimSuffix(prefix, ".")
	c.Client.GRPCClientConfig.RegisterFlagsWithPrefix(noDotPrefix, f)
}
