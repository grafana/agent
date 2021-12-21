package metrics

import (
	"context"
	"fmt"
	"sort"
	"sync"
	"time"

	"github.com/go-kit/log"
	"github.com/go-kit/log/level"
	"github.com/grafana/agent/pkg/metrics/v2/internal/metricspb"
	"github.com/prometheus/common/model"
	"github.com/prometheus/prometheus/discovery"
	"github.com/prometheus/prometheus/discovery/targetgroup"
	"github.com/rfratto/ckit"
	"google.golang.org/grpc"
)

// TODO(rfratto): would it be more stable for a discoverer to broadcast the
// full set of targets to all agents and have them pick from the bucket? It
// might be noisier at the traffic level but would be more resilient to network
// partitions.

// discovererManager manages a set of discoverers. Discoverers are launched
// based on cluster ownership: a discoverer will only run Service Discovery for
// job names that hash to the local node.
type discovererManager struct {
	log    log.Logger
	hasher *hasher
	self   metricspb.ScraperServer

	ctx    context.Context
	cancel context.CancelFunc

	mut                 sync.RWMutex
	discovererInstances map[string]*discoverer
	configCh            chan *Config
	hashReaderCh        chan *hashReader
}

// newDiscovererManager creates a new discovererManager. No discoverers are available until ApplyConfig is called.
func newDiscovererManager(log log.Logger, hasher *hasher, self metricspb.ScraperServer) *discovererManager {
	ctx, cancel := context.WithCancel(context.Background())

	dm := &discovererManager{
		log:    log,
		hasher: hasher,
		self:   self,

		ctx:    ctx,
		cancel: cancel,

		discovererInstances: make(map[string]*discoverer),
		configCh:            make(chan *Config, 1),
		hashReaderCh:        make(chan *hashReader, 1),
	}
	go dm.run(ctx)

	hasher.OnPeersChanged(func(hr *hashReader) bool {
		dm.hashReaderCh <- hr
		return true
	})
	return dm
}

// ApplyConfig will run a set of discoverers.
func (dm *discovererManager) ApplyConfig(cfg *Config) error {
	// Because we need to immediately filter jobs from cfg, we can't apply things
	// directly here. Instead, we have to queue it to a channel. Unfortunately,
	// this means that some runtime configuration errors may not be returned to
	// the user.
	//
	// TODO(rfratto): we can fix this though by making sure that our cluster node
	// is started before any of this gets invoked.
	dm.configCh <- cfg
	return nil
}

func (dm *discovererManager) run(ctx context.Context) {
	var (
		hr  *hashReader
		cfg *Config
	)

	distributeDiscovery := func() {
		if hr == nil {
			level.Debug(dm.log).Log("msg", "skipping distribution of discovery jobs because cluster is still being initialized")
			return
		}
		if cfg == nil {
			level.Debug(dm.log).Log("msg", "skipping distribution of discovery jobs because no config has been loaded yet")
			return
		}
		dm.distributeDiscovery(cfg, hr)
	}

	for {
		select {
		case <-ctx.Done():
			return

		case hr = <-dm.hashReaderCh:
			distributeDiscovery()
		case cfg = <-dm.configCh:
			distributeDiscovery()
		}
	}
}

// distributeDiscovery will assign SD jobs to our discoverers.
func (dm *discovererManager) distributeDiscovery(cfg *Config, hr *hashReader) {
	dm.mut.Lock()
	defer dm.mut.Unlock()

	level.Info(dm.log).Log("msg", "distributing discovery jobs", "configs", len(cfg.Configs))

	currentConfigs := make(map[string]struct{}, len(cfg.Configs))
	for _, ic := range cfg.Configs {
		currentConfigs[ic.Name] = struct{}{}

		disc, ok := dm.discovererInstances[ic.Name]
		if !ok {
			disc = newDiscoverer(dm.ctx, ic.Name, dm.log, dm.hasher, dm.self)
			dm.discovererInstances[ic.Name] = disc
		}
		if err := disc.ApplyConfig(shardDiscoveryJobs(&ic, hr)); err != nil {
			level.Error(dm.log).Log("msg", "failed to apply discovery jobs", "instance", ic.Name, "err", err)
			continue
		}
	}

	// Shut down old discoverers for instances that have gone away.
	for instance, disc := range dm.discovererInstances {
		_, exist := currentConfigs[instance]
		if !exist {
			level.Info(dm.log).Log("msg", "shutting down stale discoverer", "instance", instance)
			disc.Stop()
			delete(dm.discovererInstances, instance)
		}
	}
}

