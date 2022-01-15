package v2

import (
	"context"
	"errors"
	"fmt"
	"net/http"

	"github.com/go-kit/log"
	"github.com/grafana/agent/pkg/integrations/shared"
	"github.com/grafana/agent/pkg/integrations/v1/agent"
	"github.com/grafana/agent/pkg/integrations/v1/cadvisor"
	"github.com/grafana/agent/pkg/integrations/v1/consul_exporter"
	"github.com/grafana/agent/pkg/integrations/v1/dnsmasq_exporter"
	"github.com/grafana/agent/pkg/integrations/v1/elasticsearch_exporter"
	"github.com/grafana/agent/pkg/integrations/v1/github_exporter"
	"github.com/grafana/agent/pkg/integrations/v1/kafka_exporter"
	"github.com/grafana/agent/pkg/integrations/v1/memcached_exporter"
	"github.com/grafana/agent/pkg/integrations/v1/mongodb_exporter"
	"github.com/grafana/agent/pkg/integrations/v1/mysqld_exporter"
	"github.com/grafana/agent/pkg/integrations/v1/node_exporter"
	"github.com/grafana/agent/pkg/integrations/v1/postgres_exporter"
	"github.com/grafana/agent/pkg/integrations/v1/process_exporter"
	"github.com/grafana/agent/pkg/integrations/v1/redis_exporter"
	"github.com/grafana/agent/pkg/integrations/v1/statsd_exporter"
	"github.com/grafana/agent/pkg/integrations/v1/windows_exporter"
	"github.com/grafana/agent/pkg/integrations/v2/common"
	"github.com/prometheus/common/model"
)

type Integrations struct {
	Agent                        *Agent                   `yaml:"agent,omitempty"`
	Cadvisor                     *Cadvisor                `yaml:"cadvisor,omitempty"`
	ConsulExporterConfigs        []*ConsulExporter        `yaml:"consul_exporter_configs,omitempty"`
	DnsmasqExporterConfigs       []*DnsmasqExporter       `yaml:"dnsmasq_exporter_configs,omitempty"`
	ElasticsearchExporterConfigs []*ElasticsearchExporter `yaml:"elasticsearch_exporter_configs,omitempty"`
	GithubExporterConfigs        []*GithubExporter        `yaml:"github_exporter_configs,omitempty"`
	KafkaExporterConfigs         []*KafkaExporter         `yaml:"kafka_exporter_configs,omitempty"`
	MemcachedExporterConfigs     []*MemcachedExporter     `yaml:"memcached_exporter_configs,omitempty"`
	MongodbExporterConfigs       []*MongodbExporter       `yaml:"mongodb_exporter_configs,omitempty"`
	MysqldExporterConfigs        []*MysqldExporter        `yaml:"mysqld_exporter_configs,omitempty"`
	NodeExporter                 *NodeExporter            `yaml:"node_exporter,omitempty"`
	PostgresExporterConfigs      []*PostgresExporter      `yaml:"postgres_exporter_configs,omitempty"`
	ProcessExporter              *ProcessExporter         `yaml:"process_exporter,omitempty"`
	RedisExporterConfigs         []*RedisExporter         `yaml:"redis_exporter_configs,omitempty"`
	StatsdExporter               *StatsdExporter          `yaml:"statsd_exporter,omitempty"`
	WindowsExporter              *WindowsExporter         `yaml:"windows_exporter,omitempty"`
}

