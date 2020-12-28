package prom

import (
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/go-kit/kit/log"
	"github.com/go-kit/kit/log/level"
	"github.com/grafana/agent/pkg/prom/instance"
	"github.com/grafana/agent/pkg/prom/wal"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	promwal "github.com/prometheus/prometheus/tsdb/wal"
)

const (
	DefaultCleanupAge    = 12 * time.Hour
	DefaultCleanupPeriod = 30 * time.Minute
)

var (
	discoveryError = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "agent_prometheus_cleaner_storage_discovery_error",
			Help: "Errors encountered discovering local storage paths",
		},
		[]string{"storage"},
	)

	segmentError = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "agent_prometheus_cleaner_segment_error",
			Help: "Errors encountered finding most recent WAL segments",
		},
		[]string{"storage"},
	)

	managedStorage = promauto.NewGauge(
		prometheus.GaugeOpts{
			Name: "agent_prometheus_cleaner_managed_storage",
			Help: "Number of storage directories associated with managed instances",
		},
	)

	abandonedStorage = promauto.NewGauge(
		prometheus.GaugeOpts{
			Name: "agent_prometheus_cleaner_abandoned_storage",
			Help: "Number of storage directories not associated with any managed instance",
		},
	)
)

// Operations that require interacting with the Prometheus Write Ahead Log (WAL)
// file format. This interface exists to allow easier testing of the cleaner.
type walOperations interface {
	// Get the last modified time of the last segment in a WAL
	lastModified(path string) (time.Time, error)
}

type walOperationsDefault struct{}

func (w *walOperationsDefault) lastModified(path string) (time.Time, error) {
	empty := time.Time{}

	existing, err := promwal.Open(nil, path)
	if err != nil {
		return empty, err
	}

	// We don't care if there are errors closing the abandoned WAL
	defer func() { _ = existing.Close() }()

	_, last, err := existing.Segments()
	if err != nil {
		return empty, err
	}

	if last == -1 {
		return empty, fmt.Errorf("unable to determine most recent segment for %s", path)
	}

	// full path to the most recent segment in this WAL
	lastSegment := promwal.SegmentName(path, last)
	segmentFile, err := os.Stat(lastSegment)
	if err != nil {
		return empty, err
	}

	return segmentFile.ModTime(), nil
}

type WALCleanerOpts func(c *WALCleaner)

// Override default age after which abandoned WALs can be removed
func WithCleanerMinAge(d time.Duration) WALCleanerOpts {
	return func(c *WALCleaner) {
		if d > 0 {
			c.minAge = d
		}
	}
}

// Override default period for checking for abandoned WALs
func WithCleanerPeriod(d time.Duration) WALCleanerOpts {
	return func(c *WALCleaner) {
		if d > 0 {
			c.ticker = time.NewTicker(d)
		}
	}
}

//
type WALCleaner struct {
	logger          log.Logger
	instanceManager instance.Manager
	walDirectory    string
	walOperations   walOperations
	minAge          time.Duration
	ticker          *time.Ticker
	done            chan bool
}

// Create a new cleaner that looks for abandoned WALs in the given directory and
// removes them if they haven't been modified in over minAge. Starts a goroutine
// to periodically run WALCleaner.CleanupStorage in a loop
func NewWALCleaner(logger log.Logger, manager instance.Manager, walDirectory string, opts ...WALCleanerOpts) *WALCleaner {
	c := &WALCleaner{
		logger:          log.With(logger, "component", "cleaner"),
		instanceManager: manager,
		walDirectory:    filepath.Clean(walDirectory),
		walOperations:   &walOperationsDefault{},
		minAge:          DefaultCleanupAge,
		done:            make(chan bool),
	}

	for _, opt := range opts {
		opt(c)
	}

	// If the caller hasn't set a ticker, create one with the default period
	if c.ticker == nil {
		c.ticker = time.NewTicker(DefaultCleanupPeriod)
	}

	go c.run()
	return c
}

