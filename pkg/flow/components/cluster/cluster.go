package cluster

import (
	"context"
	"fmt"
	"github.com/alecthomas/units"
	"github.com/buraksezer/olric"
	"github.com/buraksezer/olric/config"
	"github.com/go-kit/log"
	"github.com/go-kit/log/level"
	"github.com/go-redis/redis/v8"
	"github.com/grafana/agent/component"
	"github.com/prometheus/common/model"
	"github.com/prometheus/prometheus/discovery"
	"github.com/prometheus/prometheus/model/relabel"
	"gopkg.in/yaml.v2"
	"net/url"
	"os"
	"sync"
	"time"
)

func init() {
	component.Register(component.Registration{
		Name:      "cluster",
		Args:      Config{},
		Exports:   Exports{},
		Singleton: true,

		Build: func(opts component.Options, args component.Arguments) (component.Component, error) {
			return NewCluster(opts, args.(Config))
		},
	})
	component.RegisterGoStruct("targetchannel", TargetReceiver{})
}

type Cluster struct {
	mut        sync.Mutex
	olricCfg   *config.Config
	server     *olric.Olric
	config     Config
	logger     log.Logger
	client     olric.Client
	configDmap olric.DMap
	target     *TargetReceiver
	hasher     *hasher
	opts       component.Options
}

func NewCluster(opts component.Options, c Config) (*Cluster, error) {

	return &Cluster{
		config: c,
		logger: opts.Logger,
		target: &TargetReceiver{
			children: make([]func([]KeyConfig), 0),
		},
		hasher: newHasher(),
		opts:   opts,
	}, nil
}

func (c *Cluster) Run(parentCtx context.Context) error {
	c.mut.Lock()
	defer c.mut.Unlock()

	// Callback function. It's called when this node is ready to accept connections.
	ctx, cancel := context.WithCancel(context.Background())
	if c.config.Self == "" {
		n, _ := os.Hostname()
		c.config.Self = n
	}
	env := "lan"
	if len(c.config.Peers) == 0 {
		env = "local"
	}
	c.olricCfg = config.New(env)
	c.olricCfg.Started = func() {
		defer cancel()
		level.Info(c.logger).Log("msg", "cluster is ready to accept connections")
	}
	db, err := olric.New(c.olricCfg)
	if err != nil {
		return err
	}
	c.server = db

	// Start the instance. It will form a single-node cluster.
	go func() {
		// Call Start at background. It's a blocker call.
		err := c.server.Start()
		if err != nil {
			level.Error(c.logger).Log("err", err, "msg", "cluster failed to start")
		}
	}()

	<-ctx.Done()
	e := c.server.NewEmbeddedClient()
	c.client = e
	_ = c.setupConfigs()
	c.opts.OnStateChange(Exports{Output: c.target})
	go c.scan()
	<-parentCtx.Done()
	return nil
}

func (c *Cluster) Update(args component.Arguments) error {
	return nil
}

func (c *Cluster) setupConfigs() error {
	ctx := context.Background()
	configDmap, err := c.client.NewDMap("configs")
	if err != nil {
		return err
	}

	c.configDmap = configDmap
	pubsub, err := c.client.NewPubSub()
	if err != nil {
		return err
	}
	err = c.getAndSendKeys()
	if err != nil {
		return err
	}
	redisPubSub := pubsub.Subscribe(ctx, "config-updates")
	msg := redisPubSub.Channel()
	go func() {
		c.receiveMessage(msg)
	}()
	return nil
}

func (c *Cluster) scan() {
	t := time.NewTicker(10 * time.Second)
	for {
		select {
		case <-t.C:
			err := c.getAndSendKeys()
			if err != nil {
				level.Error(c.logger).Log("err", err, "msg", "error scanning cluster keys")
				continue
			}

		}
	}
}

func (c *Cluster) getAndSendKeys() error {
	ctx := context.Background()
	it, err := c.configDmap.Scan(ctx)
	if err != nil {
		return err
	}
	keyConfigs := make([]string, 0)
	for it.Next() {
		response, err := c.configDmap.Get(ctx, it.Key())
		if err != nil {
			level.Error(c.logger).Log("err", err, "msg", fmt.Sprintf("failed getting key %s", it.Key()))
			continue
		}
		kc, err := response.String()
		if err != nil {
			level.Error(c.logger).Log("err", err, "msg", fmt.Sprintf("failed getting response for key %s", it.Key()))
			continue
		}
		keyConfigs = append(keyConfigs, kc)
	}
	ownedKeys, err := c.getOwnedKeys(keyConfigs, ctx)
	if err != nil {
		return err
	}
	ownedConfigs := make([]KeyConfig, 0)
	for _, k := range ownedKeys {
		response, err := c.configDmap.Get(ctx, k)
		if err != nil {
			return err
		}
		r, err := response.String()
		if err != nil {
			return err
		}
		kc := &KeyConfig{}
		err = yaml.Unmarshal([]byte(r), kc)
		if err != nil {
			return err
		}
		kc.KeyName = r
		ownedConfigs = append(ownedConfigs, *kc)
	}
	if len(ownedConfigs) == 0 {
		// lets spit something out for debugging
		ownedConfigs = append(ownedConfigs, KeyConfig{
			KeyName:       "key1",
			Name:          "Super secret friend key",
			ScrapeConfigs: nil,
		})
	}
	go c.target.Send(ownedConfigs)
	return nil
}