func (v *Integrations) ActiveConfigs() []Config {
	activeConfigs := make([]Config, 0)
	if v.Agent != nil {
		activeConfigs = append(activeConfigs, v.Agent)
	}
	if v.Cadvisor != nil {
		activeConfigs = append(activeConfigs, v.Cadvisor)
	}
	for _, i := range v.ConsulExporterConfigs {
		activeConfigs = append(activeConfigs, i)
	}
	for _, i := range v.DnsmasqExporterConfigs {
		activeConfigs = append(activeConfigs, i)
	}
	for _, i := range v.ElasticsearchExporterConfigs {
		activeConfigs = append(activeConfigs, i)
	}
	for _, i := range v.GithubExporterConfigs {
		activeConfigs = append(activeConfigs, i)
	}
	for _, i := range v.KafkaExporterConfigs {
		activeConfigs = append(activeConfigs, i)
	}
	for _, i := range v.MemcachedExporterConfigs {
		activeConfigs = append(activeConfigs, i)
	}
	for _, i := range v.MongodbExporterConfigs {
		activeConfigs = append(activeConfigs, i)
	}
	for _, i := range v.MysqldExporterConfigs {
		activeConfigs = append(activeConfigs, i)
	}
	if v.NodeExporter != nil {
		activeConfigs = append(activeConfigs, v.NodeExporter)
	}
	for _, i := range v.PostgresExporterConfigs {
		activeConfigs = append(activeConfigs, i)
	}
	if v.ProcessExporter != nil {
		activeConfigs = append(activeConfigs, v.ProcessExporter)
	}
	for _, i := range v.RedisExporterConfigs {
		activeConfigs = append(activeConfigs, i)
	}
	if v.StatsdExporter != nil {
		activeConfigs = append(activeConfigs, v.StatsdExporter)
	}
	if v.WindowsExporter != nil {
		activeConfigs = append(activeConfigs, v.WindowsExporter)
	}
	return activeConfigs
}

type Agent struct {
	agent.Config `yaml:",omitempty,inline"`
	Cmn          common.MetricsConfig `yaml:",inline"`
}

func (c *Agent) Cfg() Config {
	return c
}

func (c *Agent) Common() common.MetricsConfig {
	return c.Cmn
}

func (c *Agent) ApplyDefaults(globals Globals) error {
	c.Cmn.ApplyDefaults(globals.SubsystemOpts.Metrics.Autoscrape)
	if id, err := c.Identifier(globals); err == nil {
		c.Cmn.InstanceKey = &id
	}
	return nil
}

func (c *Agent) Identifier(globals Globals) (string, error) {
	if c.Cmn.InstanceKey != nil {
		return *c.Cmn.InstanceKey, nil
	}
	return c.Config.InstanceKey(globals.AgentIdentifier)
}

func (c *Agent) NewIntegration(logger log.Logger, globals Globals) (Integration, error) {
	return newIntegration(c, logger, globals, c.Config.NewIntegration)
}

type Cadvisor struct {
	cadvisor.Config `yaml:",omitempty,inline"`
	Cmn             common.MetricsConfig `yaml:",inline"`
}

func (c *Cadvisor) Cfg() Config {
	return c
}

func (c *Cadvisor) Common() common.MetricsConfig {
	return c.Cmn
}

func (c *Cadvisor) UnmarshalYAML(unmarshal func(interface{}) error) error {
	c.Config = cadvisor.DefaultConfig
	type plain Cadvisor
	return unmarshal((*plain)(c))
}

func (c *Cadvisor) ApplyDefaults(globals Globals) error {
	c.Cmn.ApplyDefaults(globals.SubsystemOpts.Metrics.Autoscrape)
	if id, err := c.Identifier(globals); err == nil {
		c.Cmn.InstanceKey = &id
	}
	return nil
}

func (c *Cadvisor) Identifier(globals Globals) (string, error) {
	if c.Cmn.InstanceKey != nil {
		return *c.Cmn.InstanceKey, nil
	}
	return c.Config.InstanceKey(globals.AgentIdentifier)
}

func (c *Cadvisor) NewIntegration(logger log.Logger, globals Globals) (Integration, error) {
	return newIntegration(c, logger, globals, c.Config.NewIntegration)
}

type ConsulExporter struct {
	consul_exporter.Config `yaml:",omitempty,inline"`
	Cmn                    common.MetricsConfig `yaml:",inline"`
}

func (c *ConsulExporter) Cfg() Config {
	return c
}

func (c *ConsulExporter) Common() common.MetricsConfig {
	return c.Cmn
}

