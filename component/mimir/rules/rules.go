package rules

import (
	"context"
	"time"

	"github.com/go-kit/log"
	"github.com/go-kit/log/level"
	"github.com/grafana/agent/component"
	"github.com/grafana/agent/pkg/mimir/client"
	"github.com/grafana/dskit/crypto/tls"
	v1 "github.com/prometheus-operator/prometheus-operator/pkg/apis/monitoring/v1"
)

func init() {
	component.Register(component.Registration{
		Name:    "mimir.rules",
		Args:    Arguments{},
		Exports: Exports{},
		Build: func(o component.Options, c component.Arguments) (component.Component, error) {
			return NewComponent(o, c.(Arguments))
		},
	})
}

type Arguments struct {
	ClientParams ClientArguments `river:"client,block"`
	SyncInterval time.Duration   `river:"sync_interval,attr,optional"`
}

type ClientArguments struct {
	User            string       `river:"user,attr,optional"`
	Key             string       `river:"key,attr,optional"`
	Address         string       `river:"address,attr"`
	ID              string       `river:"id,attr,optional"`
	TLS             TLSArguments `river:"tls,block,optional"`
	UseLegacyRoutes bool         `river:"use_legacy_routes,attr,optional"`
	AuthToken       string       `river:"auth_token,attr,optional"`
}

type TLSArguments struct {
	CertPath           string `river:"tls_cert_path,attr,optional"`
	KeyPath            string `river:"tls_key_path,attr,optional"`
	CAPath             string `river:"tls_ca_path,attr,optional"`
	ServerName         string `river:"tls_server_name,attr,optional"`
	InsecureSkipVerify bool   `river:"tls_insecure_skip_verify,attr,optional"`
	CipherSuites       string `river:"tls_cipher_suites,attr,optional"`
	MinVersion         string `river:"tls_min_version,attr,optional"`
}

type Exports struct {
}

type Component struct {
	log  log.Logger
	opts component.Options
	args Arguments

	client *client.MimirClient
	ticker *time.Ticker
}

var _ component.Component = (*Component)(nil)

func NewComponent(o component.Options, c Arguments) (*Component, error) {
	return &Component{
		log:  o.Logger,
		opts: o,
		args: c,
	}, nil
}

func (c *Component) Run(ctx context.Context) error {
	err := c.init()
	if err != nil {
		return err
	}

	c.start(ctx)

	return nil
}

func (c *Component) Update(newConfig component.Arguments) error {
	c.args = newConfig.(Arguments)
	return c.init()
}

func (c *Component) init() error {
	if c.args.SyncInterval == 0 {
		c.args.SyncInterval = 30 * time.Second
	}

	var err error
	c.client, err = client.New(client.Config{
		User:    c.args.ClientParams.User,
		Key:     c.args.ClientParams.Key,
		Address: c.args.ClientParams.Address,
		ID:      c.args.ClientParams.ID,
		TLS: tls.ClientConfig{
			CertPath:           c.args.ClientParams.TLS.CertPath,
			KeyPath:            c.args.ClientParams.TLS.KeyPath,
			CAPath:             c.args.ClientParams.TLS.CAPath,
			ServerName:         c.args.ClientParams.TLS.ServerName,
			InsecureSkipVerify: c.args.ClientParams.TLS.InsecureSkipVerify,
			CipherSuites:       c.args.ClientParams.TLS.CipherSuites,
			MinVersion:         c.args.ClientParams.TLS.MinVersion,
		},
		UseLegacyRoutes: c.args.ClientParams.UseLegacyRoutes,
		AuthToken:       c.args.ClientParams.AuthToken,
	})
	if err != nil {
		return err
	}

	c.ticker = time.NewTicker(c.args.SyncInterval)

	return nil
}

func (c *Component) start(ctx context.Context) {
	for {
		select {
		case <-c.ticker.C:
			level.Info(c.log).Log("msg", "syncing rules")
			err := c.syncRules(ctx)
			if err != nil {
				level.Error(c.log).Log("msg", "failed to sync rules", "err", err)
			}
		case <-ctx.Done():
			level.Info(c.log).Log("msg", "shutting down")
			return
		}
	}
}

func (c *Component) syncRules(ctx context.Context) error {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	desiredState, err := c.discoverRuleCRDs(ctx)
	if err != nil {
		return err
	}
	level.Debug(c.log).Log("msg", "found rule crds", "num_crds", len(desiredState))

	actualState, err := c.loadActiveRules(ctx)
	if err != nil {
		return err
	}
	level.Debug(c.log).Log("msg", "found active rules", "num_namespaces", len(actualState))

	diff := c.diffRuleStates(desiredState, actualState)

	return c.applyChanges(ctx, diff)
}

func (c *Component) discoverRuleCRDs(ctx context.Context) ([]v1.PrometheusRule, error) {
	return nil, nil
}

func (c *Component) loadActiveRules(ctx context.Context) (map[string][]client.RuleGroup, error) {
	return c.client.ListRules(ctx, "")
}

type RuleGroupDiff struct {
}

func (c *Component) diffRuleStates(desired []v1.PrometheusRule, actual map[string][]client.RuleGroup) []RuleGroupDiff {
	return nil
}

func (c *Component) applyChanges(ctx context.Context, diff []RuleGroupDiff) error {
	return nil
}
