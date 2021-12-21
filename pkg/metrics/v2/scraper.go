package metrics

import (
	"context"
	"sync"
	"time"

	"github.com/go-kit/log"
	"github.com/go-kit/log/level"
	pb "github.com/grafana/agent/pkg/metrics/v2/internal/metricspb"
	"github.com/grafana/agent/pkg/util"
	"github.com/prometheus/client_golang/prometheus"
	prom_config "github.com/prometheus/prometheus/config"
	"github.com/prometheus/prometheus/discovery/targetgroup"
	"github.com/prometheus/prometheus/scrape"
	"github.com/prometheus/prometheus/storage"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type scraperMetrics struct {
	util.MetricsContainer

	numberScrapers    prometheus.Gauge
	totalTargetPushes prometheus.Counter
	totalTargets      prometheus.Counter
	totalFailedPushes *prometheus.CounterVec

	scraperTargets *prometheus.GaugeVec
}

func newScraperMetrics() *scraperMetrics {
	var m scraperMetrics

	m.numberScrapers = prometheus.NewGauge(prometheus.GaugeOpts{
		Name: "agent_metrics_scrapers_count",
		Help: "Current number of running scrapers",
	})
	m.totalTargetPushes = prometheus.NewCounter(prometheus.CounterOpts{
		Name: "agent_metrics_scrapers_received_pushes_total",
		Help: "Number of times this node has received a set of targets from other nodes",
	})
	m.totalTargets = prometheus.NewCounter(prometheus.CounterOpts{
		Name: "agent_metrics_scrapers_received_targets_total",
		Help: "Total number of targets this node has received from other nodes",
	})
	m.totalFailedPushes = prometheus.NewCounterVec(prometheus.CounterOpts{
		Name: "agent_metrics_scrapers_failed_pushes_total",
		Help: "Number of times scrapers failed to receieve targets",
	}, []string{"reason"})

	m.scraperTargets = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Name: "agent_metrics_scraper_targets",
		Help: "Current number of targets for the scraper",
	}, []string{"instance"})

	m.Add(
		m.numberScrapers,
		m.totalTargetPushes,
		m.totalTargets,
		m.totalFailedPushes,

		m.scraperTargets,
	)
	return &m
}

// scraperManager implements metricspb.ScraperServer and will manage scrapers
// based on receieved targets.
type scraperManager struct {
	pb.UnimplementedScraperServer

	log     log.Logger
	ss      storageSet
	metrics *scraperMetrics

	mut              sync.RWMutex
	stopped          bool
	instanceGroups   map[string]targetGroups // Map of instance name -> target groups
	scraperInstances map[string]*scraper     // Running scrapers.
}

type targetGroups = map[string][]*targetgroup.Group

// newScraperManager creates a new scraperManager. No scapers are available
// until ApplyConfig is called.
func newScraperManager(log log.Logger, ss storageSet) *scraperManager {
	return &scraperManager{
		log:     log,
		ss:      ss,
		metrics: newScraperMetrics(),

		instanceGroups:   make(map[string]targetGroups),
		scraperInstances: make(map[string]*scraper),
	}
}

// ApplyConfig will update the managed set of scrapers. ApplyConfig should not
// be called util the storageSet is reconfigured first.
func (s *scraperManager) ApplyConfig(cfg *Config) error {
	s.mut.Lock()
	defer s.mut.Unlock()

	var firstError error
	saveError := func(e error) {
		if firstError == nil {
			firstError = e
		}
	}

	currentConfigs := make(map[string]struct{}, len(cfg.Configs))

	// TODO(rfratto): dynamically spin up/down scrapers based on current targets.
	for _, ic := range cfg.Configs {
		currentConfigs[ic.Name] = struct{}{}

		// Get or create the scraper for this instance config.
		scraper, ok := s.scraperInstances[ic.Name]
		if !ok {
			app, err := s.ss.Appendable(ic.Name)
			if err != nil {
				level.Error(s.log).Log("msg", "failed to retrieve storage for instance", "instance", ic.Name, "err", err)
				saveError(err)
				continue
			}

			l := log.With(s.log, "component", "metrics.scraper", "instance", ic.Name)
			scraper = newScraper(l, app)
			s.scraperInstances[ic.Name] = scraper
		}

		// Then give it its new set of scrape configs.
		if err := scraper.ApplyConfig(ic.ScrapeConfigs); err != nil {
			level.Error(s.log).Log("msg", "failed to apply config to scraper", "instance", ic.Name, "err", err)
			saveError(err)
		}
	}

	// Remove any scrapers that have gone away between reloads.
	for instance, inst := range s.scraperInstances {
		_, exist := currentConfigs[instance]
		if !exist {
			level.Info(s.log).Log("msg", "shutting down stale instance scraper", "instance", instance)
			inst.Stop()
			delete(s.scraperInstances, instance)

			s.metrics.scraperTargets.DeleteLabelValues(instance)
		}
	}

	s.metrics.numberScrapers.Set(float64(len(s.scraperInstances)))
	return firstError
}

// Collector returns metrics for the scraperManager.
func (s *scraperManager) Collector() prometheus.Collector { return s.metrics }

