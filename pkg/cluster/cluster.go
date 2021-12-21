// Package cluster exposes a clustering subsystem for Grafana Agent.
//
// A cluster is formed by agents that hold awareness of each other. Clustered
// agents are able to invoke gRPC endpoints against other agents. This may be
// used to build distributed capabilities onto other subsystems or components.
package cluster

import (
	"flag"
	"fmt"
	"io"
	stdlog "log"
	"net"
	"os"
	"sync"

	"github.com/go-kit/log"
	"github.com/go-kit/log/level"
	"github.com/grafana/dskit/flagext"
	"github.com/hashicorp/go-discover"
	"github.com/hashicorp/go-discover/provider/k8s"
	"github.com/rfratto/ckit"
	"github.com/rfratto/ckit/chash"
)

// ErrClusterDisabled is returned when performing a cluster operation when the
// cluster is currently disabled.
var ErrClusterDisabled = fmt.Errorf("cluster disabled")

// Config controls the clustering of Agents.
type Config struct {
	// If true, clustering will be used.
	Enable bool

	Discoverer          ckit.DiscovererConfig
	AdvertiseInterfaces flagext.StringSlice
	JoinPeers           flagext.StringSlice
	DiscoverPeers       string
}

// DefaultConfig holds default options for clustering.
var DefaultConfig = Config{
	Discoverer: ckit.DiscovererConfig{
		ListenPort: 7935,
	},
	AdvertiseInterfaces: []string{"eth0", "en0"},
}

// RegisterFlags registers flags to the provided flagset.
func (c *Config) RegisterFlags(fs *flag.FlagSet) {
	*c = DefaultConfig

	fs.BoolVar(&c.Enable, "cluster.enable", DefaultConfig.Enable, "Enables clustering.")

	fs.StringVar(&c.Discoverer.Name, "cluster.node-name", DefaultConfig.Discoverer.Name, "Name to identify node in cluster. If empty, defaults to hostname.")
	fs.StringVar(&c.Discoverer.ListenAddr, "cluster.listen-addr", DefaultConfig.Discoverer.ListenAddr, "IP address to listen for gossip traffic on. If not set, defaults to the first IP found from cluster.advertise-interfaces.")
	fs.StringVar(&c.Discoverer.AdvertiseAddr, "cluster.advertise-addr", DefaultConfig.Discoverer.AdvertiseAddr, "IP address to advertise to peers. If not set, defaults to the first IP found from cluster.advertise-interfaces.")
	fs.IntVar(&c.Discoverer.ListenPort, "cluster.listen-port", DefaultConfig.Discoverer.ListenPort, "Port to listen for TCP and UDP gossip traffic on.")

	fs.Var(&c.JoinPeers, "cluster.join-peers", "List of peers to join when starting up. Peers must have port number to connect to. Mutally exclusive with cluster.discover-peers.")
	fs.StringVar(&c.DiscoverPeers, "cluster.discover-peers", "", "Discover peers when starting up using Hashicorp cloud discover. If discovered peers do not have a port number, cluster.listen-port will be appended to each peer. Mutually exclusive with cluster.join-peers.")
}

// ApplyDefaults will validate and apply defaults to the config.
func (c *Config) ApplyDefaults(grpcPort int) error {
	if !c.Enable {
		return nil
	}

	if c.Discoverer.Name == "" {
		hn, err := os.Hostname()
		if err != nil {
			return fmt.Errorf("failed to generate node name: %w", err)
		}
		c.Discoverer.Name = hn
	}

	if c.Discoverer.ListenAddr == "" || c.Discoverer.AdvertiseAddr == "" {
		addr, err := getInstanceAddr("", c.AdvertiseInterfaces, log.NewNopLogger())
		if err != nil {
			return fmt.Errorf("failed to get advertise and listen address: %w", err)
		}
		if c.Discoverer.ListenAddr == "" {
			c.Discoverer.ListenAddr = addr
		}
		if c.Discoverer.AdvertiseAddr == "" {
			c.Discoverer.AdvertiseAddr = addr
		}
	}

	if len(c.JoinPeers) > 0 && c.DiscoverPeers != "" {
		return fmt.Errorf("only one of cluster.join-peers and cluster.discover-peers may be set")
	} else if c.DiscoverPeers != "" {
		providers := make(map[string]discover.Provider)
		for k, v := range discover.Providers {
			providers[k] = v
		}
		providers["k8s"] = &k8s.Provider{}

		d, err := discover.New(discover.WithProviders(providers))
		if err != nil {
			return fmt.Errorf("failed to bootstrap peer discovery: %w", err)
		}

		addrs, err := d.Addrs(c.DiscoverPeers, stdlog.New(io.Discard, "", 0))
		if err != nil {
			return fmt.Errorf("failed to discover peers to join: %w", err)
		}
		for _, addr := range addrs {
			c.JoinPeers = append(c.JoinPeers, appendDefaultPort(addr, c.Discoverer.ListenPort))
		}
	}

	// TODO(rfratto): this may not work if the listen address for gRPC is
	// different. Come back to this later.
	c.Discoverer.ApplicationAddr = fmt.Sprintf("%s:%d", c.Discoverer.AdvertiseAddr, grpcPort)
	return nil
}

