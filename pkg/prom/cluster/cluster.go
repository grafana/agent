package cluster

import (
	"context"
	"fmt"
	"sync"

	"github.com/go-kit/kit/log"
	"github.com/go-kit/kit/log/level"
	"github.com/golang/protobuf/ptypes/empty"
	"github.com/gorilla/mux"
	"github.com/grafana/agent/pkg/agentproto"
	"github.com/grafana/agent/pkg/prom/instance"
	"github.com/grafana/agent/pkg/prom/instance/configstore"
	"github.com/grafana/agent/pkg/util"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/common/config"
	"github.com/prometheus/prometheus/discovery"
	"github.com/prometheus/prometheus/discovery/azure"
	"github.com/prometheus/prometheus/discovery/consul"
	"github.com/prometheus/prometheus/discovery/digitalocean"
	"github.com/prometheus/prometheus/discovery/dns"
	"github.com/prometheus/prometheus/discovery/dockerswarm"
	"github.com/prometheus/prometheus/discovery/ec2"
	"github.com/prometheus/prometheus/discovery/eureka"
	"github.com/prometheus/prometheus/discovery/hetzner"
	"github.com/prometheus/prometheus/discovery/kubernetes"
	"github.com/prometheus/prometheus/discovery/marathon"
	"github.com/prometheus/prometheus/discovery/openstack"
	"github.com/prometheus/prometheus/discovery/scaleway"
	"github.com/prometheus/prometheus/discovery/triton"
	"github.com/prometheus/prometheus/discovery/zookeeper"
	"google.golang.org/grpc"
)

// Cluster connects an Agent to other Agents and allows them to distribute
// workload.
type Cluster struct {
	mut sync.RWMutex

	log            log.Logger
	cfg            Config
	baseValidation ValidationFunc

	//
	// Internally, Cluster glues together four separate pieces of logic.
	// See comments below to get an understanding of what is going on.
	//

	// node manages membership in the cluster and performs cluster-wide reshards.
	node *node

	// store connects to a configstore for changes. storeAPI is an HTTP API for it.
	store    *configstore.Remote
	storeAPI *configstore.API

	// watcher watches the store and applies changes to an instance.Manager,
	// triggering metrics to be collected and sent. configWatcher also does a
	// complete refresh of its state on an interval.
	watcher *configWatcher
}

// New creates a new Cluster.
func New(
	l log.Logger,
	reg prometheus.Registerer,
	cfg Config,
	im instance.Manager,
	validate ValidationFunc,
) (*Cluster, error) {
	l = log.With(l, "component", "cluster")

	var (
		c   = &Cluster{log: l, cfg: cfg, baseValidation: validate}
		err error
	)

	// Hold the lock for the initialization. This is necessary since newNode will
	// eventually call Reshard, and we want c.watcher to be initialized when that
	// happens.
	c.mut.Lock()
	defer c.mut.Unlock()

	c.node, err = newNode(reg, l, cfg, c)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize node membership: %w", err)
	}

	c.store, err = configstore.NewRemote(l, reg, cfg.KVStore, cfg.Enabled)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize configstore: %w", err)
	}
	c.storeAPI = configstore.NewAPI(l, c.store, c.storeValidate)
	reg.MustRegister(c.storeAPI)

	c.watcher, err = newConfigWatcher(l, cfg, c.store, im, c.node.Owns, validate)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize configwatcher: %w", err)
	}

	// NOTE(rfratto): ApplyConfig isn't necessary for the initialization but must
	// be called for any changes to the configuration.
	return c, nil
}

func (c *Cluster) storeValidate(cfg *instance.Config) error {
	c.mut.RLock()
	defer c.mut.RUnlock()

	if err := c.baseValidation(cfg); err != nil {
		return err
	}

	if c.cfg.DangerousAllowReadingFiles {
		return nil
	}

	// If configs aren't allowed to read from the store, we need to make sure no
	// configs coming in from the API set files for passwords.
	for i, rw := range cfg.RemoteWrite {
		if err := validateNoFiles(&rw.HTTPClientConfig); err != nil {
			return fmt.Errorf("failed to validate remote_write at index %d: %w", i, err)
		}
	}

	for i, sc := range cfg.ScrapeConfigs {
		if err := validateNoFiles(&sc.HTTPClientConfig); err != nil {
			return fmt.Errorf("failed to validate scrape_config at index %d: %w", i, err)
		}

		for j, disc := range sc.ServiceDiscoveryConfigs {
			if err := validateDiscoveryNoFiles(disc); err != nil {
				return fmt.Errorf("failed to validate service discovery at index %d withini scrape_config at index %d: %w", j, i, err)
			}
		}
	}

	return nil
}

func validateNoFiles(cfg *config.HTTPClientConfig) error {
	checks := []struct {
		name  string
		check func() bool
	}{
		{"bearer_token_file", func() bool { return cfg.BearerTokenFile != "" }},
		{"password_file", func() bool { return cfg.BasicAuth != nil && cfg.BasicAuth.PasswordFile != "" }},
		{"credentials_file", func() bool { return cfg.Authorization != nil && cfg.Authorization.CredentialsFile != "" }},
		{"ca_file", func() bool { return cfg.TLSConfig.CAFile != "" }},
		{"cert_file", func() bool { return cfg.TLSConfig.CertFile != "" }},
		{"key_file", func() bool { return cfg.TLSConfig.KeyFile != "" }},
	}
	for _, check := range checks {
		if check.check() {
			return fmt.Errorf("%s must be empty unless dangerous_allow_reading_files is set", check.name)
		}
	}
	return nil
}

