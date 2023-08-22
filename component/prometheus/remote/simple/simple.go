package simple

import (
	"context"
	"github.com/go-kit/log/level"
	"net/url"
	"path"
	"sync"
	"time"

	"github.com/grafana/agent/component"
	promtype "github.com/grafana/agent/component/prometheus"
	"github.com/prometheus/client_golang/prometheus"
	config_util "github.com/prometheus/common/config"
	"github.com/prometheus/common/model"
	"github.com/prometheus/prometheus/storage"
)

func init() {
	component.Register(component.Registration{
		Name:      "prometheus.remote.simple",
		Singleton: false,
		Args:      Arguments{},
		Exports:   Exports{},
		Build: func(opts component.Options, args component.Arguments) (component.Component, error) {
			return NewComponent(opts, args.(Arguments))
		},
	})
}

func NewComponent(opts component.Options, args Arguments) (*Simple, error) {
	database, err := newDBStore(false, args.TTL, path.Join(opts.DataPath, "wal"), opts.Registerer, opts.Logger)
	if err != nil {
		return nil, err
	}
	s := &Simple{
		database: database,
		opts:     opts,
	}
	return s, s.Update(args)
}

type Simple struct {
	mut      sync.RWMutex
	database *dbstore
	args     Arguments
	opts     component.Options
	wr       *writer
}

// Run starts the component, blocking until ctx is canceled or the component
// suffers a fatal error. Run is guaranteed to be called exactly once per
// Component.
//
// Implementations of Componen should perform any necessary cleanup before
// returning from Run.
func (s *Simple) Run(ctx context.Context) error {
	qm, err := s.newQueueManager()
	if err != nil {
		return err
	}
	go s.database.Run(ctx)
	wr := newWriter(s.opts.ID, qm, s.database, s.opts.Logger)
	s.wr = wr
	go wr.Start(ctx)
	go qm.Start()
	go s.cleanupDB(ctx)
	<-ctx.Done()
	return nil
}

func (s *Simple) newQueueManager() (*QueueManager, error) {
	ew := newEWMARate(ewmaWeight, shardUpdateDuration)
	endUrl, err := url.Parse(s.args.Endpoint.URL)
	if err != nil {
		return nil, err
	}
	cfgURL := &config_util.URL{URL: endUrl}
	wr, err := NewWriteClient(s.opts.ID, &ClientConfig{
		URL:              cfgURL,
		Timeout:          model.Duration(s.args.Endpoint.RemoteTimeout),
		HTTPClientConfig: *s.args.Endpoint.HTTPClientConfig.Convert(),
		SigV4Config:      nil,
		Headers:          s.args.Endpoint.Headers,
		RetryOnRateLimit: s.args.Endpoint.QueueOptions.toPrometheusType().RetryOnRateLimit,
	})
	if err != nil {
		return nil, err
	}
	met := newQueueManagerMetrics(s.opts.Registerer, "", wr.Endpoint())

	qm := NewQueueManager(
		met,
		s.opts.Logger,
		ew,
		s.args.Endpoint.QueueOptions.toPrometheusType(),
		s.args.Endpoint.MetadataOptions.toPrometheusType(),
		wr,
		1*time.Minute,
		&maxTimestamp{
			Gauge: prometheus.NewGauge(prometheus.GaugeOpts{
				Namespace: "prometheus",
				Subsystem: "remote_storage",
				Name:      "highest_timestamp_in_seconds",
				Help:      "Highest timestamp that has come into the remote storage via the Appender interface, in seconds since epoch.",
			}),
		},
		true,
		true,
	)
	return qm, nil
}

// Update provides a new Config to the component. The type of newConfig will
// always match the struct type which the component registers.
//
// Update will be called concurrently with Run. The component must be able to
// gracefully handle updating its config while still running.
//
// An error may be returned if the provided config is invalid.
func (s *Simple) Update(args component.Arguments) error {
	s.mut.Lock()
	defer s.mut.Unlock()

	s.args = args.(Arguments)
	s.opts.OnStateChange(Exports{Receiver: s})

	return nil
}

// Appender returns a new appender for the storage. The implementation
// can choose whether or not to use the context, for deadlines or to check
// for errors.
func (c *Simple) Appender(ctx context.Context) storage.Appender {
	return newAppender(c)
}

func (c *Simple) commit(a *appender) {
	c.mut.Lock()
	defer c.mut.Unlock()
	endpoint := time.Now().UnixMilli() - int64(c.args.TTL.Seconds())

	timestampedMetrics := make([]promtype.Sample, 0)
	for _, x := range a.metrics {
		// No need to write if already outside of range and a ttl is set.
		if x.Timestamp < endpoint && (c.args.TTL.Seconds() != 0) {
			continue
		}
		timestampedMetrics = append(timestampedMetrics, x)
	}

	_, err := c.database.WriteSignal(timestampedMetrics)
	if err != nil {
		level.Error(c.opts.Logger).Log("msg", "error writing signals", "err", err)
	}
}

func (c *Simple) cleanupDB(ctx context.Context) {
	c.cleanup()
	ttlTimer := time.NewTicker(5 * time.Minute)
	for {
		select {
		case <-ttlTimer.C:
			c.cleanup()
		case <-ctx.Done():
			return
		}
	}
}

func (c *Simple) cleanup() {
	level.Info(c.opts.Logger).Log("msg", "starting evict")
	oldestKey := c.wr.GetKey()
	c.database.sampleDB.DeleteKeysOlderThan(oldestKey)
	c.database.evict()
	level.Info(c.opts.Logger).Log("msg", "finishing evict")
}
