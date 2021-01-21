// Package elasticsearch_exporter instantiates the exporter from github.com/justwatchcom/elasticsearch_exporter
// Using the YAML config provided by the agent
package elasticsearch_exporter //nolint:golint

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
	"time"

	"github.com/go-kit/kit/log"
	"github.com/go-kit/kit/log/level"
	"github.com/grafana/agent/pkg/integrations"
	"github.com/grafana/agent/pkg/integrations/config"
	"github.com/prometheus/client_golang/prometheus"

	"github.com/justwatchcom/elasticsearch_exporter/collector"
	"github.com/justwatchcom/elasticsearch_exporter/pkg/clusterinfo"
)

var DefaultConfig = Config{
	URI:                       "http://localhost:9200",
	Timeout:                   5 * time.Second,
	Node:                      "_local",
	ExportClusterInfoInterval: 5 * time.Minute,
}

// Config controls the elasticsearch_exporter integration.
type Config struct {
	Common config.Common `yaml:",inline"`

	// YAML-ized exporter flags.
	// YAML keys correspond to the flags in the exporter binary.

	// HTTP API address of an Elasticsearch node.
	URI string `yaml:"es.uri"`
	// Timeout for trying to get stats from Elasticsearch.
	Timeout time.Duration `yaml:"es.timeout"`
	// Export stats for all nodes in the cluster. If used, this flag will override the flag es.node.
	AllNodes bool `yaml:"es.all"`
	// Node's name of which metrics should be exposed.
	Node string `yaml:"es.node"`
	// Export stats for indices in the cluster.
	ExportIndices bool `yaml:"es.indices"`
	// Export stats for settings of all indices of the cluster.
	ExportIndicesSettings bool `yaml:"es.indices_settings"`
	// Export stats for cluster settings.
	ExportClusterSettings bool `yaml:"es.cluster_settings"`
	// Export stats for shards in the cluster (implies es.indices).
	ExportShards bool `yaml:"es.shards"`
	// Export stats for the cluster snapshots.
	ExportSnapshots bool `yaml:"es.snapshots"`
	// Cluster info update interval for the cluster label.
	ExportClusterInfoInterval time.Duration `yaml:"es.clusterinfo.interval"`
	// Path to PEM file that contains trusted Certificate Authorities for the Elasticsearch connection.
	CA string `yaml:"es.ca"`
	// Path to PEM file that contains the private key for client auth when connecting to Elasticsearch.
	ClientPrivateKey string `yaml:"es.client-private-key"`
	// Path to PEM file that contains the corresponding cert for the private key to connect to Elasticsearch.
	ClientCert string `yaml:"es.client-cert"`
	// Skip SSL verification when connecting to Elasticsearch.
	InsecureSkipVerify bool `yaml:"es.ssl-skip-verify"`
}

// UnmarshalYAML implements yaml.Unmarshaler for Config
func (c *Config) UnmarshalYAML(unmarshal func(interface{}) error) error {
	*c = DefaultConfig

	type plain Config
	return unmarshal((*plain)(c))
}

func (c Config) Name() string {
	return "elasticsearch_exporter"
}

func (c Config) CommonConfig() config.Common {
	return c.Common
}

// NewIntegration creates the elasticsearch integration by replicating the main() function of github.com/justwatchcom/elasticsearch_exporter
// but using yaml configuration instead of kingpin flags.
func (c Config) NewIntegration(logger log.Logger) (integrations.Integration, error) {
	if c.URI == "" {
		return nil, fmt.Errorf("empty URI provided")
	}
	esURL, err := url.Parse(c.URI)
	if err != nil {
		return nil, fmt.Errorf("failed to parse es.uri: %w", err)
	}

	// returns nil if not provided and falls back to simple TCP.
	tlsConfig := createTLSConfig(c.CA, c.ClientCert, c.ClientPrivateKey, c.InsecureSkipVerify)

	httpClient := &http.Client{
		Timeout: c.Timeout,
		Transport: &http.Transport{
			TLSClientConfig: tlsConfig,
			Proxy:           http.ProxyFromEnvironment,
		},
	}

	clusterInfoRetriever := clusterinfo.New(logger, httpClient, esURL, c.ExportClusterInfoInterval)

	collectors := []prometheus.Collector{
		clusterInfoRetriever,
		collector.NewClusterHealth(logger, httpClient, esURL),
		collector.NewNodes(logger, httpClient, esURL, c.AllNodes, c.Node),
	}

	if c.ExportIndices || c.ExportShards {
		iC := collector.NewIndices(logger, httpClient, esURL, c.ExportShards)
		collectors = append(collectors, iC)
		if registerErr := clusterInfoRetriever.RegisterConsumer(iC); registerErr != nil {
			return nil, fmt.Errorf("failed to register indices collector in cluster info: %w", err)
		}
	}

	if c.ExportShards {
		collectors = append(collectors, collector.NewSnapshots(logger, httpClient, esURL))
	}

	if c.ExportClusterSettings {
		collectors = append(collectors, collector.NewClusterSettings(logger, httpClient, esURL))
	}

	if c.ExportIndicesSettings {
		collectors = append(collectors, collector.NewIndicesSettings(logger, httpClient, esURL))
	}

	start := func(ctx context.Context) error {
		// start the cluster info retriever
		switch runErr := clusterInfoRetriever.Run(ctx); runErr {
		case nil:
			_ = level.Info(logger).Log(
				"msg", "started cluster info retriever",
				"interval", c.ExportClusterInfoInterval.String(),
			)
		case clusterinfo.ErrInitialCallTimeout:
			_ = level.Info(logger).Log("msg", "initial cluster info call timed out")
		default:
			_ = level.Error(logger).Log("msg", "failed to run cluster info retriever", "err", err)
			return err
		}

		// Wait until we're done
		<-ctx.Done()
		return ctx.Err()
	}

	return integrations.NewCollectorIntegration(c.Name(),
		integrations.WithCollectors(collectors...),
		integrations.WithRunner(start),
	), nil
}

func init() {
	integrations.RegisterIntegration(&Config{})
}