func shardDiscoveryJobs(ic *InstanceConfig, hr *hashReader) map[string]discovery.Configs {
	res := make(map[string]discovery.Configs, len(ic.ScrapeConfigs)/len(hr.Peers()))
	for _, sc := range ic.ScrapeConfigs {
		// Assign the job to ourselves if we can't find the owner or the owner is us.
		peer, err := hr.Get(sc.JobName)
		if err != nil || peer == nil || peer.Self {
			res[sc.JobName] = sc.ServiceDiscoveryConfigs
		}
	}
	return res
}

func (dm *discovererManager) getDiscoveryJobs() discoveryJobs {
	dm.mut.RLock()
	defer dm.mut.RUnlock()

	var jobs discoveryJobs
	for instance, disc := range dm.discovererInstances {
		jobs.Instances = append(jobs.Instances, discoveryJobsInstance{
			Name: instance,
			Jobs: disc.jobNames,
		})
	}
	return jobs
}

func (dm *discovererManager) getDiscoveryTargets() discoveryTargets {
	dm.mut.RLock()
	defer dm.mut.RUnlock()

	var targets discoveryTargets
	for instance, disc := range dm.discovererInstances {
		targets.Instances = append(targets.Instances, discoveryTargetsInstance{
			Name:   instance,
			Groups: disc.lastTargets,
		})
	}
	sort.Slice(targets.Instances, func(i, j int) bool {
		return targets.Instances[i].Name < targets.Instances[j].Name
	})
	return targets
}

// Stop will stop dm and all running discoverers.
func (dm *discovererManager) Stop() {
	dm.mut.Lock()
	defer dm.mut.Unlock()

	// Calling cancel will immediately send the signal to our discoverers to
	// stop. We still call Stop directly on everything so we can wait for them to
	// finish running.
	dm.cancel()

	for _, disc := range dm.discovererInstances {
		disc.Stop()
	}
}

// A discoverer will perform service discovery for a specific instance.
// Discoverers are only launched when there are targets to discover. When a set
// of targets is found, a discoverer will shard targets amongst scrapers in the
// cluster.
type discoverer struct {
	log    log.Logger
	hasher *hasher
	name   string
	self   metricspb.ScraperServer

	mut         sync.Mutex
	jobNames    []string
	lastTargets targetGroups

	m            *discovery.Manager
	cancel       context.CancelFunc
	exited       chan struct{}
	hashReaderCh chan *hashReader
}

// newDiscoverer creates a new discoverer. Must call ApplyConfig to start
// discovering targets. Can be stopped by calling Stop. Discovered targets are
// sharded amongst scrapers using node.
func newDiscoverer(ctx context.Context, name string, l log.Logger, hasher *hasher, self metricspb.ScraperServer) *discoverer {
	ctx, cancel := context.WithCancel(ctx)

	l = log.With(l, "component", "metrics.discovery")
	m := discovery.NewManager(ctx, l, discovery.Name(fmt.Sprintf("metrics.discovery.%s", name)))
	go func() {
		_ = m.Run()
	}()

	disc := &discoverer{
		log:    l,
		hasher: hasher,
		name:   name,
		self:   self,

		m:            m,
		cancel:       cancel,
		exited:       make(chan struct{}),
		hashReaderCh: make(chan *hashReader),
	}
	go disc.run(ctx)

	hasher.OnPeersChanged(func(hr *hashReader) bool {
		select {
		case <-disc.exited:
			return false
		case disc.hashReaderCh <- hr:
			return true
		}
	})
	return disc
}

// ApplyConfig applies SD jobs to d.
func (d *discoverer) ApplyConfig(sd map[string]discovery.Configs) error {
	d.mut.Lock()
	defer d.mut.Unlock()

	d.jobNames = d.jobNames[:0]
	for jobName := range sd {
		d.jobNames = append(d.jobNames, jobName)
	}
	sort.Strings(d.jobNames)

	// TODO(rfratto): I'm not confident that ApplyConfig will force a write to
	// SyncCh.
	return d.m.ApplyConfig(sd)
}

func (d *discoverer) run(ctx context.Context) {
	defer close(d.exited)

	var (
		hr     *hashReader
		groups targetGroups
	)

	distributeShards := func() {
		if groups == nil {
			level.Debug(d.log).Log("msg", "skipping target distribution because no targets have been found yet")
			return
		}
		if hr == nil {
			level.Debug(d.log).Log("msg", "skipping target distribution because the cluster is still initializing")
			return
		}
		// TODO(rfratto): configurable timeout for local reshards.
		ctx, cancel := context.WithTimeout(ctx, 30*time.Second)
		defer cancel()

		level.Debug(d.log).Log("msg", "distributing targets", "instance", d.name)
		d.distributeShards(ctx, d.shard(groups, hr))
	}

	for {
		select {
		case <-ctx.Done():
			return

		case hr = <-d.hashReaderCh:
			distributeShards()
		case groups = <-d.m.SyncCh():
			d.saveGroups(groups)
			distributeShards()
		}
	}
}