func (c *ConsulExporter) UnmarshalYAML(unmarshal func(interface{}) error) error {
	c.Config = consul_exporter.DefaultConfig
	type plain ConsulExporter
	return unmarshal((*plain)(c))
}

func (c *ConsulExporter) ApplyDefaults(globals Globals) error {
	c.Cmn.ApplyDefaults(globals.SubsystemOpts.Metrics.Autoscrape)
	if id, err := c.Identifier(globals); err == nil {
		c.Cmn.InstanceKey = &id
	}
	return nil
}

func (c *ConsulExporter) Identifier(globals Globals) (string, error) {
	if c.Cmn.InstanceKey != nil {
		return *c.Cmn.InstanceKey, nil
	}
	return c.Config.InstanceKey(globals.AgentIdentifier)
}

func (c *ConsulExporter) NewIntegration(logger log.Logger, globals Globals) (Integration, error) {
	return newIntegration(c, logger, globals, c.Config.NewIntegration)
}

type DnsmasqExporter struct {
	dnsmasq_exporter.Config `yaml:",omitempty,inline"`
	Cmn                     common.MetricsConfig `yaml:",inline"`
}

func (c *DnsmasqExporter) Cfg() Config {
	return c
}

func (c *DnsmasqExporter) Common() common.MetricsConfig {
	return c.Cmn
}

func (c *DnsmasqExporter) UnmarshalYAML(unmarshal func(interface{}) error) error {
	c.Config = dnsmasq_exporter.DefaultConfig
	type plain DnsmasqExporter
	return unmarshal((*plain)(c))
}

func (c *DnsmasqExporter) ApplyDefaults(globals Globals) error {
	c.Cmn.ApplyDefaults(globals.SubsystemOpts.Metrics.Autoscrape)
	if id, err := c.Identifier(globals); err == nil {
		c.Cmn.InstanceKey = &id
	}
	return nil
}

func (c *DnsmasqExporter) Identifier(globals Globals) (string, error) {
	if c.Cmn.InstanceKey != nil {
		return *c.Cmn.InstanceKey, nil
	}
	return c.Config.InstanceKey(globals.AgentIdentifier)
}

func (c *DnsmasqExporter) NewIntegration(logger log.Logger, globals Globals) (Integration, error) {
	return newIntegration(c, logger, globals, c.Config.NewIntegration)
}

type ElasticsearchExporter struct {
	elasticsearch_exporter.Config `yaml:",omitempty,inline"`
	Cmn                           common.MetricsConfig `yaml:",inline"`
}

func (c *ElasticsearchExporter) Cfg() Config {
	return c
}

func (c *ElasticsearchExporter) Common() common.MetricsConfig {
	return c.Cmn
}

func (c *ElasticsearchExporter) UnmarshalYAML(unmarshal func(interface{}) error) error {
	c.Config = elasticsearch_exporter.DefaultConfig
	type plain ElasticsearchExporter
	return unmarshal((*plain)(c))
}

func (c *ElasticsearchExporter) ApplyDefaults(globals Globals) error {
	c.Cmn.ApplyDefaults(globals.SubsystemOpts.Metrics.Autoscrape)
	if id, err := c.Identifier(globals); err == nil {
		c.Cmn.InstanceKey = &id
	}
	return nil
}

func (c *ElasticsearchExporter) Identifier(globals Globals) (string, error) {
	if c.Cmn.InstanceKey != nil {
		return *c.Cmn.InstanceKey, nil
	}
	return c.Config.InstanceKey(globals.AgentIdentifier)
}

func (c *ElasticsearchExporter) NewIntegration(logger log.Logger, globals Globals) (Integration, error) {
	return newIntegration(c, logger, globals, c.Config.NewIntegration)
}

type GithubExporter struct {
	github_exporter.Config `yaml:",omitempty,inline"`
	Cmn                    common.MetricsConfig `yaml:",inline"`
}

func (c *GithubExporter) Cfg() Config {
	return c
}