func (c *Cluster) receiveMessage(msg <-chan *redis.Message) {
	for {
		select {
		case m := <-msg:
			ctx := context.Background()
			keyName := m.Payload
			ownedKeys, err := c.getOwnedKeys([]string{keyName}, ctx)
			if err != nil {
				level.Error(c.logger).Log("err", err, "msg", fmt.Sprintf("failure getting ownership for key %s", keyName))
				continue
			}
			if len(ownedKeys) == 0 {
				continue
			}

			// Key is owned so lets get the configuration
			configVal, err := c.configDmap.Get(ctx, keyName)
			if err != nil {
				level.Error(c.logger).Log("err", err, "msg", fmt.Sprintf("failure getting value for key %s", keyName))
				continue
			}
			cfgStr, err := configVal.String()
			if err != nil {
				level.Error(c.logger).Log("err", err, "msg", fmt.Sprintf("failure converting value for key %s", keyName))
				continue
			}
			kc := &KeyConfig{}
			err = yaml.Unmarshal([]byte(cfgStr), kc)
			if err != nil {
				level.Error(c.logger).Log("err", err, "msg", "failure processing message", "message", cfgStr)
				continue
			}
			kc.KeyName = keyName
			go c.target.Send([]KeyConfig{*kc})
		}
	}
}

func (c *Cluster) getOwnedKeys(keys []string, ctx context.Context) ([]string, error) {
	// Check to see if this key is owned
	members, err := c.client.Members(ctx)
	if err != nil {
		return nil, err
	}

	membersStr := make([]string, 0)
	for _, member := range members {
		membersStr = append(membersStr, member.Name)
	}
	ownedKeys := c.hasher.ownedKeys(keys, c.config.Self, membersStr)
	return ownedKeys, nil
}

type Config struct {
	Peers []string `hcl:"peers,optional"`
	Self  string   `hcl:"self,optional"`
}

type Exports struct {
	Output *TargetReceiver `hcl:"output"`
}

type KeyConfig struct {
	KeyName       string
	Name          string             `yaml:"name,omitempty"`
	ScrapeConfigs []*KeyScrapeConfig `yaml:"scrape_configs,omitempty"`
}

// KeyScrapeConfig configures a scraping unit for Prometheus.
type KeyScrapeConfig struct {
	// The job name to which the job label is set by default.
	JobName string `yaml:"job_name"`
	// Indicator whether the scraped metrics should remain unmodified.
	HonorLabels bool `yaml:"honor_labels,omitempty"`
	// Indicator whether the scraped timestamps should be respected.
	HonorTimestamps bool `yaml:"honor_timestamps"`
	// A set of query parameters with which the target is scraped.
	Params url.Values `yaml:"params,omitempty"`
	// How frequently to scrape the targets of this scrape config.
	ScrapeInterval model.Duration `yaml:"scrape_interval,omitempty"`
	// The timeout for scraping targets of this config.
	ScrapeTimeout model.Duration `yaml:"scrape_timeout,omitempty"`
	// The HTTP resource path on which to fetch metrics from targets.
	MetricsPath string `yaml:"metrics_path,omitempty"`
	// The URL scheme with which to fetch metrics from targets.
	Scheme string `yaml:"scheme,omitempty"`
	// An uncompressed response body larger than this many bytes will cause the
	// scrape to fail. 0 means no limit.
	BodySizeLimit units.Base2Bytes `yaml:"body_size_limit,omitempty"`
	// More than this many samples post metric-relabeling will cause the scrape to
	// fail.
	SampleLimit uint `yaml:"sample_limit,omitempty"`
	// More than this many targets after the target relabeling will cause the
	// scrapes to fail.
	TargetLimit uint `yaml:"target_limit,omitempty"`
	// More than this many labels post metric-relabeling will cause the scrape to
	// fail.
	LabelLimit uint `yaml:"label_limit,omitempty"`
	// More than this label name length post metric-relabeling will cause the
	// scrape to fail.
	LabelNameLengthLimit uint `yaml:"label_name_length_limit,omitempty"`
	// More than this label value length post metric-relabeling will cause the
	// scrape to fail.
	LabelValueLengthLimit uint `yaml:"label_value_length_limit,omitempty"`

	// We cannot do proper Go type embedding below as the parser will then parse
	// values arbitrarily into the overflow maps of further-down types.

	ServiceDiscoveryConfigs discovery.Configs `yaml:"-"`

	// List of metric relabel configurations.
	MetricRelabelConfigs []*relabel.Config `yaml:"metric_relabel_configs,omitempty"`
}
