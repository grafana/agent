package memory

import (
	"bytes"
	"context"
	"encoding/gob"
	"strconv"
	"sync"
	"time"

	"github.com/go-kit/kit/log/level"
	"github.com/grafana/agent/component"
	"github.com/grafana/agent/component/prometheus"
	"github.com/nutsdb/nutsdb/inmemory"
	"github.com/prometheus/prometheus/storage"
)

func init() {
	component.Register(component.Registration{
		Name:      "prometheus.wal.memory",
		Singleton: false,
		Args:      Arguments{},
		Exports:   Exports{},
		Build: func(opts component.Options, args component.Arguments) (component.Component, error) {
			return NewComponent(opts, args.(Arguments))
		},
	})
}

type Component struct {
	mut     sync.Mutex
	args    Arguments
	opts    component.Options
	db      *inmemory.DB
	encoder *gob.Encoder
	encBuff *bytes.Buffer
	decode  *gob.Decoder
	decBuff *bytes.Buffer
}

var _ component.Component = (*Component)(nil)

func NewComponent(o component.Options, c Arguments) (*Component, error) {
	db, err := inmemory.Open(inmemory.DefaultOptions)
	encbuff := bytes.NewBuffer([]byte{})
	decBuff := bytes.NewBuffer([]byte{})
	if err != nil {
		return nil, err
	}
	return &Component{
		args:    c,
		db:      db,
		encoder: gob.NewEncoder(encbuff),
		decode:  gob.NewDecoder(decBuff),
		opts:    o,
		encBuff: encbuff,
		decBuff: decBuff,
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

func (c *Component) hasData() bool {
	k, err := c.db.AllKeys("metrics")
	if err != nil {
		return true
	}
	if len(k) > 0 {
		return true
	}
	return false
}

func (c *Component) commit(a *appender) {
	c.mut.Lock()
	defer c.mut.Unlock()

	endpoint := time.Now().UnixMilli() - int64(c.args.TTL.Seconds())

	for _, x := range a.metrics {
		// No need to write if already outside of range.
		if x.Timestamp < endpoint {
			continue
		}
		byteTS := []byte(strconv.FormatInt(x.Timestamp, 10))
		err := c.encoder.Encode(x)
		if err != nil {
			level.Error(c.opts.Logger).Log("err", err)
			continue
		}
		entry, err := c.db.Get("metrics", byteTS)
		if err != nil {
			level.Error(c.opts.Logger).Log("err", err)
			continue
		}
		// We need to set the TTL for this bucket.
		if entry == nil {
			c.db.Put("metrics", byteTS, nil, uint32(c.args.TTL.Seconds()))
		} else {
			err = c.db.RPush("metrics", string(byteTS), c.encBuff.Bytes())
			if err != nil {
				level.Error(c.opts.Logger).Log("err", err)
				continue
			}
		}
	}
}

type Arguments struct {
	TTL       time.Duration        `river:"ttl,attr,optional"`
	ForwardTo []prometheus.WriteTo `river:"forward_to,attr"`
}

type Exports struct {
	Receiver storage.Appendable `river:"receiver,attr"`
}
