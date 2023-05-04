package vsphere

import (
	"context"
	"net/url"
	"regexp"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/go-kit/log"
	"github.com/go-kit/log/level"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/vmware/govmomi/object"
	"github.com/vmware/govmomi/performance"
	"github.com/vmware/govmomi/vim25/mo"
	"github.com/vmware/govmomi/vim25/types"
)

var isIPv4 = regexp.MustCompile(`^(?:[0-9]{1,3}\.){3}[0-9]{1,3}$`)

var isIPv6 = regexp.MustCompile(`^(?:[A-Fa-f0-9]{0,4}:){1,7}[A-Fa-f0-9]{1,4}$`)

type endpoint struct {
	cfg              *vSphereConfig
	url              *url.URL
	resourceKinds    map[string]*resourceKind
	discoveryTicker  *time.Ticker
	collectMux       sync.RWMutex
	initialized      bool
	clientFactory    *clientFactory
	busy             sync.Mutex
	metricNameLookup map[int32]string
	metricNameMux    sync.RWMutex
	log              log.Logger

	// discovery meta monitoring
	dm *discoveryMetrics
}

type resourceKind struct {
	name             string
	vcName           string
	enabled          bool
	realTime         bool
	sampling         int32
	objects          objectMap
	paths            []string
	collectInstances bool
	getObjects       func(context.Context, *endpoint, *resourceFilter) (objectMap, error)
	metrics          performance.MetricList
	parent           string
	latestSample     time.Time
}

type objectMap map[string]*objectRef

type objectRef struct {
	name      string
	altID     string
	ref       types.ManagedObjectReference
	parentRef *types.ManagedObjectReference //Pointer because it must be nillable
	guest     string
	dcname    string
	lookup    map[string]string
}

func newEndpoint(cfg *vSphereConfig, url *url.URL, log log.Logger, m prometheus.Registerer) *endpoint {
	e := endpoint{
		cfg:           cfg,
		url:           url,
		initialized:   false,
		clientFactory: newClientFactory(log, url, cfg),
		log:           log,
	}

	e.resourceKinds = map[string]*resourceKind{
		"datacenter": {
			name:             "datacenter",
			vcName:           "Datacenter",
			enabled:          true,
			realTime:         false,
			sampling:         int32(cfg.HistoricalInterval.Seconds()),
			objects:          make(objectMap),
			paths:            []string{"/*"},
			collectInstances: cfg.DatacenterInstances,
			getObjects:       getDatacenters,
			parent:           "",
		},
		"cluster": {
			name:             "cluster",
			vcName:           "ClusterComputeResource",
			enabled:          true,
			realTime:         false,
			sampling:         int32(cfg.HistoricalInterval.Seconds()),
			objects:          make(objectMap),
			paths:            []string{"/*/host/**"},
			collectInstances: cfg.ClusterInstances,
			getObjects:       getClusters,
			parent:           "datacenter",
		},
		"host": {
			name:             "host",
			vcName:           "HostSystem",
			enabled:          true,
			realTime:         true,
			sampling:         20,
			objects:          make(objectMap),
			paths:            []string{"/*/host/**"},
			collectInstances: cfg.HostInstances,
			getObjects:       getHosts,
			parent:           "cluster",
		},
		"vm": {
			name:             "vm",
			vcName:           "VirtualMachine",
			enabled:          true,
			realTime:         true,
			sampling:         20,
			objects:          make(objectMap),
			paths:            []string{"/*/vm/**"},
			collectInstances: cfg.VMInstances,
			getObjects:       getVMs,
			parent:           "host",
		},
		"datastore": {
			name:             "datastore",
			vcName:           "Datastore",
			enabled:          true,
			realTime:         false,
			sampling:         int32(cfg.HistoricalInterval.Seconds()),
			objects:          make(objectMap),
			paths:            []string{"/*/datastore/**"},
			collectInstances: cfg.DatastoreInstances,
			getObjects:       getDatastores,
			parent:           "",
		},
	}

	if m != nil {
		e.dm = newDiscoveryMetrics(m)
	}

	return &e
}

// Note that canceling the context will cancel the discovery process.
func (e *endpoint) init(ctx context.Context) error {
	if e.cfg.ObjectDiscoveryInterval > 0 {
		e.initialDiscovery(ctx)
	}
	e.initialized = true
	return nil
}

