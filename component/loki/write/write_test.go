package write

import (
	"context"
	"fmt"
	"math"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"

	"github.com/prometheus/common/model"
	"github.com/stretchr/testify/require"

	"github.com/grafana/agent/component/common/loki"
	"github.com/grafana/agent/component/common/loki/wal"
	"github.com/grafana/agent/component/discovery"
	lsf "github.com/grafana/agent/component/loki/source/file"
	"github.com/grafana/agent/pkg/flow/componenttest"
	"github.com/grafana/agent/pkg/util"
	"github.com/grafana/loki/pkg/logproto"
	loki_util "github.com/grafana/loki/pkg/util"
	"github.com/grafana/river"
)

func TestRiverConfig(t *testing.T) {
	var exampleRiverConfig = `
	endpoint {
		name           = "test-url"
		url            = "http://0.0.0.0:11111/loki/api/v1/push"
		remote_timeout = "100ms"
	}
`

	var args Arguments
	err := river.Unmarshal([]byte(exampleRiverConfig), &args)
	require.NoError(t, err)
}

func TestBadRiverConfig(t *testing.T) {
	var exampleRiverConfig = `
	endpoint {
		name           = "test-url"
		url            = "http://0.0.0.0:11111/loki/api/v1/push"
		remote_timeout = "100ms"
		bearer_token = "token"
		bearer_token_file = "/path/to/file.token"
	}
`

	// Make sure the squashed HTTPClientConfig Validate function is being utilized correctly
	var args Arguments
	err := river.Unmarshal([]byte(exampleRiverConfig), &args)
	require.ErrorContains(t, err, "at most one of bearer_token & bearer_token_file must be configured")
}

func TestUnmarshallWalAttrributes(t *testing.T) {
	type testcase struct {
		raw           string
		errorExpected bool
		expected      WalArguments
	}

	for name, tc := range map[string]testcase{
		"min read frequency higher than max": {
			raw: `
			enabled = true
			min_read_frequency = "1h"
			max_read_frequency = "1m"
			`,
			errorExpected: true,
		},
		"default config is wal disabled": {
			raw: "",
			expected: WalArguments{
				Enabled:          false,
				MaxSegmentAge:    wal.DefaultMaxSegmentAge,
				MinReadFrequency: wal.DefaultWatchConfig.MinReadFrequency,
				MaxReadFrequency: wal.DefaultWatchConfig.MaxReadFrequency,
			},
		},
		"wal enabled with defaults": {
			raw: `
			enabled = true
			`,
			expected: WalArguments{
				Enabled:          true,
				MaxSegmentAge:    wal.DefaultMaxSegmentAge,
				MinReadFrequency: wal.DefaultWatchConfig.MinReadFrequency,
				MaxReadFrequency: wal.DefaultWatchConfig.MaxReadFrequency,
			},
		},
		"wal enabled with some overrides": {
			raw: `
			enabled = true
			max_segment_age = "10m"
			min_read_frequency = "11ms"
			`,
			expected: WalArguments{
				Enabled:          true,
				MaxSegmentAge:    time.Minute * 10,
				MinReadFrequency: time.Millisecond * 11,
				MaxReadFrequency: wal.DefaultWatchConfig.MaxReadFrequency,
			},
		},
	} {
		t.Run(name, func(t *testing.T) {
			cfg := WalArguments{}
			err := river.Unmarshal([]byte(tc.raw), &cfg)
			if tc.errorExpected {
				require.Error(t, err)
				return
			}
			require.NoError(t, err)
			require.Equal(t, tc.expected, cfg)
		})
	}
}

func TestWriteToSingleEndpoint(t *testing.T) {
	t.Run("wal disabled", func(t *testing.T) {
		testSingleEndpoint(t, func(args *Arguments) {})
	})

	t.Run("wal enabled", func(t *testing.T) {
		testSingleEndpoint(t, func(args *Arguments) {
			args.WAL.Enabled = true
		})
	})
}

func testSingleEndpoint(t *testing.T, alterConfig func(arguments *Arguments)) {
	// Set up the server that will receive the log entry, and expose it on ch.
	ch := make(chan logproto.PushRequest)
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var pushReq logproto.PushRequest
		err := loki_util.ParseProtoReader(context.Background(), r.Body, int(r.ContentLength), math.MaxInt32, &pushReq, loki_util.RawSnappy)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		tenantHeader := r.Header.Get("X-Scope-OrgID")
		require.Equal(t, tenantHeader, "tenant-1")

		ch <- pushReq
	}))
	defer srv.Close()

	// Set up the component Arguments.
	cfg := fmt.Sprintf(`
		endpoint {
			url        = "%s"
			batch_wait = "10ms"
			tenant_id  = "tenant-1"
		}
	`, srv.URL)
	var args Arguments
	require.NoError(t, river.Unmarshal([]byte(cfg), &args))

	alterConfig(&args)

	// Set up and start the component.
	tc, err := componenttest.NewControllerFromID(util.TestLogger(t), "loki.write")
	require.NoError(t, err)
	go func() {
		err = tc.Run(componenttest.TestContext(t), args)
		require.NoError(t, err)
	}()
	require.NoError(t, tc.WaitExports(time.Second))

	// Send two log entries to the component's receiver
	logEntry := loki.Entry{
		Labels: model.LabelSet{"foo": "bar"},
		Entry: logproto.Entry{
			Timestamp: time.Now(),
			Line:      "very important log",
		},
	}

	exports := tc.Exports().(Exports)
	exports.Receiver.Chan() <- logEntry
	exports.Receiver.Chan() <- logEntry

	// Wait for our exporter to finish and pass data to our HTTP server.
	// Make sure the log entries were received correctly.
	select {
	case <-time.After(2 * time.Second):
		require.FailNow(t, "failed waiting for logs")
	case req := <-ch:
		require.Len(t, req.Streams, 1)
		require.Equal(t, req.Streams[0].Labels, logEntry.Labels.String())
		require.Len(t, req.Streams[0].Entries, 2)
		require.Equal(t, req.Streams[0].Entries[0].Line, logEntry.Line)
		require.Equal(t, req.Streams[0].Entries[1].Line, logEntry.Line)
	}
}

