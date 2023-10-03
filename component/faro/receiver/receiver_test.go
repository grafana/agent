package receiver

import (
	"fmt"
	"net/http"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/grafana/agent/component/common/loki"
	"github.com/grafana/agent/component/otelcol"
	"github.com/grafana/agent/pkg/flow/componenttest"
	"github.com/grafana/agent/pkg/util"
	"github.com/grafana/loki/pkg/logproto"
	"github.com/phayes/freeport"
	"github.com/prometheus/common/model"
	"github.com/stretchr/testify/require"
)

// Test performs an end-to-end test of the component.
func Test(t *testing.T) {
	ctx := componenttest.TestContext(t)

	ctrl, err := componenttest.NewControllerFromID(
		util.TestLogger(t),
		"faro.receiver",
	)
	require.NoError(t, err)

	freePort, err := freeport.GetFreePort()
	require.NoError(t, err)

	lr := newFakeLogsReceiver(t)

	go func() {
		err := ctrl.Run(ctx, Arguments{
			LogLabels: map[string]string{
				"foo": "bar",
			},

			Server: ServerArguments{
				Host: "127.0.0.1",
				Port: freePort,
			},

			Output: OutputArguments{
				Logs:   []loki.LogsReceiver{lr},
				Traces: []otelcol.Consumer{},
			},
		})
		require.NoError(t, err)
	}()

	// Wait for the server to be running.
	util.Eventually(t, func(t require.TestingT) {
		resp, err := http.Get(fmt.Sprintf("http://localhost:%d/-/ready", freePort))
		require.NoError(t, err)
		defer resp.Body.Close()

		require.Equal(t, http.StatusOK, resp.StatusCode)
	})

	// Send a sample payload to the server.
	resp, err := http.Post(
		fmt.Sprintf("http://localhost:%d/collect", freePort),
		"application/json",
		strings.NewReader(`{
			"traces": {
				"resourceSpans": []
			},
			"logs": [{
				"message": "hello, world",
				"level": "info",
				"context": {"env": "dev"},
				"timestamp": "2021-01-01T00:00:00Z",
				"trace": {
					"trace_id": "0",
					"span_id": "0"
				}
			}],
			"exceptions": [],
			"measurements": [],
			"meta": {}
		}`),
	)
	require.NoError(t, err)
	defer resp.Body.Close()

	require.Equal(t, http.StatusAccepted, resp.StatusCode)
	require.Len(t, lr.GetEntries(), 1)

	expect := loki.Entry{
		Labels: model.LabelSet{
			"foo": model.LabelValue("bar"),
		},
		Entry: logproto.Entry{
			Line: `timestamp="2021-01-01 00:00:00 +0000 UTC" kind=log message="hello, world" level=info context_env=dev traceID=0 spanID=0 browser_mobile=false`,
		},
	}
	require.Equal(t, expect, lr.entries[0])
}

type fakeLogsReceiver struct {
	ch chan loki.Entry

	entriesMut sync.RWMutex
	entries    []loki.Entry
}

var _ loki.LogsReceiver = (*fakeLogsReceiver)(nil)

func newFakeLogsReceiver(t *testing.T) *fakeLogsReceiver {
	ctx := componenttest.TestContext(t)

	lr := &fakeLogsReceiver{
		ch: make(chan loki.Entry, 1),
	}

	go func() {
		defer close(lr.ch)

		select {
		case <-ctx.Done():
			return
		case ent := <-lr.Chan():

			lr.entriesMut.Lock()
			lr.entries = append(lr.entries, loki.Entry{
				Labels: ent.Labels,
				Entry: logproto.Entry{
					Timestamp:          time.Time{}, // Use consistent time for testing.
					Line:               ent.Line,
					StructuredMetadata: ent.StructuredMetadata,
				},
			})
			lr.entriesMut.Unlock()
		}
	}()

	return lr
}

func (lr *fakeLogsReceiver) Chan() chan loki.Entry {
	return lr.ch
}

func (lr *fakeLogsReceiver) GetEntries() []loki.Entry {
	lr.entriesMut.RLock()
	defer lr.entriesMut.RUnlock()
	return lr.entries
}