func (e *endpoint) initialDiscovery(ctx context.Context) {
	err := e.discover(ctx)
	if err != nil && err != context.Canceled {
		level.Error(e.log).Log("msg", "error in initialDiscovery", "host", e.url.Host, "err", err.Error())
	}
	e.startDiscovery(ctx)
}

func (e *endpoint) startDiscovery(ctx context.Context) {
	e.discoveryTicker = time.NewTicker(e.cfg.ObjectDiscoveryInterval)
	go func() {
		for {
			select {
			case <-e.discoveryTicker.C:
				err := e.discover(ctx)
				if err != nil && err != context.Canceled {
					level.Error(e.log).Log("msg", "discovery error", "host", e.url.Host, "err", err.Error())
				}
			case <-ctx.Done():
				level.Debug(e.log).Log("msg", "existing discovery goroutine", "host", e.url.Host)
				e.discoveryTicker.Stop()
				return
			}
		}
	}()
}

func (e *endpoint) discover(ctx context.Context) error {
	level.Debug(e.log).Log("msg", "object discovery starting")
	defer level.Debug(e.log).Log("msg", "object discovery complete")

	if e.dm != nil {
		timer := prometheus.NewTimer(e.dm.duration)
		defer timer.ObserveDuration()
	}

	e.busy.Lock()
	defer e.busy.Unlock()
	if ctx.Err() != nil {
		return ctx.Err()
	}
	err := e.reloadMetricNameMap(ctx)
	if err != nil {
		return err
	}

	client, err := e.clientFactory.GetClient(ctx)
	if err != nil {
		return err
	}

	level.Debug(e.log).Log("msg", "discover new objects", "host", e.url.Host)
	dcNameCache := make(map[string]string)

	numRes := int64(0)

	// Populate resource objects, and endpoint instance info.
	newObjects := make(map[string]objectMap)
	for k, res := range e.resourceKinds {
		level.Debug(e.log).Log("msg", "discovering resources", "name", res.name)
		// Need to do this for all resource types even if they are not enabled
		if res.enabled || k != "vm" {
			rf := resourceFilter{
				finder:  &finder{client},
				resType: res.vcName,
				paths:   res.paths,
			}

			ctx1, cancel1 := context.WithTimeout(ctx, e.cfg.Timeout)
			objects, err := res.getObjects(ctx1, e, &rf)
			cancel1()
			if err != nil {
				return err
			}

			// Fill in datacenter names where available (no need to do it for Datacenters)
			if res.name != "datacenter" {
				for k, obj := range objects {
					if obj.parentRef != nil {
						obj.dcname, _ = e.getDatacenterName(ctx, client, dcNameCache, *obj.parentRef)
						objects[k] = obj
					}
				}
			}

			// No need to collect metric metadata if resource type is not enabled
			if res.enabled {
				e.simpleMetadataSelect(ctx, client, res)
			}
			newObjects[k] = objects
			numRes += int64(len(objects))
		}
	}

	// Atomically swap maps
	e.collectMux.Lock()
	defer e.collectMux.Unlock()

	for k, v := range newObjects {
		e.resourceKinds[k].objects = v
	}

	if e.dm != nil {
		e.dm.datacenters.Set(float64(len(e.resourceKinds["datacenter"].objects)))
		e.dm.datacentersTotal.Add(float64(len(e.resourceKinds["datacenter"].objects)))
		e.dm.clusters.Set(float64(len(e.resourceKinds["cluster"].objects)))
		e.dm.clustersTotal.Add(float64(len(e.resourceKinds["cluster"].objects)))
		e.dm.hosts.Set(float64(len(e.resourceKinds["host"].objects)))
		e.dm.hostsTotal.Add(float64(len(e.resourceKinds["host"].objects)))
		e.dm.virtualMachines.Set(float64(len(e.resourceKinds["vm"].objects)))
		e.dm.virtualMachinesTotal.Add(float64(len(e.resourceKinds["vm"].objects)))
		e.dm.datastores.Set(float64(len(e.resourceKinds["datastore"].objects)))
		e.dm.datastoresTotal.Add(float64(len(e.resourceKinds["datastore"].objects)))
		e.dm.discoveryCount.Inc()
	}

	return nil
}

func (e *endpoint) getDatacenterName(ctx context.Context, client *client, cache map[string]string, r types.ManagedObjectReference) (string, bool) {
	return e.getAncestorName(ctx, client, "Datacenter", cache, r)
}