// Stop stops all of the scrapers.
func (s *scraperManager) Stop() {
	s.mut.Lock()
	defer s.mut.Unlock()
	s.stopped = true

	for _, scraper := range s.scraperInstances {
		scraper.Stop()
	}
}

func (s *scraperManager) ScrapeTargets(ctx context.Context, req *pb.ScrapeTargetsRequest) (*pb.ScrapeTargetsResponse, error) {
	s.mut.Lock()
	defer s.mut.Unlock()

	if s.stopped {
		return nil, status.Errorf(codes.Unavailable, "scraper is shutting down")
	}

	s.metrics.totalTargetPushes.Inc()
	for _, groups := range req.GetTargets() {
		for _, group := range groups.GetGroups() {
			s.metrics.totalTargets.Add(float64(len(group.GetTargets())))
		}
	}

	// Find the scraper to update. We don't want to bother tracking targets we
	// can't scrape from, so we use this as an early validation.
	scraper, ok := s.scraperInstances[req.GetInstanceName()]
	if !ok {
		// TODO(rfratto): This actually doesn't work that well. It's entirely
		// possible that a new config is rolled out gradually and scraperServer
		// just doesn't know about it yet.
		//
		// Rather, we _should_ keep targets from jobs we don't know about in case
		// they show up eventually. We can garbage collect them if they've been
		// around too long with nowhere to assign them, though.
		level.Error(s.log).Log("msg", "unknown instance, can't scrape targets", "instance", req.InstanceName)
		s.metrics.totalFailedPushes.WithLabelValues("unknown_instance").Inc()
		return nil, status.Errorf(codes.NotFound, "instance %q not known to scraper", req.InstanceName)
	}

	// Merge our instance groups.
	inGroupSet := pb.PrometheusGroups(req.GetTargets())
	outGroupSet, ok := s.instanceGroups[req.InstanceName]
	if !ok {
		outGroupSet = make(targetGroups, len(inGroupSet))
		s.instanceGroups[req.InstanceName] = outGroupSet
	}

	// Iterate over our input and override the set of targets for everything in
	// that job.
	for job, groups := range inGroupSet {
		outGroupSet[job] = groups
	}

	// Target calculation. This is done after the merge to accurately track the
	// new final set of targets.
	var numTargets int
	for _, groups := range outGroupSet {
		for _, group := range groups {
			numTargets += len(group.Targets)
		}
	}

	select {
	case scraper.syncCh <- outGroupSet:
		level.Debug(s.log).Log("msg", "passed new targets to instance scraper", "instance", req.InstanceName)
		s.metrics.scraperTargets.WithLabelValues(req.GetInstanceName()).Set(float64(numTargets))
	case <-ctx.Done():
		level.Error(s.log).Log("msg", "context canceled while assigning new targets to instance scraper", "insatnce", req.InstanceName, "err", ctx.Err())
		s.metrics.totalFailedPushes.WithLabelValues("timeout").Inc()
	}
	return &pb.ScrapeTargetsResponse{}, nil
}

// getScrapeTargets lists all current scrape targets.
func (sm *scraperManager) getScrapeTargets() []scrapeTarget {
	sm.mut.RLock()
	defer sm.mut.RUnlock()

	var targets []scrapeTarget
	for instName, inst := range sm.scraperInstances {
		targets = append(targets, inst.getScrapeTargets(instName)...)
	}
	return targets
}

// scraper manages all of the scraping for a metrics instance.
type scraper struct {
	syncCh chan<- targetGroups
	sm     *scrape.Manager
}

// newScraper constructs a new scraper which will deliver all sent metrics to app.
// newScraper will run until Stop is called.
func newScraper(l log.Logger, app storage.Appendable) *scraper {
	sm := scrape.NewManager(&scrape.Options{}, l, app)

	syncCh := make(chan targetGroups)
	go sm.Run(syncCh)

	return &scraper{
		syncCh: syncCh,
		sm:     sm,
	}
}

// ApplyConfig will inform the scraper of jobs it is responsible for. This MUST
// be called before sending it any targets.
func (s *scraper) ApplyConfig(cc []*prom_config.ScrapeConfig) error {
	return s.sm.ApplyConfig(&prom_config.Config{ScrapeConfigs: cc})
}

// getScrapeTargets returns the current scrape targets.
func (s *scraper) getScrapeTargets(instanceName string) []scrapeTarget {
	var targets []scrapeTarget
	activeTargets := s.sm.TargetsActive()

	for groupName, groupedTargets := range activeTargets {
		for _, target := range groupedTargets {
			outTarget := scrapeTarget{
				Instance:    instanceName,
				TargetGroup: groupName,

				Endpoint:         target.URL().String(),
				State:            string(target.Health()),
				Labels:           target.Labels(),
				DiscoveredLabels: target.DiscoveredLabels(),
				LastScrape: func() *time.Time {
					t := target.LastScrape()
					if t.IsZero() {
						return nil
					}
					return &t
				}(),
				ScrapeDuration: target.LastScrapeDuration().Milliseconds(),
				ScrapeError: func() string {
					err := target.LastError()
					if err != nil {
						return err.Error()
					}
					return ""
				}(),
			}
			targets = append(targets, outTarget)
		}
	}

	return targets
}

// Stop stops the scraper and all scrape jobs.
func (s *scraper) Stop() {
	s.sm.Stop()
}
