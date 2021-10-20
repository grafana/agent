package mongodb_exporter //nolint:golint

import (
	"context"
	"fmt"

	"github.com/gaantunes/mongodb_exporter/exporter"
	"github.com/go-kit/kit/log"
	"github.com/go-kit/kit/log/level"
	"github.com/grafana/agent/pkg/integrations"
	"github.com/grafana/agent/pkg/integrations/config"
	loki "github.com/grafana/agent/pkg/logs"
	"github.com/grafana/agent/pkg/tempo"
	"github.com/prometheus/client_golang/prometheus"
	"go.mongodb.org/mongo-driver/mongo"
)

// Exporter holds Exporter methods and attributes.
type Exporter struct {
	client       *mongo.Client
	topologyInfo exporter.LabelsGetter
	context      context.Context
	config       Config
}

// Config controls mongodb_exporter
type Config struct {
	Common config.Common `yaml:",inline"`

	// MongoDB connection URI. example:mongodb://user:pass@127.0.0.1:27017/admin?ssl=true"
	URI string `yaml:"mongodb_uri"`
}

// UnmarshalYAML implements yaml.Unmarshaler for Config
func (c *Config) UnmarshalYAML(unmarshal func(interface{}) error) error {

	type plain Config
	return unmarshal((*plain)(c))
}

// Name returns the name of the integration that this config represents.
func (c *Config) Name() string {
	return "mongodb_exporter"
}

// CommonConfig returns the common settings shared across all configs for
// integrations.
func (c *Config) CommonConfig() config.Common {
	return c.Common
}

// NewIntegration creates a new mongodb_exporter
func (c *Config) NewIntegration(logger log.Logger, loki *loki.Logs, tempo *tempo.Tempo) (integrations.Integration, error) {
	return New(logger, c)
}

func init() {
	integrations.RegisterIntegration(&Config{})
}

// New creates a new mongodb_exporter integration.
func New(logger log.Logger, c *Config) (integrations.Integration, error) {

	logrusLogger := NewLogger(logger)

	e := &Exporter{}
	e.config = *c

	context := context.Background()
	e.context = context

	var err error
	e.client, err = exporter.Connect(context, c.URI, true)
	if err != nil {
		return nil, err
	}

	level.Debug(logger).Log("initialized mongodb client")

	e.topologyInfo, err = exporter.NewTopologyInfo(context, e.client)
	if err != nil {
		return nil, err
	}

	nodeType, err := exporter.GetNodeType(context, e.client)
	if err != nil {
		return nil, fmt.Errorf("cannot get node type to check if this is a mongos: %s", err)
	}

	collectors := []prometheus.Collector{}

	gc := exporter.GeneralCollector{
		Ctx:    context,
		Client: e.client,
		Logger: logrusLogger,
	}

	collectors = append(collectors, &gc)

	var ddc = exporter.DiagnosticDataCollector{
		Ctx:            context,
		Client:         e.client,
		CompatibleMode: true,
		Logger:         logrusLogger,
		TopologyInfo:   e.topologyInfo,
	}

	collectors = append(collectors, &ddc)

	// replSetGetStatus is not supported through mongos
	if nodeType != exporter.TypeMongos {
		var rsgsc = exporter.ReplSetGetStatusCollector{
			Ctx:            context,
			Client:         e.client,
			CompatibleMode: true,
			Logger:         logrusLogger,
			TopologyInfo:   e.topologyInfo,
		}
		collectors = append(collectors, &rsgsc)
	}

	return integrations.NewCollectorIntegration(c.Name(), integrations.WithCollectors(collectors...)), nil

}