func (e *endpoint) simpleMetadataSelect(ctx context.Context, client *client, res *resourceKind) {
	//e.log.Debugf("Using fast metric metadata selection for %s", res.name)
	m, err := client.counterInfoByName(ctx)
	if err != nil {
		//e.log.Errorf("Getting metric metadata. Discovery will be incomplete. Error: %s", err.Error())
		return
	}
	res.metrics = make(performance.MetricList, 0, len(m))
	for _, pci := range m {
		cnt := types.PerfMetricId{
			CounterId: pci.Key,
		}
		if res.collectInstances {
			cnt.Instance = "*"
		} else {
			cnt.Instance = ""
		}
		res.metrics = append(res.metrics, cnt)
	}
}

func (e *endpoint) reloadMetricNameMap(ctx context.Context) error {
	e.metricNameMux.Lock()
	defer e.metricNameMux.Unlock()
	client, err := e.clientFactory.GetClient(ctx)
	if err != nil {
		return err
	}

	mn, err := client.counterInfoByKey(ctx)
	if err != nil {
		return err
	}
	e.metricNameLookup = make(map[int32]string)
	for key, m := range mn {
		e.metricNameLookup[key] = m.Name()
	}
	return nil
}

func (e *endpoint) getAncestorName(ctx context.Context, client *client, resourceType string, cache map[string]string, r types.ManagedObjectReference) (string, bool) {
	path := make([]string, 0)
	returnVal := ""
	here := r
	done := false
	for !done {
		done = func() bool {
			if name, ok := cache[here.Reference().String()]; ok {
				// Populate cache for the entire chain of objects leading here.
				returnVal = name
				return true
			}
			path = append(path, here.Reference().String())
			o := object.NewCommon(client.Client.Client, r)
			var result mo.ManagedEntity
			ctx1, cancel1 := context.WithTimeout(ctx, e.cfg.Timeout)
			defer cancel1()
			err := o.Properties(ctx1, here, []string{"parent", "name"}, &result)
			if err != nil {
				//e.cfg.Log.Warnf("Error while resolving parent. Assuming no parent exists. Error: %s", err.Error())
				return true
			}
			if result.Reference().Type == resourceType {
				// Populate cache for the entire chain of objects leading here.
				returnVal = result.Name
				return true
			}
			if result.Parent == nil {
				//e.cfg.Log.Debugf("No parent found for %s (ascending from %s)", here.Reference(), r.Reference())
				return true
			}
			here = result.Parent.Reference()
			return false
		}()
	}
	for _, s := range path {
		cache[s] = returnVal
	}
	return returnVal, returnVal != ""
}

func getDatacenters(ctx context.Context, e *endpoint, resourceFilter *resourceFilter) (objectMap, error) {
	var resources []mo.Datacenter
	ctx1, cancel1 := context.WithTimeout(ctx, e.cfg.Timeout)
	defer cancel1()
	err := resourceFilter.findAll(ctx1, &resources)
	if err != nil {
		return nil, err
	}
	m := make(objectMap, len(resources))
	for _, r := range resources {
		m[r.ExtensibleManagedObject.Reference().Value] = &objectRef{
			name:      r.Name,
			ref:       r.ExtensibleManagedObject.Reference(),
			parentRef: r.Parent,
			dcname:    r.Name,
		}
	}
	return m, nil
}

func getClusters(ctx context.Context, e *endpoint, resourceFilter *resourceFilter) (objectMap, error) {
	var resources []mo.ClusterComputeResource
	ctx1, cancel1 := context.WithTimeout(ctx, e.cfg.Timeout)
	defer cancel1()
	err := resourceFilter.findAll(ctx1, &resources)
	if err != nil {
		return nil, err
	}
	cache := make(map[string]*types.ManagedObjectReference)
	m := make(objectMap, len(resources))
	for _, r := range resources {
		// Wrap in a function to make defer work correctly.
		err := func() error {
			// We're not interested in the immediate parent (a folder), but the data center.
			p, ok := cache[r.Parent.Value]
			if !ok {
				ctx2, cancel2 := context.WithTimeout(ctx, e.cfg.Timeout)
				defer cancel2()
				client, err := e.clientFactory.GetClient(ctx2)
				if err != nil {
					return err
				}
				o := object.NewFolder(client.Client.Client, *r.Parent)
				var folder mo.Folder
				ctx3, cancel3 := context.WithTimeout(ctx, e.cfg.Timeout)
				defer cancel3()
				err = o.Properties(ctx3, *r.Parent, []string{"parent"}, &folder)
				if err != nil {
					level.Warn(e.log).Log("msg", "error while getting folder parent", "err", err.Error())
					p = nil
				} else {
					pp := folder.Parent.Reference()
					p = &pp
					cache[r.Parent.Value] = p
				}
			}
			m[r.ExtensibleManagedObject.Reference().Value] = &objectRef{
				name:      r.Name,
				ref:       r.ExtensibleManagedObject.Reference(),
				parentRef: p,
			}
			if e.dm != nil {
				e.dm.clusters.Inc()
			}
			return nil
		}()
		if err != nil {
			return nil, err
		}
	}
	return m, nil
}