func (c *GithubExporter) Common() common.MetricsConfig {
	return c.Cmn
}

func (c *GithubExporter) UnmarshalYAML(unmarshal func(interface{}) error) error {
	c.Config = github_exporter.DefaultConfig
	type plain GithubExporter
	return unmarshal((*plain)(c))
}

func (c *GithubExporter) ApplyDefaults(globals Globals) error {
	c.Cmn.ApplyDefaults(globals.SubsystemOpts.Metrics.Autoscrape)
	if id, err := c.Identifier(globals); err == nil {
		c.Cmn.InstanceKey = &id
	}
	return nil
}

func (c *GithubExporter) Identifier(globals Globals) (string, error) {
	if c.Cmn.InstanceKey != nil {
		return *c.Cmn.InstanceKey, nil
	}
	return c.Config.InstanceKey(globals.AgentIdentifier)
}

func (c *GithubExporter) NewIntegration(logger log.Logger, globals Globals) (Integration, error) {
	return newIntegration(c, logger, globals, c.Config.NewIntegration)
}

type KafkaExporter struct {
	kafka_exporter.Config `yaml:",omitempty,inline"`
	Cmn                   common.MetricsConfig `yaml:",inline"`
}

func (c *KafkaExporter) Cfg() Config {
	return c
}

func (c *KafkaExporter) Common() common.MetricsConfig {
	return c.Cmn
}

func (c *KafkaExporter) UnmarshalYAML(unmarshal func(interface{}) error) error {
	c.Config = kafka_exporter.DefaultConfig
	type plain KafkaExporter
	return unmarshal((*plain)(c))
}

func (c *KafkaExporter) ApplyDefaults(globals Globals) error {
	c.Cmn.ApplyDefaults(globals.SubsystemOpts.Metrics.Autoscrape)
	if id, err := c.Identifier(globals); err == nil {
		c.Cmn.InstanceKey = &id
	}
	return nil
}

func (c *KafkaExporter) Identifier(globals Globals) (string, error) {
	if c.Cmn.InstanceKey != nil {
		return *c.Cmn.InstanceKey, nil
	}
	return c.Config.InstanceKey(globals.AgentIdentifier)
}

func (c *KafkaExporter) NewIntegration(logger log.Logger, globals Globals) (Integration, error) {
	return newIntegration(c, logger, globals, c.Config.NewIntegration)
}

type MemcachedExporter struct {
	memcached_exporter.Config `yaml:",omitempty,inline"`
	Cmn                       common.MetricsConfig `yaml:",inline"`
}

func (c *MemcachedExporter) Cfg() Config {
	return c
}

func (c *MemcachedExporter) Common() common.MetricsConfig {
	return c.Cmn
}

func (c *MemcachedExporter) UnmarshalYAML(unmarshal func(interface{}) error) error {
	c.Config = memcached_exporter.DefaultConfig
	type plain MemcachedExporter
	return unmarshal((*plain)(c))
}

func (c *MemcachedExporter) ApplyDefaults(globals Globals) error {
	c.Cmn.ApplyDefaults(globals.SubsystemOpts.Metrics.Autoscrape)
	if id, err := c.Identifier(globals); err == nil {
		c.Cmn.InstanceKey = &id
	}
	return nil
}

func (c *MemcachedExporter) Identifier(globals Globals) (string, error) {
	if c.Cmn.InstanceKey != nil {
		return *c.Cmn.InstanceKey, nil
	}
	return c.Config.InstanceKey(globals.AgentIdentifier)
}

func (c *MemcachedExporter) NewIntegration(logger log.Logger, globals Globals) (Integration, error) {
	return newIntegration(c, logger, globals, c.Config.NewIntegration)
}

type MongodbExporter struct {
	mongodb_exporter.Config `yaml:",omitempty,inline"`
	Cmn                     common.MetricsConfig `yaml:",inline"`
}

func (c *MongodbExporter) Cfg() Config {
	return c
}

