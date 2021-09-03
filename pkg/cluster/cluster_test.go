package cluster

import (
	"flag"
	"sync"
	"testing"
	"time"

	"github.com/go-kit/log"
	"github.com/go-kit/log/level"
	"github.com/grafana/agent/pkg/util"
	"github.com/rfratto/ckit"
	"github.com/stretchr/testify/require"
)

func TestCluster(t *testing.T) {
	opts := func(v ...string) []string {
		baseOptions := []string{
			"--cluster.listen-addr=127.0.0.1",
			"--cluster.advertise-addr=127.0.0.1",
		}
		return append(baseOptions, v...)
	}

	configs := []*Config{
		buildConfig(t, opts("--cluster.node-name=a", "--cluster.listen-port=7935")),
		buildConfig(t, opts("--cluster.node-name=b", "--cluster.listen-port=7936", "--cluster.join-peers=127.0.0.1:7935")),
		buildConfig(t, opts("--cluster.node-name=c", "--cluster.listen-port=7937", "--cluster.join-peers=127.0.0.1:7935")),
		buildConfig(t, opts("--cluster.node-name=d", "--cluster.listen-port=7938", "--cluster.join-peers=127.0.0.1:7935")),
		buildConfig(t, opts("--cluster.node-name=e", "--cluster.listen-port=7939", "--cluster.join-peers=127.0.0.1:7935")),
	}

	// Start up every node
	var (
		nodes []*Node

		peerMap sync.Map
	)
	for _, cfg := range configs {
		node := NewNode(cfg.Discoverer.Log, cfg)

		var (
			log  = cfg.Discoverer.Log
			name = cfg.Discoverer.Name
		)
		node.OnPeersChanged(func(ps ckit.PeerSet) (reregister bool) {
			level.Debug(log).Log("msg", "peers changed", "peers", ps)
			peerMap.Store(name, ps)
			return true
		})

		require.NoError(t, node.Start())
		nodes = append(nodes, node)

		t.Cleanup(func() {
			require.NoError(t, node.Close())
		})
	}

	// Wait for the dust to settle before testing anything.
	time.Sleep(1 * time.Second)

	// All nodes must have 5 peers.
	peerMap.Range(func(key, value interface{}) bool {
		require.Len(t, value.(ckit.PeerSet), 5, "node %s does not have the appropriate amount of peers", key.(string))
		return true
	})

	// Test a random key, ensuring that all nodes agree on an owner.
	key := "a random key"
	owner, err := nodes[0].Get(key)
	require.NoError(t, err)

	for _, node := range nodes[1:] {
		actual, err := node.Get(key)
		require.NoError(t, err)
		require.Equal(t, owner.Name, actual.Name)
	}
}

func buildConfig(t *testing.T, args []string) *Config {
	fs := flag.NewFlagSet(t.Name(), flag.PanicOnError)

	var cfg Config
	cfg.RegisterFlags(fs)
	require.NoError(t, fs.Parse(args))
	require.NoError(t, cfg.ApplyDefaults(8080))

	cfg.Discoverer.Log = log.With(util.TestLogger(t), "node", cfg.Discoverer.Name)
	return &cfg
}