func getHosts(ctx context.Context, e *endpoint, resourceFilter *resourceFilter) (objectMap, error) {
	var resources []mo.HostSystem
	err := resourceFilter.findAll(ctx, &resources)
	if err != nil {
		return nil, err
	}
	m := make(objectMap)
	for _, r := range resources {
		m[r.ExtensibleManagedObject.Reference().Value] = &objectRef{
			name:      r.Name,
			ref:       r.ExtensibleManagedObject.Reference(),
			parentRef: r.Parent,
		}
	}
	return m, nil
}

func getVMs(ctx context.Context, e *endpoint, resourceFilter *resourceFilter) (objectMap, error) {
	var resources []mo.VirtualMachine
	ctx1, cancel1 := context.WithTimeout(ctx, e.cfg.Timeout)
	defer cancel1()
	err := resourceFilter.findAll(ctx1, &resources)
	if err != nil {
		return nil, err
	}
	m := make(objectMap)
	for _, r := range resources {
		if r.Runtime.PowerState != "poweredOn" {
			continue
		}
		guest := "unknown"
		uuid := ""
		lookup := make(map[string]string)

		// Extract host name
		if r.Guest != nil && r.Guest.HostName != "" {
			lookup["guesthostname"] = r.Guest.HostName
		}

		// Collect network information
		for _, net := range r.Guest.Net {
			if net.DeviceConfigId == -1 {
				continue
			}
			if net.IpConfig == nil || net.IpConfig.IpAddress == nil {
				continue
			}
			ips := make(map[string][]string)
			for _, ip := range net.IpConfig.IpAddress {
				addr := ip.IpAddress
				for _, ipType := range e.cfg.IPAddresses {
					if !(ipType == "ipv4" && isIPv4.MatchString(addr) ||
						ipType == "ipv6" && isIPv6.MatchString(addr)) {

						continue
					}

					// By convention, we want the preferred addresses to appear first in the array.
					if _, ok := ips[ipType]; !ok {
						ips[ipType] = make([]string, 0)
					}
					if ip.State == "preferred" {
						ips[ipType] = append([]string{addr}, ips[ipType]...)
					} else {
						ips[ipType] = append(ips[ipType], addr)
					}
				}
			}
			for ipType, ipList := range ips {
				lookup["nic/"+strconv.Itoa(int(net.DeviceConfigId))+"/"+ipType] = strings.Join(ipList, ",")
			}
		}

		// Sometimes Config is unknown and returns a nil pointer
		if r.Config != nil {
			guest = strings.TrimSuffix(r.Config.GuestId, "Guest")
			uuid = r.Config.Uuid
		}

		m[r.ExtensibleManagedObject.Reference().Value] = &objectRef{
			name:      r.Name,
			ref:       r.ExtensibleManagedObject.Reference(),
			parentRef: r.Runtime.Host,
			guest:     guest,
			altID:     uuid,
			lookup:    lookup,
		}
	}
	return m, nil
}

func getDatastores(ctx context.Context, e *endpoint, resourceFilter *resourceFilter) (objectMap, error) {
	var resources []mo.Datastore
	ctx1, cancel1 := context.WithTimeout(ctx, e.cfg.Timeout)
	defer cancel1()
	err := resourceFilter.findAll(ctx1, &resources)
	if err != nil {
		return nil, err
	}
	m := make(objectMap)
	for _, r := range resources {
		lunID := ""
		if r.Info != nil {
			info := r.Info.GetDatastoreInfo()
			if info != nil {
				lunID = info.Url
			}
		}
		m[r.ExtensibleManagedObject.Reference().Value] = &objectRef{
			name:      r.Name,
			ref:       r.ExtensibleManagedObject.Reference(),
			parentRef: r.Parent,
			altID:     lunID,
		}
	}
	return m, nil
}