func (c *MongodbExporter) Common() common.MetricsConfig {
	return c.Cmn
}

func (c *MongodbExporter) ApplyDefaults(globals Globals) error {
	c.Cmn.ApplyDefaults(globals.SubsystemOpts.Metrics.Autoscrape)
	if id, err := c.Identifier(globals); err == nil {
		c.Cmn.InstanceKey = &id
	}
	return nil
}

func (c *MongodbExporter) Identifier(globals Globals) (string, error) {
	if c.Cmn.InstanceKey != nil {
		return *c.Cmn.InstanceKey, nil
	}
	return c.Config.InstanceKey(globals.AgentIdentifier)
}

func (c *MongodbExporter) NewIntegration(logger log.Logger, globals Globals) (Integration, error) {
	return newIntegration(c, logger, globals, c.Config.NewIntegration)
}

type MysqldExporter struct {
	mysqld_exporter.Config `yaml:",omitempty,inline"`
	Cmn                    common.MetricsConfig `yaml:",inline"`
}

func (c *MysqldExporter) Cfg() Config {
	return c
}

func (c *MysqldExporter) Common() common.MetricsConfig {
	return c.Cmn
}

func (c *MysqldExporter) UnmarshalYAML(unmarshal func(interface{}) error) error {
	c.Config = mysqld_exporter.DefaultConfig
	type plain MysqldExporter
	return unmarshal((*plain)(c))
}

func (c *MysqldExporter) ApplyDefaults(globals Globals) error {
	c.Cmn.ApplyDefaults(globals.SubsystemOpts.Metrics.Autoscrape)
	if id, err := c.Identifier(globals); err == nil {
		c.Cmn.InstanceKey = &id
	}
	return nil
}

func (c *MysqldExporter) Identifier(globals Globals) (string, error) {
	if c.Cmn.InstanceKey != nil {
		return *c.Cmn.InstanceKey, nil
	}
	return c.Config.InstanceKey(globals.AgentIdentifier)
}

func (c *MysqldExporter) NewIntegration(logger log.Logger, globals Globals) (Integration, error) {
	return newIntegration(c, logger, globals, c.Config.NewIntegration)
}

type NodeExporter struct {
	node_exporter.Config `yaml:",omitempty,inline"`
	Cmn                  common.MetricsConfig `yaml:",inline"`
}

func (c *NodeExporter) Cfg() Config {
	return c
}

func (c *NodeExporter) Common() common.MetricsConfig {
	return c.Cmn
}

func (c *NodeExporter) UnmarshalYAML(unmarshal func(interface{}) error) error {
	c.Config = node_exporter.DefaultConfig
	type plain NodeExporter
	return unmarshal((*plain)(c))
}

func (c *NodeExporter) ApplyDefaults(globals Globals) error {
	c.Cmn.ApplyDefaults(globals.SubsystemOpts.Metrics.Autoscrape)
	if id, err := c.Identifier(globals); err == nil {
		c.Cmn.InstanceKey = &id
	}
	return nil
}

func (c *NodeExporter) Identifier(globals Globals) (string, error) {
	if c.Cmn.InstanceKey != nil {
		return *c.Cmn.InstanceKey, nil
	}
	return c.Config.InstanceKey(globals.AgentIdentifier)
}

func (c *NodeExporter) NewIntegration(logger log.Logger, globals Globals) (Integration, error) {
	return newIntegration(c, logger, globals, c.Config.NewIntegration)
}

type PostgresExporter struct {
	postgres_exporter.Config `yaml:",omitempty,inline"`
	Cmn                      common.MetricsConfig `yaml:",inline"`
}

func (c *PostgresExporter) Cfg() Config {
	return c
}

func (c *PostgresExporter) Common() common.MetricsConfig {
	return c.Cmn
}

func (c *PostgresExporter) ApplyDefaults(globals Globals) error {
	c.Cmn.ApplyDefaults(globals.SubsystemOpts.Metrics.Autoscrape)
	if id, err := c.Identifier(globals); err == nil {
		c.Cmn.InstanceKey = &id
	}
	return nil
}

