package kafka

import (
	"context"
	"reflect"
	"sync"

	"github.com/Shopify/sarama"
	"github.com/go-kit/log/level"
	"github.com/grafana/dskit/flagext"

	"github.com/grafana/agent/component"
	"github.com/grafana/agent/component/common/config"
	"github.com/grafana/agent/component/common/loki"
	flow_relabel "github.com/grafana/agent/component/common/relabel"
	kt "github.com/grafana/agent/component/loki/source/kafka/internal/kafkatarget"
	"github.com/grafana/loki/clients/pkg/promtail/scrapeconfig"
	"github.com/prometheus/common/model"
)

func init() {
	component.Register(component.Registration{
		Name: "loki.source.kafka",
		Args: Arguments{},

		Build: func(opts component.Options, args component.Arguments) (component.Component, error) {
			return New(opts, args.(Arguments))
		},
	})
}

// Arguments holds values which are used to configure the loki.source.kafka
// component.
type Arguments struct {
	Labels               map[string]string   `river:"labels,attr,optional"`
	UseIncomingTimestamp bool                `river:"use_incoming_timestamp,attr,optional"`
	Brokers              []string            `river:"brokers,attr"`
	GroupID              string              `river:"group_id,attr"`
	Topics               []string            `river:"topics,attr"`
	Version              string              `river:"version,attr"`
	Assignor             string              `river:"assignor,attr"`
	Authentication       KafkaAuthentication `river:"authentication"`

	ForwardTo    []loki.LogsReceiver `river:"forward_to,attr"`
	RelabelRules flow_relabel.Rules  `river:"relabel_rules,attr,optional"`
}

// KafkaAuthentication describe the configuration for authentication with Kafka brokers
type KafkaAuthentication struct {
	// Type is authentication type
	// Possible values: none, sasl and ssl (defaults to none).
	Type KafkaAuthenticationType `river:"type,attr"`

	// TLSConfig is used for TLS encryption and authentication with Kafka brokers
	TLSConfig config.TLSConfig `river:"tls_config,block,optional"`

	// SASLConfig is used for SASL authentication with Kafka brokers
	SASLConfig KafkaSASLConfig `yaml:"sasl_config,omitempty"`
}

// KafkaAuthenticationType specifies method to authenticate with Kafka brokers
type KafkaAuthenticationType string

// KafkaSASLConfig describe the SASL configuration for authentication with Kafka brokers
type KafkaSASLConfig struct {
	// SASL mechanism. Supports PLAIN, SCRAM-SHA-256 and SCRAM-SHA-512
	Mechanism sarama.SASLMechanism `river:"mechanism,attr"`

	// SASL Username
	User string `river:"user,attr"`

	// SASL Password for the User
	Password flagext.Secret `river:"password,attr"`

	// UseTLS sets whether TLS is used with SASL
	UseTLS bool `river:"use_tls,attr"`

	// TLSConfig is used for SASL over TLS. It is used only when UseTLS is true
	TLSConfig config.TLSConfig `river:"tls_config,block,optional"`
}

// ListenerConfig defines a kafka listener.
type ListenerConfig struct {
	ListenAddress string `river:"address,attr,optional"`
	ListenPort    int    `river:"port,attr"`
	// TODO - add the rest of the server config from Promtail
}

// DefaultListenerConfig provides the default arguments for a kafka listener.
var DefaultListenerConfig = ListenerConfig{
	ListenAddress: "0.0.0.0",
}

// UnmarshalRiver implements river.Unmarshaler.
func (lc *ListenerConfig) UnmarshalRiver(f func(interface{}) error) error {
	*lc = DefaultListenerConfig

	type kafkacfg ListenerConfig
	err := f((*kafkacfg)(lc))
	if err != nil {
		return err
	}

	return nil
}

// Component implements the loki.source.kafka component.
type Component struct {
	opts component.Options

	mut    sync.RWMutex
	args   Arguments
	fanout []loki.LogsReceiver
	target *kt.TargetSyncer

	handler loki.LogsReceiver
}

// New creates a new loki.source.kafka component.
func New(o component.Options, args Arguments) (*Component, error) {
	c := &Component{
		opts:    o,
		mut:     sync.RWMutex{},
		args:    Arguments{},
		fanout:  args.ForwardTo,
		target:  nil,
		handler: make(loki.LogsReceiver),
	}

	// Call to Update() to start readers and set receivers once at the start.
	if err := c.Update(args); err != nil {
		return nil, err
	}

	return c, nil
}

// Run implements component.Component.
func (c *Component) Run(ctx context.Context) error {
	defer func() {
		c.mut.Lock()
		defer c.mut.Unlock()

		level.Info(c.opts.Logger).Log("msg", "loki.source.kafka component shutting down, stopping listener")
		if c.target != nil {
			err := c.target.Stop()
			if err != nil {
				level.Error(c.opts.Logger).Log("msg", "error while stopping kafka listener", "err", err)
			}
		}
	}()

	for {
		select {
		case <-ctx.Done():
			return nil
		case entry := <-c.handler:
			c.mut.RLock()
			for _, receiver := range c.fanout {
				receiver <- entry
			}
			c.mut.RUnlock()
		}
	}
}

// Update implements component.Component.
func (c *Component) Update(args component.Arguments) error {
	c.mut.Lock()
	defer c.mut.Unlock()

	newArgs := args.(Arguments)
	c.fanout = newArgs.ForwardTo

	/* TODO
	var rcs []*relabel.Config
	if newArgs.RelabelRules != nil && len(newArgs.RelabelRules) > 0 {
		rcs = flow_relabel.ComponentToPromRelabelConfigs(newArgs.RelabelRules)
	}
	*/

	if listenerChanged(c.args, newArgs) || relabelRulesChanged(c.args.RelabelRules, newArgs.RelabelRules) {
		if c.target != nil {
			err := c.target.Stop()
			if err != nil {
				level.Error(c.opts.Logger).Log("msg", "error while stopping kafka listener", "err", err)
			}
		}

		entryHandler := loki.NewEntryHandler(c.handler, func() {})
		t, err := kt.NewSyncer(c.opts.Registerer, c.opts.Logger, c.args.Convert(), entryHandler)
		if err != nil {
			level.Error(c.opts.Logger).Log("msg", "failed to create kafka client with provided config", "err", err)
			return err
		}

		c.target = t
		c.args = newArgs
	}

	return nil
}

// Convert is used to bridge between the River and Promtail types.
func (args *Arguments) Convert() scrapeconfig.Config {
	lbls := make(model.LabelSet, len(args.Labels))
	for k, v := range args.Labels {
		lbls[model.LabelName(k)] = model.LabelValue(v)
	}

	return scrapeconfig.Config{}
}

// DebugInfo returns information about the status of listener.
func (c *Component) DebugInfo() interface{} {
	c.mut.RLock()
	defer c.mut.RUnlock()

	var res readerDebugInfo = readerDebugInfo{
		//Ready: c.target.Ready(),
	}

	return res
}

type readerDebugInfo struct {
	//Ready bool `river:"ready,attr"`
}

func listenerChanged(prev, next Arguments) bool {
	return !reflect.DeepEqual(prev, next)
}
func relabelRulesChanged(prev, next flow_relabel.Rules) bool {
	return !reflect.DeepEqual(prev, next)
}