func validateDiscoveryNoFiles(disc discovery.Config) error {
	switch d := disc.(type) {
	case *azure.SDConfig:
		// no-op
	case *consul.SDConfig:
		if err := validateNoFiles(&config.HTTPClientConfig{TLSConfig: d.TLSConfig}); err != nil {
			return err
		}
	case *digitalocean.SDConfig:
		if err := validateNoFiles(&d.HTTPClientConfig); err != nil {
			return err
		}
	case *dns.SDConfig:
		// no-op
	case *dockerswarm.SDConfig:
		if err := validateNoFiles(&d.HTTPClientConfig); err != nil {
			return err
		}
	case *ec2.SDConfig:
		// no-op
	case *eureka.SDConfig:
		if err := validateNoFiles(&d.HTTPClientConfig); err != nil {
			return err
		}
	case *hetzner.SDConfig:
		if err := validateNoFiles(&d.HTTPClientConfig); err != nil {
			return err
		}
	case *kubernetes.SDConfig:
		if err := validateNoFiles(&d.HTTPClientConfig); err != nil {
			return err
		}
	case *marathon.SDConfig:
		if err := validateNoFiles(&d.HTTPClientConfig); err != nil {
			return err
		}
		if d.AuthTokenFile != "" {
			return fmt.Errorf("auth_token_file must be empty unless dangerous_allow_reading_files is set")
		}
	case *openstack.SDConfig:
		if err := validateNoFiles(&config.HTTPClientConfig{TLSConfig: d.TLSConfig}); err != nil {
			return err
		}
	case *scaleway.SDConfig:
		if err := validateNoFiles(&d.HTTPClientConfig); err != nil {
			return err
		}
	case *triton.SDConfig:
		if err := validateNoFiles(&config.HTTPClientConfig{TLSConfig: d.TLSConfig}); err != nil {
			return err
		}
	case *zookeeper.NerveSDConfig:
		// no-op
	case *zookeeper.ServersetSDConfig:
		// no-op
	default:
		return fmt.Errorf("unknown service discovery %s; rejecting config for safety. set dangerous_allow_reading_files to ignore", d.Name())
	}

	return nil
}

// Reshard implements agentproto.ScrapingServiceServer, and syncs the state of
// configs with the configstore.
func (c *Cluster) Reshard(ctx context.Context, _ *agentproto.ReshardRequest) (*empty.Empty, error) {
	err := c.watcher.Refresh(ctx)
	if err != nil {
		level.Error(c.log).Log("msg", "failed to perform local reshard", "err", err)
	}
	return &empty.Empty{}, err
}

// ApplyConfig applies configuration changes to Cluster.
func (c *Cluster) ApplyConfig(
	cfg Config,
) error {
	c.mut.Lock()
	defer c.mut.Unlock()

	if util.CompareYAML(c.cfg, cfg) {
		return nil
	}

	if err := c.node.ApplyConfig(cfg); err != nil {
		return fmt.Errorf("failed to apply config to node membership: %w", err)
	}

	if err := c.store.ApplyConfig(cfg.Lifecycler.RingConfig.KVStore, cfg.Enabled); err != nil {
		return fmt.Errorf("failed to apply config to config store: %w", err)
	}

	if err := c.watcher.ApplyConfig(cfg); err != nil {
		return fmt.Errorf("failed to apply config to watcher: %w", err)
	}

	c.cfg = cfg

	// Force a refresh so all the configs get updated with new defaults.
	level.Info(c.log).Log("msg", "cluster config changed, refreshing from configstore in background")
	go func() {
		ctx := context.Background()
		if c.cfg.ReshardTimeout > 0 {
			var cancel context.CancelFunc
			ctx, cancel = context.WithTimeout(ctx, c.cfg.ReshardTimeout)
			defer cancel()
		}
		err := c.watcher.Refresh(ctx)
		if err != nil {
			level.Error(c.log).Log("msg", "failed to perform local reshard", "err", err)
		}
	}()

	return nil
}

// WireAPI injects routes into the provided mux router for the config
// management API.
func (c *Cluster) WireAPI(r *mux.Router) {
	c.storeAPI.WireAPI(r)
	c.node.WireAPI(r)
}

// WireGRPC injects gRPC server handlers into the provided gRPC server.
func (c *Cluster) WireGRPC(srv *grpc.Server) {
	agentproto.RegisterScrapingServiceServer(srv, c)
}

// Stop stops the cluster and all of its dependencies.
func (c *Cluster) Stop() {
	c.mut.Lock()
	defer c.mut.Unlock()

	deps := []struct {
		name   string
		closer func() error
	}{
		{"node", c.node.Stop},
		{"config store", c.store.Close},
		{"config watcher", c.watcher.Stop},
	}
	for _, dep := range deps {
		err := dep.closer()
		if err != nil {
			level.Error(c.log).Log("msg", "failed to stop dependency", "dependency", dep.name, "err", err)
		}
	}
}
