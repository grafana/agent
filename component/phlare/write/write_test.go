package write

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"
	"time"

	"github.com/bufbuild/connect-go"
	"github.com/grafana/agent/component"
	pushv1 "github.com/grafana/agent/component/phlare/push/v1"
	"github.com/grafana/agent/component/phlare/push/v1/pushv1connect"
	"github.com/grafana/agent/component/pprof"
	"github.com/grafana/agent/pkg/river"
	"github.com/grafana/agent/pkg/util"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/prometheus/model/labels"
	"github.com/stretchr/testify/require"
	"go.uber.org/atomic"
)

type PushFunc func(context.Context, *connect.Request[pushv1.PushRequest]) (*connect.Response[pushv1.PushResponse], error)

func (p PushFunc) Push(ctx context.Context, r *connect.Request[pushv1.PushRequest]) (*connect.Response[pushv1.PushResponse], error) {
	return p(ctx, r)
}

func Test_Write_FanOut(t *testing.T) {
	var (
		export      Exports
		argument                       = DefaultArguments()
		pushTotal                      = atomic.NewInt32(0)
		serverCount                    = int32(10)
		servers     []*httptest.Server = make([]*httptest.Server, serverCount)
	)
	argument.ExternalLabels = map[string]string{"foo": "buzz"}
	handlerFn := func(err error) http.Handler {
		_, handler := pushv1connect.NewPusherServiceHandler(PushFunc(
			func(_ context.Context, req *connect.Request[pushv1.PushRequest]) (*connect.Response[pushv1.PushResponse], error) {
				pushTotal.Inc()
				require.Equal(t, "test", req.Header()["X-Test-Header"][0])
				require.Contains(t, req.Header()["User-Agent"][0], "GrafanaAgent/")
				require.Equal(t, []*pushv1.LabelPair{
					{Name: "__name__", Value: "test"},
					{Name: "foo", Value: "buzz"},
					{Name: "job", Value: "foo"},
				}, req.Msg.Series[0].Labels)
				require.Equal(t, []byte("pprofraw"), req.Msg.Series[0].Samples[0].RawProfile)
				return &connect.Response[pushv1.PushResponse]{}, err
			},
		))
		return handler
	}

	for i := int32(0); i < serverCount; i++ {
		if i == 0 {
			servers[i] = httptest.NewServer(handlerFn(errors.New("test")))
		} else {
			servers[i] = httptest.NewServer(handlerFn(nil))
		}
		argument.Endpoints = append(argument.Endpoints, &EndpointOptions{
			URL:           servers[i].URL,
			RemoteTimeout: DefaultEndpointOptions().RemoteTimeout,
			Headers: map[string]string{
				"X-Test-Header": "test",
			},
		})
	}
	defer func() {
		for _, s := range servers {
			s.Close()
		}
	}()
	createReceiver := func(t *testing.T, arg Arguments) pprof.Appendable {
		var wg sync.WaitGroup
		wg.Add(1)
		c, err := NewComponent(component.Options{
			ID:     "1",
			Logger: util.TestLogger(t),
			OnStateChange: func(e component.Exports) {
				defer wg.Done()
				export = e.(Exports)
			},
			Registerer: prometheus.NewRegistry(),
		}, arg)
		require.NoError(t, err)
		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()
		go c.Run(ctx)
		wg.Wait() // wait for the state change to happen
		require.NotNil(t, export.Receiver)
		return export.Receiver
	}

	t.Run("with_failure", func(t *testing.T) {
		r := createReceiver(t, argument)
		pushTotal.Store(0)
		err := r.Appender().Append(context.Background(), labels.FromMap(map[string]string{
			"__name__": "test",
			"__type__": "type",
			"job":      "foo",
			"foo":      "bar",
		}), []*pprof.RawSample{
			{RawProfile: []byte("pprofraw")},
		})
		require.EqualErrorf(t, err, "unknown: test", "expected error to be test")
		require.Equal(t, serverCount, pushTotal.Load())
	})

	t.Run("all_success", func(t *testing.T) {
		argument.Endpoints = argument.Endpoints[1:]
		r := createReceiver(t, argument)
		pushTotal.Store(0)
		err := r.Appender().Append(context.Background(), labels.FromMap(map[string]string{
			"__name__": "test",
			"__type__": "type",
			"job":      "foo",
			"foo":      "bar",
		}), []*pprof.RawSample{
			{RawProfile: []byte("pprofraw")},
		})
		require.NoError(t, err)
		require.Equal(t, serverCount-1, pushTotal.Load())
	})
}

func Test_Write_Update(t *testing.T) {
	var (
		export    Exports
		argument  = DefaultArguments()
		pushTotal = atomic.NewInt32(0)
	)
	var wg sync.WaitGroup
	wg.Add(1)
	c, err := NewComponent(component.Options{
		ID:     "1",
		Logger: util.TestLogger(t),
		OnStateChange: func(e component.Exports) {
			defer wg.Done()
			export = e.(Exports)
		},
		Registerer: prometheus.NewRegistry(),
	}, argument)
	require.NoError(t, err)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	go c.Run(ctx)
	wg.Wait() // wait for the state change to happen
	require.NotNil(t, export.Receiver)
	// First one is a noop
	err = export.Receiver.Appender().Append(context.Background(), labels.FromMap(map[string]string{
		"__name__": "test",
	}), []*pprof.RawSample{
		{RawProfile: []byte("pprofraw")},
	})
	require.NoError(t, err)

	_, handler := pushv1connect.NewPusherServiceHandler(PushFunc(
		func(_ context.Context, req *connect.Request[pushv1.PushRequest]) (*connect.Response[pushv1.PushResponse], error) {
			pushTotal.Inc()
			return &connect.Response[pushv1.PushResponse]{}, err
		},
	))
	server := httptest.NewServer(handler)
	defer server.Close()
	argument.Endpoints = []*EndpointOptions{
		{
			URL:           server.URL,
			RemoteTimeout: DefaultEndpointOptions().RemoteTimeout,
		},
	}
	wg.Add(1)
	require.NoError(t, c.Update(argument))
	wg.Wait()
	err = export.Receiver.Appender().Append(context.Background(), labels.FromMap(map[string]string{
		"__name__": "test",
	}), []*pprof.RawSample{
		{RawProfile: []byte("pprofraw")},
	})
	require.NoError(t, err)
	require.Equal(t, int32(1), pushTotal.Load())
}

func Test_Unmarshal_Config(t *testing.T) {
	var arg Arguments
	river.Unmarshal([]byte(`
	endpoint {
		url = "http://localhost:4100"
		remote_timeout = "10s"
	}
	endpoint {
		url = "http://localhost:4200"
		remote_timeout = "5s"
	}
	external_labels = {
		"foo" = "bar",
	}`), &arg)
	require.Equal(t, "http://localhost:4100", arg.Endpoints[0].URL)
	require.Equal(t, "http://localhost:4200", arg.Endpoints[1].URL)
	require.Equal(t, time.Second*10, arg.Endpoints[0].RemoteTimeout)
	require.Equal(t, time.Second*5, arg.Endpoints[1].RemoteTimeout)
	require.Equal(t, "bar", arg.ExternalLabels["foo"])
}
