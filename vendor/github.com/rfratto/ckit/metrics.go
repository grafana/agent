package ckit

import (
	"github.com/hashicorp/memberlist"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/rfratto/ckit/internal/metricsutil"
)

// Possible label values for metrics.gossipEventsTotal
const (
	eventStateChange      = "state_change_message"
	eventUnkownMessage    = "unknown_message"
	eventGetLocalState    = "get_local_state"
	eventMergeRemoteState = "merge_remote_state"
	eventNodeJoin         = "node_join"
	eventNodeLeave        = "node_leave"
	eventNodeUpdate       = "node_update"
	eventNodeConflict     = "node_conflict"
)

// metrics holds the set of metrics for a Node. Additional Collectors can be
// registered by calling Add.
type metrics struct {
	metricsutil.Container

	gossipEventsTotal  *prometheus.CounterVec
	nodePeers          *prometheus.GaugeVec
	nodeUpdating       prometheus.Gauge
	nodeUpdateDuration prometheus.Histogram
	nodeObservers      prometheus.Gauge
	nodeInfo           *metricsutil.InfoCollector
}

var _ prometheus.Collector = (*metrics)(nil)

func newMetrics() *metrics {
	var m metrics

	m.gossipEventsTotal = prometheus.NewCounterVec(prometheus.CounterOpts{
		Name: "cluster_node_gossip_received_events_total",
		Help: "Total number of gossip messages handled by the node.",
	}, []string{"event"})

	m.nodePeers = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Name: "cluster_node_peers",
		Help: "Current number of healthy peers by state",
	}, []string{"state"})

	m.nodeUpdating = prometheus.NewGauge(prometheus.GaugeOpts{
		Name: "cluster_node_updating",
		Help: "1 if the node is currently processing a change to the cluster state.",
	})

	m.nodeUpdateDuration = prometheus.NewHistogram(prometheus.HistogramOpts{
		Name:    "cluster_node_update_duration_seconds",
		Help:    "Histogram of the latency it took to process a change to the cluster state.",
		Buckets: prometheus.DefBuckets,
	})

	m.nodeObservers = prometheus.NewGauge(prometheus.GaugeOpts{
		Name: "cluster_node_update_observers",
		Help: "Number of internal observers waiting for changes to cluster state.",
	})

	m.nodeInfo = metricsutil.NewInfoCollector(metricsutil.InfoOpts{
		Name: "cluster_node_info",
		Help: "Info about the local node. Label values will change as the node changes state.",
	}, "state")

	m.Add(
		m.gossipEventsTotal,
		m.nodePeers,
		m.nodeUpdating,
		m.nodeUpdateDuration,
		m.nodeObservers,
		m.nodeInfo,
	)

	return &m
}

func newMemberlistCollector(ml *memberlist.Memberlist) prometheus.Collector {
	var container metricsutil.Container

	gossipProtoVersion := prometheus.NewGaugeFunc(prometheus.GaugeOpts{
		Name: "cluster_node_gossip_proto_version",
		Help: "Gossip protocol version used by nodes to maintain the cluster",
	}, func() float64 {
		// NOTE(rfratto): while this is static at the time of writing, the internal
		// documentation for memberlist claims that ProtocolVersion may one day be
		// updated at runtime.
		return float64(ml.ProtocolVersion())
	})

	gossipHealthScore := prometheus.NewGaugeFunc(prometheus.GaugeOpts{
		Name: "cluster_node_gossip_health_score",
		Help: "Health value of a node; lower values means healthier. 0 is the minimum.",
	}, func() float64 {
		return float64(ml.GetHealthScore())
	})

	gossipPeers := prometheus.NewGaugeFunc(prometheus.GaugeOpts{
		Name: "cluster_node_gossip_alive_peers",
		Help: "How many alive gossip peers a node has, including the local node.",
	}, func() float64 {
		return float64(ml.NumMembers())
	})

	container.Add(
		gossipProtoVersion,
		gossipHealthScore,
		gossipPeers,
	)

	return &container
}
