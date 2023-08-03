package cluster

import (
	"fmt"
	"io"
	stdlog "log"
	"math/rand"
	"net"
	"net/http"
	"os"
	"strings"
	"testing"

	"github.com/go-kit/log"
	"github.com/grafana/ckit/advertise"
	"github.com/hashicorp/go-discover"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/stretchr/testify/require"
)

// NOTE(rfratto): we don't test methods against GossipNode that just shim to
// ckit, since we can rely on the existing ckit tests for correctness.

const examplePort = 8888

var rnd = rand.New(rand.NewSource(1337))

func TestConfig_ApplyDefaults(t *testing.T) {
	ifaces, err := net.Interfaces()
	require.NoError(t, err)

	var advertiseInterfaces []string
	for _, iface := range ifaces {
		if iface.Flags != net.FlagLoopback {
			advertiseInterfaces = append(advertiseInterfaces, iface.Name)
		}
	}

	defaultConfig := DefaultGossipConfig
	defaultConfig.AdvertiseInterfaces = advertiseInterfaces
	defaultConfig.DefaultPort = examplePort

	setTestProviders(t, map[string]discover.Provider{
		"static": &staticProvider{},
	})

	hostName, err := os.Hostname()
	require.NoError(t, err, "failed to get hostname for test assertions")

	t.Run("node name defaults to hostname", func(t *testing.T) {
		gc := defaultConfig
		gc.NodeName = ""

		err := gc.ApplyDefaults()
		require.NoError(t, err)
		require.Equal(t, hostName, gc.NodeName)
	})

	t.Run("node name can be overridden", func(t *testing.T) {
		gc := defaultConfig
		gc.NodeName = "foobar"

		err := gc.ApplyDefaults()
		require.NoError(t, err)
		require.Equal(t, "foobar", gc.NodeName)
	})

	t.Run("one of advertise addr or advertise interfaces must be set", func(t *testing.T) {
		gc := defaultConfig
		gc.AdvertiseInterfaces = nil

		err := gc.ApplyDefaults()
		require.EqualError(t, err, "one of advertise address or advertise interfaces must be set")
	})

	t.Run("advertise address is inferred from advertise interfaces", func(t *testing.T) {
		gc := defaultConfig
		gc.AdvertiseInterfaces = advertiseInterfaces

		err := gc.ApplyDefaults()
		require.NoError(t, err)

		expect, err := advertise.FirstAddress(gc.AdvertiseInterfaces)
		require.NoError(t, err)
		require.Equal(t, fmt.Sprintf("%s:%d", expect, examplePort), gc.AdvertiseAddr)
	})

	t.Run("explicit advertise address can be set", func(t *testing.T) {
		gc := defaultConfig
		gc.AdvertiseAddr = "foobar:9999"

		err := gc.ApplyDefaults()
		require.NoError(t, err)
		require.Equal(t, "foobar:9999", gc.AdvertiseAddr)
	})

	t.Run("explicit advertise address can use default port", func(t *testing.T) {
		gc := defaultConfig
		gc.AdvertiseAddr = "foobar"

		err := gc.ApplyDefaults()
		require.NoError(t, err)
		require.Equal(t, fmt.Sprintf("foobar:%d", examplePort), gc.AdvertiseAddr)
	})

	t.Run("join peers and discover peers can't both be set", func(t *testing.T) {
		gc := defaultConfig
		gc.JoinPeers = []string{"foobar:9999"}
		gc.DiscoverPeers = `provider=static addrs=fizzbuzz:5555`

		err := gc.ApplyDefaults()
		require.EqualError(t, err, "at most one of join peers and discover peers may be set")
	})

	t.Run("explicit join peers can be set", func(t *testing.T) {
		gc := defaultConfig
		gc.JoinPeers = []string{"foobar:9999"}

		err := gc.ApplyDefaults()
		require.NoError(t, err)
		require.Equal(t, []string{"foobar:9999"}, []string(gc.JoinPeers))
	})

	t.Run("join peers can be discovered", func(t *testing.T) {
		gc := defaultConfig
		gc.DiscoverPeers = `provider=static addrs=fizzbuzz:5555`

		err := gc.ApplyDefaults()
		require.NoError(t, err)
		require.Equal(t, []string{"fizzbuzz:5555"}, getPeers(t, gc))
	})

	t.Run("peers can use default port", func(t *testing.T) {
		gc := defaultConfig
		gc.JoinPeers = []string{"192.168.1.14"}

		err := gc.ApplyDefaults()
		require.NoError(t, err)
		require.Equal(t, []string{fmt.Sprintf("192.168.1.14:%d", examplePort)}, getPeers(t, gc))
	})

	t.Run("discovered peers can use default port", func(t *testing.T) {
		gc := defaultConfig
		gc.DiscoverPeers = `provider=static addrs=fizzbuzz`

		err := gc.ApplyDefaults()
		require.NoError(t, err)
		require.Equal(t, []string{fmt.Sprintf("fizzbuzz:%d", examplePort)}, getPeers(t, gc))
	})
}

