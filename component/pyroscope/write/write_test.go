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
	"github.com/grafana/agent/component/pyroscope"
	"github.com/grafana/agent/pkg/util"
	pushv1 "github.com/grafana/pyroscope/api/gen/proto/go/push/v1"
	"github.com/grafana/pyroscope/api/gen/proto/go/push/v1/pushv1connect"
	typesv1 "github.com/grafana/pyroscope/api/gen/proto/go/types/v1"
	"github.com/grafana/river"
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
		endpoints   []*EndpointOptions = make([]*EndpointOptions, 0, serverCount)
	)
	argument.ExternalLabels = map[string]string{"foo": "buzz"}
	handlerFn := func(err error) http.Handler {
		_, handler := pushv1connect.NewPusherServiceHandler(PushFunc(
			func(_ context.Context, req *connect.Request[pushv1.PushRequest]) (*connect.Response[pushv1.PushResponse], error) {
				pushTotal.Inc()
				require.Equal(t, "test", req.Header()["X-Test-Header"][0])
				require.Contains(t, req.Header()["User-Agent"][0], "GrafanaAgent/")
				require.Equal(t, []*typesv1.LabelPair{
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
		endpoints = append(endpoints, &EndpointOptions{
			URL:               servers[i].URL,
			MinBackoff:        100 * time.Millisecond,
			MaxBackoff:        200 * time.Millisecond,
			MaxBackoffRetries: 1,
			RemoteTimeout:     GetDefaultEndpointOptions().RemoteTimeout,
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
	createReceiver := func(t *testing.T, arg Arguments) pyroscope.Appendable {
		t.Helper()
		var wg sync.WaitGroup
		wg.Add(1)
		c, err := New(component.Options{
			ID:         "1",
			Logger:     util.TestFlowLogger(t),
			Registerer: prometheus.NewRegistry(),
			OnStateChange: func(e component.Exports) {
				defer wg.Done()
				export = e.(Exports)
			},
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
		argument.Endpoints = endpoints
		r := createReceiver(t, argument)
		pushTotal.Store(0)
		err := r.Appender().Append(context.Background(), labels.FromMap(map[string]string{
			"__name__": "test",
			"__type__": "type",
			"job":      "foo",
			"foo":      "bar",
		}), []*pyroscope.RawSample{
			{RawProfile: []byte("pprofraw")},
		})
		require.EqualErrorf(t, err, "unknown: test", "expected error to be test")
		require.Equal(t, serverCount, pushTotal.Load())
	})

	t.Run("all_success", func(t *testing.T) {
		argument.Endpoints = endpoints[1:]
		r := createReceiver(t, argument)
		pushTotal.Store(0)
		err := r.Appender().Append(context.Background(), labels.FromMap(map[string]string{
			"__name__": "test",
			"__type__": "type",
			"job":      "foo",
			"foo":      "bar",
		}), []*pyroscope.RawSample{
			{RawProfile: []byte("pprofraw")},
		})
		require.NoError(t, err)
		require.Equal(t, serverCount-1, pushTotal.Load())
	})

	t.Run("with_backoff", func(t *testing.T) {
		argument.Endpoints = endpoints[:1]
		argument.Endpoints[0].MaxBackoffRetries = 3
		r := createReceiver(t, argument)
		pushTotal.Store(0)
		err := r.Appender().Append(context.Background(), labels.FromMap(map[string]string{
			"__name__": "test",
			"__type__": "type",
			"job":      "foo",
			"foo":      "bar",
		}), []*pyroscope.RawSample{
			{RawProfile: []byte("pprofraw")},
		})
		require.Error(t, err)
		require.Equal(t, int32(3), pushTotal.Load())
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
	c, err := New(component.Options{
		ID:         "1",
		Logger:     util.TestFlowLogger(t),
		Registerer: prometheus.NewRegistry(),
		OnStateChange: func(e component.Exports) {
			defer wg.Done()
			export = e.(Exports)
		},
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
	}), []*pyroscope.RawSample{
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
			RemoteTimeout: GetDefaultEndpointOptions().RemoteTimeout,
		},
	}
	wg.Add(1)
	require.NoError(t, c.Update(argument))
	wg.Wait()
	err = export.Receiver.Appender().Append(context.Background(), labels.FromMap(map[string]string{
		"__name__": "test",
	}), []*pyroscope.RawSample{
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
		min_backoff_period = "1s"
		max_backoff_period = "10s"
		max_backoff_retries = 10
	}
	external_labels = {
		"foo" = "bar",
	}`), &arg)
	require.Equal(t, "http://localhost:4100", arg.Endpoints[0].URL)
	require.Equal(t, "http://localhost:4200", arg.Endpoints[1].URL)
	require.Equal(t, time.Second*10, arg.Endpoints[0].RemoteTimeout)
	require.Equal(t, time.Second*5, arg.Endpoints[1].RemoteTimeout)
	require.Equal(t, "bar", arg.ExternalLabels["foo"])
	require.Equal(t, time.Second, arg.Endpoints[1].MinBackoff)
	require.Equal(t, time.Second*10, arg.Endpoints[1].MaxBackoff)
	require.Equal(t, 10, arg.Endpoints[1].MaxBackoffRetries)
}

func TestBadRiverConfig(t *testing.T) {
	exampleRiverConfig := `
	endpoint {
		url = "http://localhost:4100"
		remote_timeout = "10s"
		bearer_token = "token"
		bearer_token_file = "/path/to/file.token"
	}
	external_labels = {
		"foo" = "bar",
	}
`

	// Make sure the squashed HTTPClientConfig Validate function is being utilized correctly
	var args Arguments
	err := river.Unmarshal([]byte(exampleRiverConfig), &args)
	require.ErrorContains(t, err, "at most one of bearer_token & bearer_token_file must be configured")
}
