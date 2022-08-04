package remotewrite

// NOTE: This is a placeholder component for remote_write for testing of the appendable, it should NOT be considered final

import (
	"context"
	"fmt"
	"math"
	"path/filepath"
	"sync"
	"time"

	"github.com/grafana/agent/component/metrics"

	"github.com/go-kit/log"
	"github.com/go-kit/log/level"
	"github.com/grafana/agent/component"
	"github.com/grafana/agent/pkg/build"
	"github.com/grafana/agent/pkg/metrics/wal"
	"github.com/prometheus/prometheus/model/timestamp"
	"github.com/prometheus/prometheus/storage"
	"github.com/prometheus/prometheus/storage/remote"
)

// Options.
//
// TODO(rfratto): These should be flags. How do we want to handle static
// options for components?
var (
	walTruncateFrequency = 2 * time.Hour
	minWALTime           = 5 * time.Minute
	maxWALTime           = 8 * time.Hour
	remoteFlushDeadline  = 1 * time.Minute
)

func init() {
	remote.UserAgent = fmt.Sprintf("GrafanaAgent/%s", build.Version)

	component.Register(component.Registration{
		Name:    "metrics.remote_write",
		Args:    RemoteConfig{},
		Exports: Export{},
		Build: func(o component.Options, c component.Arguments) (component.Component, error) {
			return NewComponent(o, c.(RemoteConfig))
		},
	})
}

// Component is the metrics_forwarder component.
type Component struct {
	log  log.Logger
	opts component.Options

	walStore    *wal.Storage
	remoteStore *remote.Storage
	storage     storage.Storage

	mut sync.RWMutex
	cfg RemoteConfig

	receiver *metrics.Receiver
}

// NewComponent creates a new metrics_forwarder component.
func NewComponent(o component.Options, c RemoteConfig) (*Component, error) {
	walLogger := log.With(o.Logger, "subcomponent", "wal")
	dataPath := filepath.Join(o.DataPath, "wal", o.ID)
	walStorage, err := wal.NewStorage(walLogger, o.Registerer, dataPath)
	if err != nil {
		return nil, err
	}

	remoteLogger := log.With(o.Logger, "subcomponent", "rw")
	remoteStore := remote.NewStorage(remoteLogger, o.Registerer, startTime, dataPath, remoteFlushDeadline, nil)

	res := &Component{
		log:         o.Logger,
		opts:        o,
		walStore:    walStorage,
		remoteStore: remoteStore,
		storage:     storage.NewFanout(o.Logger, walStorage, remoteStore),
	}
	res.receiver = &metrics.Receiver{Receive: res.Receive}
	if err := res.Update(c); err != nil {
		return nil, err
	}
	return res, nil
}

func startTime() (int64, error) { return 0, nil }

var _ component.Component = (*Component)(nil)

// Run implements Component.
func (c *Component) Run(ctx context.Context) error {
	c.opts.OnStateChange(Export{Receiver: c.receiver})
	defer func() {
		level.Debug(c.log).Log("msg", "closing storage")
		err := c.storage.Close()
		level.Debug(c.log).Log("msg", "storage closed")
		if err != nil {
			level.Error(c.log).Log("msg", "error when closing storage", "err", err)
		}
	}()

	// Track the last timestamp we truncated for to prevent segments from getting
	// deleted until at least some new data has been sent.
	var lastTs = int64(math.MinInt64)

	for {
		select {
		case <-ctx.Done():
			return nil
		case <-time.After(walTruncateFrequency):
			// The timestamp ts is used to determine which series are not receiving
			// samples and may be deleted from the WAL. Their most recent append
			// timestamp is compared to ts, and if that timestamp is older then ts,
			// they are considered inactive and may be deleted.
			//
			// Subtracting a duration from ts will delay when it will be considered
			// inactive and scheduled for deletion.
			ts := c.remoteStore.LowestSentTimestamp() - minWALTime.Milliseconds()
			if ts < 0 {
				ts = 0
			}

			// Network issues can prevent the result of getRemoteWriteTimestamp from
			// changing. We don't want data in the WAL to grow forever, so we set a cap
			// on the maximum age data can be. If our ts is older than this cutoff point,
			// we'll shift it forward to start deleting very stale data.
			if maxTS := timestamp.FromTime(time.Now().Add(-maxWALTime)); ts < maxTS {
				ts = maxTS
			}

			if ts == lastTs {
				level.Debug(c.log).Log("msg", "not truncating the WAL, remote_write timestamp is unchanged", "ts", ts)
				continue
			}
			lastTs = ts

			level.Debug(c.log).Log("msg", "truncating the WAL", "ts", ts)
			err := c.walStore.Truncate(ts)
			if err != nil {
				// The only issue here is larger disk usage and a greater replay time,
				// so we'll only log this as a warning.
				level.Warn(c.log).Log("msg", "could not truncate WAL", "err", err)
			}
		}
	}
}

// Update implements Component.
func (c *Component) Update(newConfig component.Arguments) error {
	cfg := newConfig.(RemoteConfig)

	c.mut.Lock()
	defer c.mut.Unlock()

	convertedConfig, err := convertConfigs(cfg)
	if err != nil {
		return err
	}
	err = c.remoteStore.ApplyConfig(convertedConfig)
	if err != nil {
		return err
	}

	c.cfg = cfg
	return nil
}

// Receive implements the receiver.receive func that allows an array of metrics to be passed
func (c *Component) Receive(ts int64, metricArr []*metrics.FlowMetric) {
	app := c.storage.Appender(context.Background())
	for _, m := range metricArr {
		// TODO this should all be simplified into one call
		if m.GlobalRefID == 0 {
			globalID := metrics.GlobalRefMapping.GetOrAddGlobalRefID(m.Labels)
			m.GlobalRefID = globalID
		}
		localID := metrics.GlobalRefMapping.GetLocalRefID(c.opts.ID, m.GlobalRefID)
		newLocal, err := app.Append(storage.SeriesRef(localID), m.Labels, ts, m.Value)
		// Add link if there wasn't one before, and we received a valid local id
		if localID == 0 && newLocal != 0 {
			metrics.GlobalRefMapping.GetOrAddLink(c.opts.ID, uint64(newLocal), m.Labels)
		}
		if err != nil {
			_ = app.Rollback()
			//TODO what should we log and behave?
			level.Error(c.log).Log("err", err, "msg", "error receiving metrics", "component", c.opts.ID)
			return
		}
	}
	_ = app.Commit()
}

// Config implements Component.
func (c *Component) Config() RemoteConfig {
	c.mut.RLock()
	defer c.mut.RUnlock()
	return c.cfg
}
