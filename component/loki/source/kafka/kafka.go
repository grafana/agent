package kafka

import (
	"context"
	"sync"

	"github.com/Shopify/sarama"
	"github.com/go-kit/log/level"

	"github.com/grafana/agent/component"
	"github.com/grafana/agent/component/common/loki"
	flow_relabel "github.com/grafana/agent/component/common/relabel"
	kt "github.com/grafana/agent/component/loki/source/kafka/internal/kafkatarget"
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
	KafkaTargetConfig kt.KafkaTargetConfig `river:",squash"`
	ForwardTo         []loki.LogsReceiver  `river:"forward_to,attr"`
	RelabelRules      flow_relabel.Rules   `river:"relabel_rules,attr,optional"`
}

// DefaultArguments provides the default arguments for a kafka component.
var DefaultArguments = Arguments{
	KafkaTargetConfig: kt.KafkaTargetConfig{
		GroupID:  "loki.source.kafka",
		Assignor: "range",
		Version:  "2.2.1",
		Authentication: kt.KafkaAuthentication{
			Type: "none",
			SASLConfig: kt.KafkaSASLConfig{
				Mechanism: sarama.SASLTypePlaintext,
				UseTLS:    false,
			},
		},
		UseIncomingTimestamp: false,
	},
}

// UnmarshalRiver implements river.Unmarshaler.
func (a *Arguments) UnmarshalRiver(f func(interface{}) error) error {
	*a = DefaultArguments

	type kafkacfg Arguments
	err := f((*kafkacfg)(a))
	if err != nil {
		return err
	}

	return nil
}

// Component implements the loki.source.kafka component.
type Component struct {
	opts component.Options

	mut    sync.RWMutex
	fanout []loki.LogsReceiver
	target *kt.TargetSyncer

	handler loki.LogsReceiver
}

// New creates a new loki.source.kafka component.
func New(o component.Options, args Arguments) (*Component, error) {
	c := &Component{
		opts:    o,
		mut:     sync.RWMutex{},
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

		level.Info(c.opts.Logger).Log("msg", "loki.source.kafka component shutting down, stopping target")
		if c.target != nil {
			err := c.target.Stop()
			if err != nil {
				level.Error(c.opts.Logger).Log("msg", "error while stopping kafka target", "err", err)
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

	if c.target != nil {
		err := c.target.Stop()
		if err != nil {
			level.Error(c.opts.Logger).Log("msg", "error while stopping kafka target", "err", err)
		}
	}

	entryHandler := loki.NewEntryHandler(c.handler, func() {})
	t, err := kt.NewSyncer(c.opts.Registerer, c.opts.Logger, newArgs.KafkaTargetConfig, newArgs.RelabelRules, entryHandler)
	if err != nil {
		level.Error(c.opts.Logger).Log("msg", "failed to create kafka client with provided config", "err", err)
		return err
	}

	c.target = t

	return nil
}

// Convert is used to bridge between the River and Promtail types.
/*
func (args *Arguments) Convert() scrapeconfig.Config {
	lbls := make(model.LabelSet, len(args.Labels))
	for k, v := range args.Labels {
		lbls[model.LabelName(k)] = model.LabelValue(v)
	}

	return scrapeconfig.Config{
		KafkaConfig: &scrapeconfig.KafkaTargetConfig{
			Labels:               lbls,
			UseIncomingTimestamp: args.KafkaTargetConfig.UseIncomingTimestamp,
			Brokers:              args.KafkaTargetConfig.Brokers,
			GroupID:              args.KafkaTargetConfig.GroupID,
			Topics:               args.KafkaTargetConfig.Topics,
			Version:              args.KafkaTargetConfig.Version,
			Assignor:             args.KafkaTargetConfig.Assignor,
			Authentication:       args.KafkaTargetConfig.Authentication.Convert(),
		},
		RelabelConfigs: flow_relabel.ComponentToPromRelabelConfigs(args.RelabelRules),
	}
}
*/