// Get storage directories used for each ManagedInstance
func (c *WALCleaner) getManagedStorage(instances map[string]instance.ManagedInstance) map[string]bool {
	out := make(map[string]bool)

	for _, inst := range instances {
		out[inst.StorageDirectory()] = true
	}

	return out
}

// Get all Storage directories under walDirectory
func (c *WALCleaner) getAllStorage() []string {
	var out []string

	_ = filepath.Walk(c.walDirectory, func(p string, info os.FileInfo, err error) error {
		if err != nil {
			if os.IsNotExist(err) {
				// The root WAL directory doesn't exist. Maybe this Agent isn't responsible for any
				// instances yet. Log at debug since this isn't a big deal. We'll just try to crawl
				// the direction again on the next periodic run.
				level.Debug(c.logger).Log("msg", "WAL storage path does not exist", "path", p, err)
			} else {
				// Just log any errors traversing the WAL directory. This will potentially result
				// in a WAL (that has incorrect permissions or some similar problem) not being cleaned
				// up. This is  better than preventing *all* other WALs from being cleaned up.
				discoveryError.WithLabelValues(p).Inc()
				level.Warn(c.logger).Log("msg", "unable to traverse WAL storage path", "path", p, "err", err)
			}
		} else if info.IsDir() && filepath.Dir(p) == c.walDirectory {
			// Single level below the root are instance storage directories (including WALs)
			out = append(out, p)
		}
		return nil
	})

	return out
}

// Get the full path of storage directories that aren't associated with an active instance
// and haven't been written to within a configured duration (usually several hours or more).
func (c *WALCleaner) getAbandonedStorage(all []string, managed map[string]bool, now time.Time) []string {
	var out []string

	for _, dir := range all {
		if !managed[dir] {
			walDir := wal.SubDirectory(dir)
			mtime, err := c.walOperations.lastModified(walDir)
			if err != nil {
				segmentError.WithLabelValues(dir).Inc()
				level.Warn(c.logger).Log("msg", "unable to find segment mtime of WAL", "name", dir, "err", err)
				continue
			}

			diff := now.Sub(mtime)
			if diff > c.minAge {
				// The last segment for this WAL was modified more then $minAge (positive number of hours)
				// in the past. This makes it a candidate for deletion since it's also not associated with
				// any Instances this agent knows about.
				out = append(out, dir)
			}

			level.Debug(c.logger).Log("msg", "abandoned WAL", "name", dir, "mtime", mtime, "diff", diff)
		} else {
			level.Debug(c.logger).Log("msg", "active WAL", "name", dir)
		}
	}

	return out
}

func (c *WALCleaner) run() {
	for {
		select {
		case <-c.done:
			return
		case <-c.ticker.C:
			c.CleanupStorage()
		}
	}
}

// Remove any abandoned and unused WAL directories. Note that it shouldn't be
// necessary to call this method explicitly in most cases since it will be run
// periodically in a goroutine (started when WALCleaner is created).
func (c *WALCleaner) CleanupStorage() {
	all := c.getAllStorage()
	managed := c.getManagedStorage(c.instanceManager.ListInstances())
	abandoned := c.getAbandonedStorage(all, managed, time.Now())

	managedStorage.Set(float64(len(managed)))
	abandonedStorage.Set(float64(len(abandoned)))

	for _, a := range abandoned {
		level.Info(c.logger).Log("msg", "deleting abandoned WAL", "name", a)
		err := os.RemoveAll(a)
		if err != nil {
			level.Error(c.logger).Log("msg", "failed to delete abandoned WAL", "name", a, "err", err)
		}
	}
}

// Stop the cleaner and any background tasks running
func (c *WALCleaner) Stop() {
	level.Debug(c.logger).Log("msg", "stopping cleaner...")
	c.ticker.Stop()
	close(c.done)
}
