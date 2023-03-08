package write

import (
	"context"
	"fmt"
	"math"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/grafana/agent/component/common/loki"
	"github.com/grafana/agent/pkg/flow/componenttest"
	"github.com/grafana/agent/pkg/river"
	"github.com/grafana/agent/pkg/util"
	"github.com/grafana/loki/pkg/logproto"
	loki_util "github.com/grafana/loki/pkg/util"
	"github.com/prometheus/common/model"
	"github.com/stretchr/testify/require"
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

func Test(t *testing.T) {
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
	exports.Receiver <- logEntry
	exports.Receiver <- logEntry

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
