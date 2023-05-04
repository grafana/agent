package vsphere

import (
	"time"
)

// vSphereConfig is the top level type for the vSphereConfig input plugin. It contains all the configuration
// and monitors a single vSphere endpoint
type vSphereConfig struct {
	Username            string
	Password            string
	DatacenterInstances bool
	ClusterInstances    bool
	HostInstances       bool
	VMInstances         bool
	DatastoreInstances  bool
	Separator           string
	UseIntSamples       bool
	IPAddresses         []string
	MetricLookback      int

	RefChunkSize            int
	MaxQueryObjects         int
	MaxQueryMetrics         int
	CollectConcurrency      int
	DiscoverConcurrency     int
	ForceDiscoverOnInit     bool
	ObjectDiscoveryInterval time.Duration
	Timeout                 time.Duration
	HistoricalInterval      time.Duration
}

var defaultVSphere = &vSphereConfig{
	DatacenterInstances: false,
	ClusterInstances:    false,
	HostInstances:       true,
	VMInstances:         true,
	DatastoreInstances:  false,
	Separator:           "_",
	UseIntSamples:       true,
	IPAddresses:         []string{},

	MaxQueryObjects:         256,
	MaxQueryMetrics:         256,
	CollectConcurrency:      1,
	DiscoverConcurrency:     1,
	MetricLookback:          3,
	ForceDiscoverOnInit:     true,
	ObjectDiscoveryInterval: time.Second * 300,
	Timeout:                 time.Second * 60,
	HistoricalInterval:      time.Second * 300,
}
