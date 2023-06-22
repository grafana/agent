package badger

import (
	"context"
	"path"
	"sync"
	"time"

	"github.com/grafana/agent/component"
	"github.com/grafana/agent/component/prometheus"
	"github.com/prometheus/prometheus/storage"
)

func init() {
	component.Register(component.Registration{
		Name:      "prometheus.wal.badger",
		Singleton: false,
		Args:      Arguments{},
		Exports:   Exports{},
		Build: func(opts component.Options, args component.Arguments) (component.Component, error) {
			return NewComponent(opts, args.(Arguments))
		},
	})
}

type Component struct {
	mut     sync.RWMutex
	args    Arguments
	opts    component.Options
	metrics *db
}

var _ component.Component = (*Component)(nil)

func NewComponent(o component.Options, c Arguments) (*Component, error) {
	db, err := newDb(path.Join(o.DataPath, "metrics"), o.Logger)
	if err != nil {
		return nil, err
	}

	return &Component{
		args:    c,
		opts:    o,
		metrics: db,
	}, nil
}

func (c *Component) Run(ctx context.Context) error {
	ticker := time.NewTicker(1 * time.Second)
	for {
		select {
		case <-ctx.Done():
			return nil
		case <-ticker.C:

			// TODO check to see if anything needs to be written
		}
	}
}

func (c *Component) Update(args component.Arguments) error {
	c.args = args.(Arguments)
	c.opts.OnStateChange(Exports{Receiver: c})
	return nil
}

// Appender returns a new appender for the storage. The implementation
// can choose whether or not to use the context, for deadlines or to check
// for errors.
func (c *Component) Appender(ctx context.Context) storage.Appender {
	return newAppender(c)
}

func (c *Component) commit(a *appender) {
	c.mut.RLock()
	defer c.mut.Unlock()

	endpoint := time.Now().UnixMilli() - int64(c.args.TTL.Seconds())

	timestampedMetrics := make([]any, 0)
	for _, x := range a.metrics {
		// No need to write if already outside of range.
		if x.Timestamp < endpoint {
			continue
		}
		timestampedMetrics = append(timestampedMetrics, x)
	}

	c.metrics.writeRecords(timestampedMetrics, c.args.TTL)
}

type seqSettable interface {
	SetSeq(uint64)
}
type Arguments struct {
	TTL       time.Duration        `river:"ttl,attr,optional"`
	ForwardTo []prometheus.WriteTo `river:"forward_to,attr"`
}

type Exports struct {
	Receiver storage.Appendable `river:"receiver,attr"`
}
