package flowmode

import (
	"fmt"
	stdlog "log"
	"net"
	"os"
	"strconv"
	"time"

	"github.com/go-kit/log"
	"github.com/go-kit/log/level"
	"github.com/grafana/agent/service/cluster"
	"github.com/grafana/ckit/advertise"
	"github.com/hashicorp/go-discover"
	"github.com/hashicorp/go-discover/provider/k8s"
	"github.com/prometheus/client_golang/prometheus"
	"go.opentelemetry.io/otel/trace"
)

type clusterOptions struct {
	Log     log.Logger
	Metrics prometheus.Registerer
	Tracer  trace.TracerProvider

	EnableClustering    bool
	NodeName            string
	AdvertiseAddress    string
	ListenAddress       string
	JoinPeers           []string
	DiscoverPeers       string
	RejoinInterval      time.Duration
	AdvertiseInterfaces []string
	ClusterMaxJoinPeers int
	ClusterName         string
}

func buildClusterService(opts clusterOptions) (*cluster.Service, error) {
	listenPort := findPort(opts.ListenAddress, 80)

	config := cluster.Options{
		Log:     opts.Log,
		Metrics: opts.Metrics,
		Tracer:  opts.Tracer,

		EnableClustering:    opts.EnableClustering,
		NodeName:            opts.NodeName,
		AdvertiseAddress:    opts.AdvertiseAddress,
		RejoinInterval:      opts.RejoinInterval,
		ClusterMaxJoinPeers: opts.ClusterMaxJoinPeers,
		ClusterName:         opts.ClusterName,
	}

	if config.NodeName == "" {
		hostname, err := os.Hostname()
		if err != nil {
			return nil, fmt.Errorf("generating node name: %w", err)
		}
		config.NodeName = hostname
	}

	if config.AdvertiseAddress == "" {
		advertiseAddress := fmt.Sprintf("%s:%d", net.ParseIP("127.0.0.1"), listenPort)
		if opts.EnableClustering {
			advertiseInterfaces := opts.AdvertiseInterfaces
			if useAllInterfaces(advertiseInterfaces) {
				advertiseInterfaces = nil
			}
			addr, err := advertise.FirstAddress(advertiseInterfaces)
			if err != nil {
				level.Warn(opts.Log).Log("msg", "could not find advertise address using network interfaces", opts.AdvertiseInterfaces,
					"falling back to localhost", "err", err)
			} else if addr.Is4() {
				advertiseAddress = fmt.Sprintf("%s:%d", addr.String(), listenPort)
			} else if addr.Is6() {
				advertiseAddress = fmt.Sprintf("[%s]:%d", addr.String(), listenPort)
			} else {
				return nil, fmt.Errorf("type unknown for address: %s", addr.String())
			}
		}
		config.AdvertiseAddress = advertiseAddress
	} else {
		config.AdvertiseAddress = appendDefaultPort(config.AdvertiseAddress, listenPort)
	}

	switch {
	case len(opts.JoinPeers) > 0 && opts.DiscoverPeers != "":
		return nil, fmt.Errorf("at most one of join peers and discover peers may be set")

	case len(opts.JoinPeers) > 0:
		config.DiscoverPeers = newStaticDiscovery(opts.JoinPeers, listenPort)

	case opts.DiscoverPeers != "":
		discoverFunc, err := newDynamicDiscovery(config.Log, opts.DiscoverPeers, listenPort)
		if err != nil {
			return nil, err
		}
		config.DiscoverPeers = discoverFunc

	default:
		// Here, both JoinPeers and DiscoverPeers are empty. This is desirable when
		// starting a seed node that other nodes connect to, so we don't require
		// one of the fields to be set.
	}

	return cluster.New(config)
}

func useAllInterfaces(interfaces []string) bool {
	return len(interfaces) == 1 && interfaces[0] == "all"
}

func findPort(addr string, defaultPort int) int {
	_, portStr, err := net.SplitHostPort(addr)
	if err != nil {
		return defaultPort
	}
	port, err := strconv.Atoi(portStr)
	if err != nil {
		return defaultPort
	}
	return port
}

func appendDefaultPort(addr string, port int) string {
	_, _, err := net.SplitHostPort(addr)
	if err == nil {
		// No error means there was a port in the string
		return addr
	}
	return fmt.Sprintf("%s:%d", addr, port)
}

type discoverFunc func() ([]string, error)

func newStaticDiscovery(peers []string, defaultPort int) discoverFunc {
	return func() ([]string, error) {
		var addrs []string

		for _, addr := range peers {
			addrs = appendJoinAddr(addrs, addr)
		}

		for i := range addrs {
			// Default to using the same advertise port as the local node. This may
			// break in some cases, so the user should make sure the port numbers
			// align on as many nodes as possible.
			addrs[i] = appendDefaultPort(addrs[i], defaultPort)
		}

		return addrs, nil
	}
}

func appendJoinAddr(addrs []string, in string) []string {
	_, _, err := net.SplitHostPort(in)
	if err == nil {
		addrs = append(addrs, in)
		return addrs
	}

	ip := net.ParseIP(in)
	if ip != nil {
		addrs = append(addrs, ip.String())
		return addrs
	}

	_, srvs, err := net.LookupSRV("", "", in)
	if err == nil {
		for _, srv := range srvs {
			addrs = append(addrs, srv.Target)
		}
	}

	return addrs
}

func newDynamicDiscovery(l log.Logger, config string, defaultPort int) (discoverFunc, error) {
	providers := make(map[string]discover.Provider, len(discover.Providers)+1)
	for k, v := range discover.Providers {
		providers[k] = v
	}

	// Custom providers that aren't enabled by default
	providers["k8s"] = &k8s.Provider{}

	discoverer, err := discover.New(discover.WithProviders(providers))
	if err != nil {
		return nil, fmt.Errorf("bootstrapping peer discovery: %w", err)
	}

	return func() ([]string, error) {
		addrs, err := discoverer.Addrs(config, stdlog.New(log.NewStdlibAdapter(l), "", 0))
		if err != nil {
			return nil, fmt.Errorf("discovering peers: %w", err)
		}

		for i := range addrs {
			// Default to using the same advertise port as the local node. This may
			// break in some cases, so the user should make sure the port numbers
			// align on as many nodes as possible.
			addrs[i] = appendDefaultPort(addrs[i], defaultPort)
		}

		return addrs, nil
	}, nil
}
