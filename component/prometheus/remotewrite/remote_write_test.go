package remotewrite_test

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/grafana/agent/component/prometheus/remotewrite"
	"github.com/grafana/agent/pkg/flow/componenttest"
	"github.com/grafana/agent/pkg/river"
	"github.com/grafana/agent/pkg/util"
	"github.com/prometheus/prometheus/model/labels"
	"github.com/prometheus/prometheus/prompb"
	"github.com/prometheus/prometheus/storage/remote"
	"github.com/stretchr/testify/require"
)

// Test is an integration-level test which ensures that metrics can get sent to
// a prometheus.remote_write component and forwarded to a
// remote_write-compatible server.
func Test(t *testing.T) {
	writeResult := make(chan *prompb.WriteRequest)

	// Create a remote_write server which forwards any received payloads to the
	// writeResult channel.
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		req, err := remote.DecodeWriteRequest(r.Body)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		select {
		case writeResult <- req:
		default:
			require.Fail(t, "failed to send remote_write result over channel")
		}
	}))
	defer srv.Close()

	// Configure our component to remote_write to the server we just created. We
	// configure batch_send_deadline to 100ms so this test executes fairly
	// quickly.
	cfg := fmt.Sprintf(`
		external_labels = {
			cluster = "local",
		}

		endpoint {
			name           = "test-url"
			url            = "%s/api/v1/write"
			remote_timeout = "100ms"

			queue_config {
				batch_send_deadline = "100ms"
			}
		}
	`, srv.URL)

	var args remotewrite.Arguments
	require.NoError(t, river.Unmarshal([]byte(cfg), &args))

	// Create our component and wait for it to start running so we can write
	// metrics to the WAL.
	tc, err := componenttest.NewControllerFromID(util.TestLogger(t), "prometheus.remote_write")
	require.NoError(t, err)
	go func() {
		err = tc.Run(componenttest.TestContext(t), args)
		require.NoError(t, err)
	}()
	require.NoError(t, tc.WaitRunning(time.Second))

	// We need to use a future timestamp since remote_write will ignore any
	// sample which is earlier than the time when it started. Adding a minute
	// ensures that our samples will never get ignored.
	sampleTimestamp := time.Now().Add(time.Minute).UnixMilli()

	// Send metrics to our component. These will be written to the WAL and
	// subsequently written to our HTTP server.
	rwExports := tc.Exports().(remotewrite.Exports)
	appender := rwExports.Receiver.Appender(context.Background())
	_, err = appender.Append(0, labels.FromStrings("foo", "bar"), sampleTimestamp, 12)
	require.NoError(t, err)
	_, err = appender.Append(0, labels.FromStrings("fizz", "buzz"), sampleTimestamp, 34)
	require.NoError(t, err)
	err = appender.Commit()
	require.NoError(t, err)

	expect := []prompb.TimeSeries{{
		Labels: []prompb.Label{
			{Name: "cluster", Value: "local"},
			{Name: "foo", Value: "bar"},
		},
		Samples: []prompb.Sample{
			{Timestamp: sampleTimestamp, Value: 12},
		},
	}, {
		Labels: []prompb.Label{
			{Name: "cluster", Value: "local"},
			{Name: "fizz", Value: "buzz"},
		},
		Samples: []prompb.Sample{
			{Timestamp: sampleTimestamp, Value: 34},
		},
	}}

	select {
	case <-time.After(time.Minute):
		require.FailNow(t, "timed out waiting for metrics")
	case res := <-writeResult:
		require.Equal(t, expect, res.Timeseries)
	}
}