func (c *PostgresExporter) Identifier(globals Globals) (string, error) {
	if c.Cmn.InstanceKey != nil {
		return *c.Cmn.InstanceKey, nil
	}
	return c.Config.InstanceKey(globals.AgentIdentifier)
}

func (c *PostgresExporter) NewIntegration(logger log.Logger, globals Globals) (Integration, error) {
	return newIntegration(c, logger, globals, c.Config.NewIntegration)
}

type ProcessExporter struct {
	process_exporter.Config `yaml:",omitempty,inline"`
	Cmn                     common.MetricsConfig `yaml:",inline"`
}

func (c *ProcessExporter) Cfg() Config {
	return c
}

func (c *ProcessExporter) Common() common.MetricsConfig {
	return c.Cmn
}

func (c *ProcessExporter) UnmarshalYAML(unmarshal func(interface{}) error) error {
	c.Config = process_exporter.DefaultConfig
	type plain ProcessExporter
	return unmarshal((*plain)(c))
}

func (c *ProcessExporter) ApplyDefaults(globals Globals) error {
	c.Cmn.ApplyDefaults(globals.SubsystemOpts.Metrics.Autoscrape)
	if id, err := c.Identifier(globals); err == nil {
		c.Cmn.InstanceKey = &id
	}
	return nil
}

func (c *ProcessExporter) Identifier(globals Globals) (string, error) {
	if c.Cmn.InstanceKey != nil {
		return *c.Cmn.InstanceKey, nil
	}
	return c.Config.InstanceKey(globals.AgentIdentifier)
}

func (c *ProcessExporter) NewIntegration(logger log.Logger, globals Globals) (Integration, error) {
	return newIntegration(c, logger, globals, c.Config.NewIntegration)
}

type RedisExporter struct {
	redis_exporter.Config `yaml:",omitempty,inline"`
	Cmn                   common.MetricsConfig `yaml:",inline"`
}

func (c *RedisExporter) Cfg() Config {
	return c
}

func (c *RedisExporter) Common() common.MetricsConfig {
	return c.Cmn
}

func (c *RedisExporter) UnmarshalYAML(unmarshal func(interface{}) error) error {
	c.Config = redis_exporter.DefaultConfig
	type plain RedisExporter
	return unmarshal((*plain)(c))
}

func (c *RedisExporter) ApplyDefaults(globals Globals) error {
	c.Cmn.ApplyDefaults(globals.SubsystemOpts.Metrics.Autoscrape)
	if id, err := c.Identifier(globals); err == nil {
		c.Cmn.InstanceKey = &id
	}
	return nil
}

func (c *RedisExporter) Identifier(globals Globals) (string, error) {
	if c.Cmn.InstanceKey != nil {
		return *c.Cmn.InstanceKey, nil
	}
	return c.Config.InstanceKey(globals.AgentIdentifier)
}

func (c *RedisExporter) NewIntegration(logger log.Logger, globals Globals) (Integration, error) {
	return newIntegration(c, logger, globals, c.Config.NewIntegration)
}

type StatsdExporter struct {
	statsd_exporter.Config `yaml:",omitempty,inline"`
	Cmn                    common.MetricsConfig `yaml:",inline"`
}

func (c *StatsdExporter) Cfg() Config {
	return c
}

func (c *StatsdExporter) Common() common.MetricsConfig {
	return c.Cmn
}

func (c *StatsdExporter) UnmarshalYAML(unmarshal func(interface{}) error) error {
	c.Config = statsd_exporter.DefaultConfig
	type plain StatsdExporter
	return unmarshal((*plain)(c))
}

func (c *StatsdExporter) ApplyDefaults(globals Globals) error {
	c.Cmn.ApplyDefaults(globals.SubsystemOpts.Metrics.Autoscrape)
	if id, err := c.Identifier(globals); err == nil {
		c.Cmn.InstanceKey = &id
	}
	return nil
}

