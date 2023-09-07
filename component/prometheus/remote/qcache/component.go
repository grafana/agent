package qcache

import (
	"arena"
	"context"
	"net/url"
	"path"
	"sync"
	"time"

	"github.com/go-kit/log/level"
	"github.com/grafana/agent/component"
	"github.com/grafana/agent/component/prometheus/remote/queue/pebble"
	"github.com/prometheus/client_golang/prometheus"
	config_util "github.com/prometheus/common/config"
	"github.com/prometheus/common/model"
	"github.com/prometheus/prometheus/storage"
)

func init() {
	component.Register(component.Registration{
		Name:      "prometheus.remote.qcache",
		Singleton: false,
		Args:      Arguments{},
		Exports:   Exports{},
		Build: func(opts component.Options, args component.Arguments) (component.Component, error) {
			return NewComponent(opts, args.(Arguments))
		},
	})
}

func NewComponent(opts component.Options, args Arguments) (*Queue, error) {
	database, err := newDBStore(args.TTL, path.Join(opts.DataPath, "wal"), opts.Registerer, opts.Logger)
	if err != nil {
		return nil, err
	}
	s := &Queue{
		database: database,
		opts:     opts,
		pool:     pebble.NewByteBufferPool(),
	}

	return s, s.Update(args)
}

// Queue is a queue based WAL used to send data to a remote_write endpoint. Queue supports replaying
// sending and TTLs.
type Queue struct {
	mut        sync.RWMutex
	database   *dbstore
	args       Arguments
	opts       component.Options
	wr         *writer
	testClient WriteClient
	pool       *pebble.ByteBufferPool
}

// Run starts the component, blocking until ctx is canceled or the component
// suffers a fatal error. Run is guaranteed to be called exactly once per
// Component.
func (s *Queue) Run(ctx context.Context) error {
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

func (s *Queue) newQueueManager() (*QueueManager, error) {
	wr, err := s.newWriteClient()
	if err != nil {
		return nil, err
	}
	met := newQueueManagerMetrics(s.opts.Registerer, "", wr.Endpoint())

	qm := NewQueueManager(
		met,
		s.opts.Logger,
		s.args.Endpoint.QueueOptions,
		s.args.Endpoint.MetadataOptions,
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
func (s *Queue) newWriteClient() (WriteClient, error) {
	if s.testClient != nil {
		return s.testClient, nil
	}
	endUrl, err := url.Parse(s.args.Endpoint.URL)
	if err != nil {
		return nil, err
	}
	cfgURL := &config_util.URL{URL: endUrl}
	if err != nil {
		return nil, err
	}

	wr, err := NewWriteClient(s.opts.ID, &ClientConfig{
		URL:              cfgURL,
		Timeout:          model.Duration(s.args.Endpoint.RemoteTimeout),
		HTTPClientConfig: *s.args.Endpoint.HTTPClientConfig.Convert(),
		SigV4Config:      nil,
		Headers:          s.args.Endpoint.Headers,
		RetryOnRateLimit: s.args.Endpoint.QueueOptions.RetryOnHTTP429,
	}, s.opts.Registerer)

	return wr, err
}

// Update provides a new Config to the component. The type of newConfig will
// always match the struct type which the component registers.
//
// Update will be called concurrently with Run. The component must be able to
// gracefully handle updating its config while still running.
//
// An error may be returned if the provided config is invalid.
func (s *Queue) Update(args component.Arguments) error {
	s.mut.Lock()
	defer s.mut.Unlock()

	s.args = args.(Arguments)
	s.opts.OnStateChange(Exports{Receiver: s})

	return nil
}

// Appender returns a new appender for the storage. The implementation
// can choose whether or not to use the context, for deadlines or to check
// for errors.
func (c *Queue) Appender(ctx context.Context) storage.Appender {
	c.mut.RLock()
	defer c.mut.RUnlock()

	return newAppender(c, c.args.TTL)
}

func (c *Queue) commit(a *appender) {
	c.mut.Lock()
	defer c.mut.Unlock()

	var err error
	start := 0
	buf := c.pool.Get()
	defer c.pool.Put(buf)
	mem := arena.NewArena()
	defer mem.Free()
	// Write cache information.
	for _, s := range a.samples {
		s.Hash, err = c.database.WriteHash(s.L)
		if err != nil {
			return
		}
	}
	for {
		end := 0
		if start+c.args.BatchSize >= len(a.samples) {
			end = len(a.samples)
		} else {
			end = start + c.args.BatchSize
		}
		toEnc := a.samples[start:end]
		if len(toEnc) <= 0 {
			return
		}
		err := marshalSamples(toEnc, buf, mem)
		if err != nil {
			return
		}
		key, _ := c.database.WriteSignal(buf.Bytes(), 1, len(toEnc))
		// Update our last seen.
		for _, k := range toEnc {
			c.database.WriteLastSeen(k.Hash, key)
		}
		buf.Reset()
		start = end
		if start == len(a.samples) {
			return
		}
	}
}

func (c *Queue) cleanupDB(ctx context.Context) {
	c.cleanup()
	ttlTimer := time.NewTicker(c.args.Evict)
	for {
		select {
		case <-ttlTimer.C:
			c.cleanup()
		case <-ctx.Done():
			return
		}
	}
}

func (c *Queue) cleanup() {
	level.Info(c.opts.Logger).Log("msg", "starting evict")
	oldestKey := c.wr.GetKey()
	c.database.sampleDB.DeleteKeysOlderThan(oldestKey)
	c.database.evict()
	level.Info(c.opts.Logger).Log("msg", "finishing evict")
}

func (c *Queue) createHashs() {

}