func setTestProviders(t *testing.T, set map[string]discover.Provider) {
	t.Helper()

	restore := extraDiscoverProviders
	t.Cleanup(func() {
		extraDiscoverProviders = restore
	})
	extraDiscoverProviders = set
}

type staticProvider struct{}

var _ discover.Provider = (*staticProvider)(nil)

func (sp *staticProvider) Addrs(args map[string]string, l *stdlog.Logger) ([]string, error) {
	if args["provider"] != "static" {
		return nil, fmt.Errorf("discover-static: invalid provider " + args["provider"])
	}
	if rawSet, ok := args["addrs"]; ok {
		return strings.Split(rawSet, ","), nil
	}
	return nil, nil
}

func (sp *staticProvider) Help() string {
	return `static:

    provider: "static"
		addrs:    Comma-separated list of addresses to return`
}

type randomProvider struct{}

var _ discover.Provider = (*randomProvider)(nil)

func (sp *randomProvider) Addrs(args map[string]string, l *stdlog.Logger) ([]string, error) {
	if args["provider"] != "random" {
		return nil, fmt.Errorf("discover-random: invalid provider " + args["provider"])
	}
	if rawSet, ok := args["addrs"]; ok {
		addrs := strings.Split(rawSet, ",")
		if len(addrs) == 0 {
			return nil, nil
		}

		return []string{addrs[rnd.Intn(len(addrs))]}, nil
	}
	return nil, nil
}

func (sp *randomProvider) Help() string {
	return `random:

    provider: "random"
		addrs:    Returns a random address from a comma-separated list of addresses`
}

func getPeers(t *testing.T, gc GossipConfig) []string {
	gc.NodeName = "gossip-node"
	node, err := NewGossipNode(log.NewLogfmtLogger(io.Discard), prometheus.NewRegistry(), &http.Client{}, &gc)
	require.NoError(t, err)
	peers, err := node.GetPeers()
	require.NoError(t, err)
	return peers
}

func TestGetPeers(t *testing.T) {
	ifaces, err := net.Interfaces()
	require.NoError(t, err)

	var advertiseInterfaces []string
	for _, iface := range ifaces {
		if iface.Flags != net.FlagLoopback {
			advertiseInterfaces = append(advertiseInterfaces, iface.Name)
		}
	}

	defaultConfig := DefaultGossipConfig
	defaultConfig.AdvertiseInterfaces = advertiseInterfaces
	defaultConfig.DefaultPort = examplePort

	setTestProviders(t, map[string]discover.Provider{
		"random": &randomProvider{},
	})

	t.Run("GetPeers refreshes the list from DiscoverPeers", func(t *testing.T) {
		gc := defaultConfig
		gc.DiscoverPeers = `provider=random addrs=one,two,three,four,five,six`

		err := gc.ApplyDefaults()
		require.NoError(t, err)
		require.Equal(t, []string{fmt.Sprintf("five:%d", examplePort)}, getPeers(t, gc))
		require.Equal(t, []string{fmt.Sprintf("three:%d", examplePort)}, getPeers(t, gc))
		require.Equal(t, []string{fmt.Sprintf("six:%d", examplePort)}, getPeers(t, gc))
		require.Equal(t, []string{fmt.Sprintf("six:%d", examplePort)}, getPeers(t, gc))
	})
}
