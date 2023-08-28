package simple

import (
	"context"
	"net"
	"sync"
	"testing"
	"time"

	"github.com/prometheus/client_golang/prometheus"

	"github.com/gogo/protobuf/proto"
	"github.com/golang/snappy"
	"github.com/grafana/agent/component"
	"github.com/grafana/agent/pkg/flow/logging"
	"github.com/prometheus/prometheus/model/exemplar"
	"github.com/prometheus/prometheus/model/histogram"
	"github.com/prometheus/prometheus/model/labels"
	"github.com/prometheus/prometheus/prompb"
	"github.com/stretchr/testify/require"
)

func TestSimpleMetricTypes(t *testing.T) {
	l := logging.New(nil)
	opts := component.Options{
		ID:       "test",
		Logger:   l,
		DataPath: t.TempDir(),
		OnStateChange: func(e component.Exports) {
		},
		Registerer:     prometheus.DefaultRegisterer,
		Tracer:         nil,
		Clusterer:      nil,
		HTTPListenAddr: "",
		DialFunc: func(ctx context.Context, network string, address string) (net.Conn, error) {
			return nil, nil
		},
		HTTPPath: "",
	}
	args := defaultArgs()
	args.TTL = 0

	s, err := NewComponent(opts, args)
	fk := &fakeWriter{samples: make([]*prompb.WriteRequest, 0)}
	s.testClient = fk
	ctx := context.Background()
	ctx, cancelFunc := context.WithTimeout(ctx, 90*time.Second)
	defer cancelFunc()
	go s.Run(ctx)
	// This feels like a smell but there needs to be a slight delay in calling run and then calling append.
	time.Sleep(2 * time.Second)
	require.NoError(t, err)

	appender := s.Appender(context.Background())
	_, err = appender.Append(0, labels.FromStrings("name", "metric"), time.Now().Unix(), 0)
	require.NoError(t, err)
	_, err = appender.AppendExemplar(0, labels.FromStrings("name", "exemplar"), exemplar.Exemplar{
		Labels: labels.FromStrings("exemplar", "exemplar"),
		Value:  1,
		Ts:     time.Now().Unix(),
		HasTs:  true,
	})
	require.NoError(t, err)
	_, err = appender.AppendHistogram(0, labels.FromStrings("name", "histogram"), time.Now().Unix(), &histogram.Histogram{}, nil)
	require.NoError(t, err)
	_, err = appender.AppendHistogram(0, labels.FromStrings("name", "floathistogram"), time.Now().Unix(), nil, &histogram.FloatHistogram{})
	require.NoError(t, err)

	err = appender.Commit()
	require.NoError(t, err)
	require.Eventuallyf(t, func() bool {
		return fk.count() == 4
	}, 30*time.Second, 1*time.Second, "")

	require.Len(t, fk.samples, 4)
	require.True(t, fk.find("metric"))
	require.True(t, fk.find("exemplar"))
	require.True(t, fk.find("histogram"))
	require.True(t, fk.find("floathistogram"))
}

type fakeWriter struct {
	mut     sync.Mutex
	samples []*prompb.WriteRequest
}

func (fk *fakeWriter) count() int {
	fk.mut.Lock()
	defer fk.mut.Unlock()
	return len(fk.samples)
}

func (fk *fakeWriter) find(name string) bool {
	fk.mut.Lock()
	defer fk.mut.Unlock()

	for _, sample := range fk.samples {
		for _, metric := range sample.Timeseries {
			for _, lbls := range metric.Labels {
				if lbls.GetName() == "name" && lbls.GetValue() == name {
					return true
				}
			}
		}
	}
	return false
}

// Store stores the given samples in the remote storage.
func (fk *fakeWriter) Store(ctx context.Context, buf []byte) error {
	fk.mut.Lock()
	defer fk.mut.Unlock()
	undecode, err := snappy.Decode(nil, buf)
	if err != nil {
		return err
	}
	req := &prompb.WriteRequest{}

	pBuf := proto.NewBuffer(undecode) // For convenience in tests. Not efficient.
	err = pBuf.Unmarshal(req)
	if err != nil {
		return err
	}
	fk.samples = append(fk.samples, req)
	return nil
}

// Name uniquely identifies the remote storage.
func (fk *fakeWriter) Name() string {
	return "fake"
}

// Endpoint is the remote read or write endpoint for the storage client.
func (fk *fakeWriter) Endpoint() string {
	return "127.0.0.1:9999"
}