func appendDefaultPort(addr string, port int) string {
	_, _, err := net.SplitHostPort(addr)
	if err == nil {
		// No error means there was a port in the string
		return addr
	}
	return fmt.Sprintf("%s:%d", addr, port)
}

// Node is a node within a cluster.
type Node struct {
	cfg  *Config
	disc *ckit.Discoverer
	node *ckit.Node
	log  log.Logger

	peersMut sync.RWMutex
	peers    ckit.PeerSet

	watcherMut sync.Mutex
	watchers   []PeersChangedWatcher
}

// PeersChangedWatcher is a function that will be invoked whenever
// peers change in the cluster. Returning true means the watcher
// should be invoked again next time peers change. Return false
// for one-shot changes. Do not modify the values in ps.
type PeersChangedWatcher func(ps ckit.PeerSet) (reregister bool)

// NewNode creates a new Node.
func NewNode(l log.Logger, c *Config) *Node {
	if l == nil {
		l = log.NewNopLogger()
	}

	c.Discoverer.Log = log.With(l, "component", "discoverer")

	l = log.With(l, "component", "cluster")
	n := &Node{cfg: c, log: l}

	if !c.Enable {
		return n
	}

	n.node = ckit.NewNode(chash.Ring(256), n.handlePeersChanged)
	return n
}

func (n *Node) handlePeersChanged(ps ckit.PeerSet) {
	level.Debug(n.log).Log("msg", "cluster peers changed", "peers", ps)

	n.peersMut.Lock()
	n.peers = ps
	n.peersMut.Unlock()

	n.watcherMut.Lock()
	defer n.watcherMut.Unlock()

	newWatchers := make([]PeersChangedWatcher, 0, len(n.watchers))
	for _, w := range n.watchers {
		rereg := w(ps)
		if rereg {
			newWatchers = append(newWatchers, w)
		}
	}
	n.watchers = newWatchers
}

// OnPeersChanged registers a watcher to be invoked every time the set of peers
// changes. w should return true as long as it should continue to get invoked.
// If w returns false, it will never be called again.
func (n *Node) OnPeersChanged(w PeersChangedWatcher) {
	n.watcherMut.Lock()
	defer n.watcherMut.Unlock()

	n.watchers = append(n.watchers, w)
}

// Start starts the node. If it was configured to join any peers, it will join
// them now. This will also invoke any PeersChangedWatchers to inform them that
// peers are joining.
//
// Start may not be called concurrently.
func (n *Node) Start() error {
	if !n.cfg.Enable {
		return nil
	}

	// The discoverer MUST be initialized in the Start method and not NewNode.
	// NewDiscoverer can immediately register the local node as a peer, which
	// would cause handlePeersChanged to be called before anything was registered.
	if n.disc == nil {
		var err error
		n.disc, err = ckit.NewDiscoverer(&n.cfg.Discoverer, n.node)
		if err != nil {
			return fmt.Errorf("failed to create node discoverer: %w", err)
		}
	}

	level.Info(n.log).Log("msg", "joining peers", "peers", n.cfg.JoinPeers)
	return n.disc.Start(n.cfg.JoinPeers)
}

// Peers returns the set of current peers. The resulting slice should not be
// modified.
func (n *Node) Peers() ckit.PeerSet {
	n.peersMut.RLock()
	defer n.peersMut.RUnlock()
	return n.peers
}

// Get retrieves the owner for a key.
func (n *Node) Get(key string) (*ckit.Peer, error) {
	if !n.cfg.Enable {
		return nil, ErrClusterDisabled
	}

	ps, err := n.node.Get(key, 1)
	if err != nil {
		return nil, fmt.Errorf("failed to get owner: %w", err)
	} else if len(ps) == 0 {
		return nil, fmt.Errorf("failed to get owner: not enough peers")
	}
	return &ps[0], nil
}

// Close closes the Node.
func (n *Node) Close() error {
	if !n.cfg.Enable {
		return nil
	}

	var firstErr error

	if n.disc != nil {
		level.Info(n.log).Log("msg", "leaving cluster")
		if err := n.disc.Close(); err != nil {
			level.Error(n.log).Log("msg", "failed to stop node discovery", "err", err)
			firstErr = err
		}
	}

	level.Info(n.log).Log("msg", "shutting down cluster node")
	err := n.node.Close()
	if err != nil {
		level.Error(n.log).Log("msg", "failed to stop node", "err", err)
	}
	if firstErr == nil && err != nil {
		firstErr = err
	}

	return firstErr
}