func (d *discoverer) saveGroups(groups targetGroups) {
	d.mut.Lock()
	defer d.mut.Unlock()
	d.lastTargets = groups
}

func (d *discoverer) shard(set targetGroups, hr *hashReader) map[*ckit.Peer]targetGroups {
	if set == nil {
		return nil
	}

	// Store the set of all
	var ourselves *ckit.Peer
	for _, p := range hr.Peers() {
		if p.Self {
			ourselves = p
		}
	}

	// Create our full set of shards.
	shards := make(map[*ckit.Peer]targetGroups)
	for _, p := range hr.Peers() {
		shards[p] = make(map[string][]*targetgroup.Group)
	}

	for job, groups := range set {
		// Each shard must have an entry for job. This informs other peers when
		// they must shut down any targets they may have previously had for a
		// specific job.
		jobShards := make(map[*ckit.Peer][]*targetgroup.Group)
		for _, p := range hr.Peers() {
			// We initialize the capacity as if distribution would be perfect. This
			// won't cause us to overallocate on average.
			jobShards[p] = make([]*targetgroup.Group, 0, len(groups)/len(shards))
		}

		for _, group := range groups {
			// For simplicity, we're also going to shard the groups here. However, we
			// won't actually put them in the jobShard if they're empty. This will cause
			// some overallocations, but it's the easiest way of filling everything in.
			groupShards := make(map[*ckit.Peer]*targetgroup.Group)
			for _, p := range hr.Peers() {
				groupShards[p] = &targetgroup.Group{
					Targets: make([]model.LabelSet, 0, len(groups)/len(shards)),
					Labels:  group.Labels,
					Source:  group.Source,
				}
			}

			for _, target := range group.Targets {
				// Find which node in the cluster owns the target. If we fail to get a
				// peer, then we'll force ourselves to own it to keep things working.
				address := target[model.AddressLabel]
				peer, err := hr.Get(string(address))
				if err != nil || peer == nil {
					peer = ourselves
				}
				groupShards[peer].Targets = append(groupShards[peer].Targets, target)
			}

			for p, groupShard := range groupShards {
				if len(groupShard.Targets) == 0 {
					continue
				}
				jobShards[p] = append(jobShards[p], groupShard)
			}
		}

		for p, jobShard := range jobShards {
			shards[p][job] = jobShard
		}
	}

	return shards
}

// distributeShards will send the targetGroups to all peers in shards.
func (d *discoverer) distributeShards(ctx context.Context, shards map[*ckit.Peer]targetGroups) error {
	var (
		firstError    error
		firstErrorMut sync.Mutex
	)
	saveError := func(e error) {
		if e == nil {
			return
		}
		firstErrorMut.Lock()
		defer firstErrorMut.Unlock()
		if firstError == nil {
			firstError = e
		}
	}

	var wg sync.WaitGroup
	wg.Add(len(shards))

	for p, tgroups := range shards {
		go func(peer *ckit.Peer, tgroups targetGroups) {
			defer wg.Done()

			req := &metricspb.ScrapeTargetsRequest{
				InstanceName: d.name,
				Targets:      metricspb.ProtoGroups(tgroups),
			}

			var err error
			if peer.Self {
				// Never use the network for self-delivery. This allows the discoverer
				// to be unaware of whether a single-node cluster is listening for gRPC
				// network traffic at all.
				//
				// TODO(rfratto): create an in-memory connection to gRPC instead? That
				// would be helpful to reduce duplication here.
				_, err = d.self.ScrapeTargets(ctx, req)
			} else {
				var cc *grpc.ClientConn
				cc, err = grpc.Dial(peer.ApplicationAddr, grpc.WithInsecure())
				if err != nil {
					level.Error(d.log).Log("msg", "cannot send targets to peer", "peer", peer.Name, "addr", peer.ApplicationAddr, "err", err)
					saveError(err)
					return
				}
				cli := metricspb.NewScraperClient(cc)
				_, err = cli.ScrapeTargets(ctx, req)
			}
			if err != nil {
				level.Error(d.log).Log("msg", "failed to send targets to peer", "peer", peer.Name, "addr", peer.ApplicationAddr, "err", err)
				saveError(err)
			}
		}(p, tgroups)
	}

	wg.Wait()
	return ctx.Err()
}

// Stop will stop the discoverer.
func (d *discoverer) Stop() {
	d.cancel()
	<-d.exited
}
