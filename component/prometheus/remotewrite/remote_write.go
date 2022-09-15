package remotewrite

import (
	"context"
	"fmt"
	"math"
	"path/filepath"
	"sync"
	"time"

	"github.com/go-kit/log"
	"github.com/go-kit/log/level"
	"github.com/grafana/agent/component"
	"github.com/grafana/agent/component/prometheus"
	"github.com/grafana/agent/pkg/build"
	"github.com/grafana/agent/pkg/metrics/wal"
	"github.com/prometheus/prometheus/model/timestamp"
	"github.com/prometheus/prometheus/storage"
	"github.com/prometheus/prometheus/storage/remote"
)

// Options.
//
// TODO(rfratto): This should be exposed. How do we want to expose this?
var remoteFlushDeadline = 1 * time.Minute

func init() {
	remote.UserAgent = fmt.Sprintf("GrafanaAgent/%s", build.Version)

	component.Register(component.Registration{
		Name:    "prometheus.remote_write",
		Args:    Arguments{},
		Exports: Exports{},
		Build: func(o component.Options, c component.Arguments) (component.Component, error) {
			return NewComponent(o, c.(Arguments))
		},
	})
}

// Component is the prometheus.remote_write component.
type Component struct {
	log  log.Logger
	opts component.Options

	walStore    *wal.Storage
	remoteStore *remote.Storage
	storage     storage.Storage

	mut sync.RWMutex
	cfg Arguments

	receiver *prometheus.Receiver
}

// NewComponent creates a new prometheus.remote_write component.
func NewComponent(o component.Options, c Arguments) (*Component, error) {
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
	res.receiver = &prometheus.Receiver{Receive: res.Receive}
	if err := res.Update(c); err != nil {
		return nil, err
	}
	return res, nil
}

func startTime() (int64, error) { return 0, nil }

var _ component.Component = (*Component)(nil)

// Run implements Component.
func (c *Component) Run(ctx context.Context) error {
	c.opts.OnStateChange(Exports{Receiver: c.receiver})
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
		case <-time.After(c.truncateFrequency()):
			// We retrieve the current min/max keepalive time at once, since
			// retrieving them separately could lead to issues where we have an older
			// value for min which is now larger than max.
			c.mut.RLock()
			var (
				minWALTime = c.cfg.WALOptions.MinKeepaliveTime
				maxWALTime = c.cfg.WALOptions.MaxKeepaliveTime
			)
			c.mut.RUnlock()

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

func (c *Component) truncateFrequency() time.Duration {
	c.mut.RLock()
	defer c.mut.RUnlock()
	return c.cfg.WALOptions.TruncateFrequency
}

// Update implements Component.
func (c *Component) Update(newConfig component.Arguments) error {
	cfg := newConfig.(Arguments)

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
func (c *Component) Receive(ts int64, metricArr []*prometheus.FlowMetric) {
	app := c.storage.Appender(context.Background())
	for _, m := range metricArr {
		localID := prometheus.GlobalRefMapping.GetLocalRefID(c.opts.ID, m.GlobalRefID())
		// Currently it doesn't look like the storage interfaces mutate the labels, but thats not a strong
		// promise. So this should be treated with care.
		newLocal, err := app.Append(storage.SeriesRef(localID), m.RawLabels(), ts, m.Value())
		// Add link if there wasn't one before, and we received a valid local id
		if localID == 0 && newLocal != 0 {
			prometheus.GlobalRefMapping.GetOrAddLink(c.opts.ID, uint64(newLocal), m)
		}
		if err != nil {
			_ = app.Rollback()
			//TODO what should we log and behave?
			level.Error(c.log).Log("err", err, "msg", "error receiving metrics", "component", c.opts.ID)
			return
		}
	}

	err := app.Commit()
	if err != nil {
		level.Error(c.log).Log("msg", "failed to commit samples", "err", err)
	}
}

// Config implements Component.
func (c *Component) Config() Arguments {
	c.mut.RLock()
	defer c.mut.RUnlock()
	return c.cfg
}