func (c *StatsdExporter) Identifier(globals Globals) (string, error) {
	if c.Cmn.InstanceKey != nil {
		return *c.Cmn.InstanceKey, nil
	}
	return c.Config.InstanceKey(globals.AgentIdentifier)
}

func (c *StatsdExporter) NewIntegration(logger log.Logger, globals Globals) (Integration, error) {
	return newIntegration(c, logger, globals, c.Config.NewIntegration)
}

type WindowsExporter struct {
	windows_exporter.Config `yaml:",omitempty,inline"`
	Cmn                     common.MetricsConfig `yaml:",inline"`
}

func (c *WindowsExporter) Cfg() Config {
	return c
}

func (c *WindowsExporter) Common() common.MetricsConfig {
	return c.Cmn
}

func (c *WindowsExporter) UnmarshalYAML(unmarshal func(interface{}) error) error {
	c.Config = windows_exporter.DefaultConfig
	type plain WindowsExporter
	return unmarshal((*plain)(c))
}

func (c *WindowsExporter) ApplyDefaults(globals Globals) error {
	c.Cmn.ApplyDefaults(globals.SubsystemOpts.Metrics.Autoscrape)
	if id, err := c.Identifier(globals); err == nil {
		c.Cmn.InstanceKey = &id
	}
	return nil
}

func (c *WindowsExporter) Identifier(globals Globals) (string, error) {
	if c.Cmn.InstanceKey != nil {
		return *c.Cmn.InstanceKey, nil
	}
	return c.Config.InstanceKey(globals.AgentIdentifier)
}

func (c *WindowsExporter) NewIntegration(logger log.Logger, globals Globals) (Integration, error) {
	return newIntegration(c, logger, globals, c.Config.NewIntegration)
}

func newIntegration(c IntegrationConfig, logger log.Logger, globals Globals, newInt func(l log.Logger) (shared.Integration, error)) (Integration, error) {

	v1Integration, err := newInt(logger)
	if err != nil {
		return nil, err
	}

	id, err := c.Cfg().Identifier(globals)
	if err != nil {
		return nil, err
	}

	// Generate our handler. Original integrations didn't accept a prefix, and
	// just assumed that they would be wired to /metrics somewhere.
	handler, err := v1Integration.MetricsHandler()
	if err != nil {
		return nil, fmt.Errorf("generating http handler: %w", err)
	} else if handler == nil {
		handler = http.NotFoundHandler()
	}

	// Generate targets. Original integrations used a static set of targets,
	// so this mapping can always be generated just once.
	//
	// Targets are generated from the result of ScrapeConfigs(), which returns a
	// tuple of job name and relative metrics path.
	//
	// Job names were prefixed at the subsystem level with integrations/, so we
	// will retain that behavior here.
	v1ScrapeConfigs := v1Integration.ScrapeConfigs()
	targets := make([]handlerTarget, 0, len(v1ScrapeConfigs))
	for _, sc := range v1ScrapeConfigs {
		targets = append(targets, handlerTarget{
			MetricsPath: sc.MetricsPath,
			Labels: model.LabelSet{
				model.JobLabel: model.LabelValue("integrations/" + sc.JobName),
			},
		})
	}

	// Convert he run function. Original integrations sometimes returned
	// ctx.Err() on exit. This isn't recommended anymore, but we need to hide the
	// error if it happens, since the error was previously ignored.
	runFunc := func(ctx context.Context) error {
		err := v1Integration.Run(ctx)
		switch {
		case err == nil:
			return nil
		case errors.Is(err, context.Canceled) && ctx.Err() != nil:
			// Hide error that no longer happens in newer integrations.
			return nil
		default:
			return err
		}
	}

	// Aggregate our converted settings into a v2 integration.
	return &metricsHandlerIntegration{
		integrationName: c.Cfg().Name(),
		instanceID:      id,

		common:  c.Common(),
		globals: globals,
		handler: handler,
		targets: targets,

		runFunc: runFunc,
	}, nil
}