type discoveryMetrics struct {
	// object counters
	datacenters          prometheus.Gauge
	clusters             prometheus.Gauge
	hosts                prometheus.Gauge
	virtualMachines      prometheus.Gauge
	datastores           prometheus.Gauge
	datacentersTotal     prometheus.Counter
	clustersTotal        prometheus.Counter
	hostsTotal           prometheus.Counter
	virtualMachinesTotal prometheus.Counter
	datastoresTotal      prometheus.Counter

	// misc
	duration       prometheus.Histogram
	discoveryCount prometheus.Counter
}

func newDiscoveryMetrics(reg prometheus.Registerer) *discoveryMetrics {
	m := &discoveryMetrics{}

	// vmx_discovery_datacenter_last
	m.datacenters = prometheus.NewGauge(prometheus.GaugeOpts{
		Namespace: "vmx",
		Subsystem: "discovery",
		Name:      "datacenter_last",
		Help:      "Count of datacenters discovered during last object discovery.",
	})

	// vmx_discovery_cluster_last
	m.clusters = prometheus.NewGauge(prometheus.GaugeOpts{
		Namespace: "vmx",
		Subsystem: "discovery",
		Name:      "cluster_last",
		Help:      "Count of clusters discovered during last object discovery.",
	})

	// vmx_discovery_host_last
	m.hosts = prometheus.NewGauge(prometheus.GaugeOpts{
		Namespace: "vmx",
		Subsystem: "discovery",
		Name:      "host_last",
		Help:      "Count of hosts discovered during last object discovery.",
	})

	// vmx_discovery_vm_last
	m.virtualMachines = prometheus.NewGauge(prometheus.GaugeOpts{
		Namespace: "vmx",
		Subsystem: "discovery",
		Name:      "vm_last",
		Help:      "Count of virtual machines discovered during last object discovery.",
	})

	// vmx_discovery_datastore_last
	m.datastores = prometheus.NewGauge(prometheus.GaugeOpts{
		Namespace: "vmx",
		Subsystem: "discovery",
		Name:      "datastore_last",
		Help:      "Count of datastores discovered during last object discovery.",
	})

	// vmx_discovery_datacenter_total
	m.datacentersTotal = prometheus.NewCounter(prometheus.CounterOpts{
		Namespace: "vmx",
		Subsystem: "discovery",
		Name:      "datacenter_total",
		Help:      "Total count of (non-unique) datacenters discovered.",
	})

	// vmx_discovery_cluster_total
	m.clustersTotal = prometheus.NewCounter(prometheus.CounterOpts{
		Namespace: "vmx",
		Subsystem: "discovery",
		Name:      "cluster_total",
		Help:      "Total count of (non-unique) clusters discovered.",
	})

	// vmx_discovery_host_total
	m.hostsTotal = prometheus.NewCounter(prometheus.CounterOpts{
		Namespace: "vmx",
		Subsystem: "discovery",
		Name:      "host_total",
		Help:      "Total count of (non-unique) hosts discovered.",
	})

	// vmx_discovery_vm_total
	m.virtualMachinesTotal = prometheus.NewCounter(prometheus.CounterOpts{
		Namespace: "vmx",
		Subsystem: "discovery",
		Name:      "vm_total",
		Help:      "Total count of (non-unique) virtual machines discovered.",
	})

	// vmx_discovery_datastore_total
	m.datastoresTotal = prometheus.NewCounter(prometheus.CounterOpts{
		Namespace: "vmx",
		Subsystem: "discovery",
		Name:      "datastore_total",
		Help:      "Total count of (non-unique) datastores discovered.",
	})

	// vmx_discovery_duration_seconds
	m.duration = prometheus.NewHistogram(prometheus.HistogramOpts{
		Namespace:   "vmx",
		Subsystem:   "discovery",
		Name:        "duration_seconds",
		Help:        "Histogram for discovery duration.",
		ConstLabels: nil,
		Buckets:     prometheus.ExponentialBuckets(0.1, 2, 10),
	})

	// vmx_discover_total
	m.discoveryCount = prometheus.NewCounter(prometheus.CounterOpts{
		Namespace: "vmx",
		Subsystem: "discovery",
		Name:      "total",
		Help:      "The number of times object discovery has executed successfully.",
	})

	if reg != nil {
		reg.MustRegister(
			m.datacenters,
			m.datacentersTotal,
			m.clusters,
			m.clustersTotal,
			m.hosts,
			m.hostsTotal,
			m.virtualMachines,
			m.virtualMachinesTotal,
			m.datastores,
			m.datastoresTotal,
			m.duration,
			m.discoveryCount,
		)
	}

	return m
}