func TestEntrySentToTwoWriteComponents(t *testing.T) {
	t.Run("wal disabled", func(t *testing.T) {
		testMultipleEndpoint(t, func(arguments *Arguments) {})
	})

	t.Run("wal enabled", func(t *testing.T) {
		testMultipleEndpoint(t, func(arguments *Arguments) {
			arguments.WAL.Enabled = true
		})
	})
}

func testMultipleEndpoint(t *testing.T, alterArgs func(arguments *Arguments)) {
	ch1, ch2 := make(chan logproto.PushRequest), make(chan logproto.PushRequest)
	srv1 := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var pushReq logproto.PushRequest
		require.NoError(t, loki_util.ParseProtoReader(context.Background(), r.Body, int(r.ContentLength), math.MaxInt32, &pushReq, loki_util.RawSnappy))
		ch1 <- pushReq
	}))
	srv2 := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var pushReq logproto.PushRequest
		require.NoError(t, loki_util.ParseProtoReader(context.Background(), r.Body, int(r.ContentLength), math.MaxInt32, &pushReq, loki_util.RawSnappy))
		ch2 <- pushReq
	}))
	defer srv1.Close()
	defer srv2.Close()

	// Set up two different loki.write components.
	cfg1 := fmt.Sprintf(`
		endpoint {
			url        = "%s"
		}
		external_labels = { "lbl" = "foo" }
	`, srv1.URL)
	cfg2 := fmt.Sprintf(`
		endpoint {
			url        = "%s"
		}
		external_labels = { "lbl" = "bar" }
	`, srv2.URL)
	var args1, args2 Arguments
	require.NoError(t, river.Unmarshal([]byte(cfg1), &args1))
	require.NoError(t, river.Unmarshal([]byte(cfg2), &args2))
	alterArgs(&args1)
	alterArgs(&args2)

	// Set up and start the component.
	tc1, err := componenttest.NewControllerFromID(util.TestLogger(t), "loki.write")
	require.NoError(t, err)
	tc2, err := componenttest.NewControllerFromID(util.TestLogger(t), "loki.write")
	require.NoError(t, err)
	go func() {
		require.NoError(t, tc1.Run(componenttest.TestContext(t), args1))
	}()
	go func() {
		require.NoError(t, tc2.Run(componenttest.TestContext(t), args2))
	}()
	require.NoError(t, tc1.WaitExports(time.Second))
	require.NoError(t, tc2.WaitExports(time.Second))

	// Create a file to log to.
	f, err := os.CreateTemp(t.TempDir(), "example")
	require.NoError(t, err)
	defer f.Close()

	// Create and start a component that will read from that file and fan out to both components.
	ctrl, err := componenttest.NewControllerFromID(util.TestLogger(t), "loki.source.file")
	require.NoError(t, err)

	go func() {
		err := ctrl.Run(context.Background(), lsf.Arguments{
			Targets: []discovery.Target{{"__path__": f.Name(), "somelbl": "somevalue"}},
			ForwardTo: []loki.LogsReceiver{
				tc1.Exports().(Exports).Receiver,
				tc2.Exports().(Exports).Receiver,
			},
		})
		require.NoError(t, err)
	}()
	ctrl.WaitRunning(time.Minute)

	// Write a line to the file.
	_, err = f.Write([]byte("writing some text\n"))
	require.NoError(t, err)

	wantLabelSet := model.LabelSet{
		"filename": model.LabelValue(f.Name()),
		"somelbl":  "somevalue",
	}

	// The two entries have been received with their
	for i := 0; i < 2; i++ {
		select {
		case <-time.After(2 * time.Second):
			require.FailNow(t, "failed waiting for logs")
		case req := <-ch1:
			require.Len(t, req.Streams, 1)
			require.Equal(t, req.Streams[0].Labels, wantLabelSet.Clone().Merge(model.LabelSet{"lbl": "foo"}).String())
			require.Len(t, req.Streams[0].Entries, 1)
			require.Equal(t, req.Streams[0].Entries[0].Line, "writing some text")
		case req := <-ch2:
			require.Len(t, req.Streams, 1)
			require.Equal(t, req.Streams[0].Labels, wantLabelSet.Clone().Merge(model.LabelSet{"lbl": "bar"}).String())
			require.Len(t, req.Streams[0].Entries, 1)
			require.Equal(t, req.Streams[0].Entries[0].Line, "writing some text")
		}
	}
}
