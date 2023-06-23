package elasticsearch

import (
	"testing"
	"time"

	"github.com/grafana/agent/component/discovery"
	"github.com/grafana/agent/pkg/integrations/elasticsearch_exporter"
	"github.com/grafana/agent/pkg/river"
	"github.com/stretchr/testify/require"
)

func TestRiverUnmarshal(t *testing.T) {
	riverConfig := `
	address              = "http://localhost:9300"
	timeout              = "10s"
	all                  = true
	node                 = "some_node"
	indices              = true
	indices_settings     = true
	cluster_settings     = true
	shards               = true
	aliases              = true
	snapshots            = true
	clusterinfo_interval = "10s"
	ca                   = "some_ca"
	client_private_key   = "some_client_cert"
	ssl_skip_verify      = true
	data_stream          = true
	slm                  = true
	`

	var args Arguments
	err := river.Unmarshal([]byte(riverConfig), &args)
	require.NoError(t, err)

	expected := Arguments{
		Address:                   "http://localhost:9300",
		Timeout:                   time.Duration(10) * time.Second,
		AllNodes:                  true,
		Node:                      "some_node",
		ExportIndices:             true,
		ExportIndicesSettings:     true,
		ExportClusterSettings:     true,
		ExportShards:              true,
		IncludeAliases:            true,
		ExportSnapshots:           true,
		ExportClusterInfoInterval: time.Duration(10) * time.Second,
		CA:                        "some_ca",
		ClientPrivateKey:          "some_client_cert",
		InsecureSkipVerify:        true,
		ExportDataStreams:         true,
		ExportSLM:                 true,
	}

	require.Equal(t, expected, args)
}

func TestConvert(t *testing.T) {
	riverConfig := `
	address              = "http://localhost:9300"
	timeout              = "10s"
	all                  = true
	node                 = "some_node"
	indices              = true
	indices_settings     = true
	cluster_settings     = true
	shards               = true
	aliases              = true
	snapshots            = true
	clusterinfo_interval = "10s"
	ca                   = "some_ca"
	client_private_key   = "some_client_cert"
	ssl_skip_verify      = true
	data_stream          = true
	slm                  = true
	`
	var args Arguments
	err := river.Unmarshal([]byte(riverConfig), &args)
	require.NoError(t, err)

	res := args.Convert()

	expected := elasticsearch_exporter.Config{
		Address:                   "http://localhost:9300",
		Timeout:                   time.Duration(10) * time.Second,
		AllNodes:                  true,
		Node:                      "some_node",
		ExportIndices:             true,
		ExportIndicesSettings:     true,
		ExportClusterSettings:     true,
		ExportShards:              true,
		IncludeAliases:            true,
		ExportSnapshots:           true,
		ExportClusterInfoInterval: time.Duration(10) * time.Second,
		CA:                        "some_ca",
		ClientPrivateKey:          "some_client_cert",
		InsecureSkipVerify:        true,
		ExportDataStreams:         true,
		ExportSLM:                 true,
	}
	require.Equal(t, expected, *res)
}

func TestCustomizeTarget(t *testing.T) {
	args := Arguments{
		Address: "http://localhost:9300",
	}

	baseTarget := discovery.Target{}
	newTargets := customizeTarget(baseTarget, args)
	require.Equal(t, 1, len(newTargets))
	require.Equal(t, "http://localhost:9300", newTargets[0]["instance"])
}
