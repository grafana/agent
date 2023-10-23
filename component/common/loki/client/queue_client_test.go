package client

import (
	"github.com/go-kit/log"
	"github.com/grafana/agent/component/common/loki/utils"
	"github.com/grafana/dskit/backoff"
	"github.com/grafana/dskit/flagext"
	"github.com/grafana/loki/pkg/ingester/wal"
	"github.com/grafana/loki/pkg/logproto"
	lokiflag "github.com/grafana/loki/pkg/util/flagext"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/common/config"
	"github.com/prometheus/prometheus/model/labels"
	"github.com/prometheus/prometheus/tsdb/chunks"
	"github.com/prometheus/prometheus/tsdb/record"
	"github.com/stretchr/testify/require"
	"os"
	"testing"
	"time"
)

func TestQueueClient(t *testing.T) {
	reg := prometheus.NewRegistry()

	// Create a buffer channel where we do enqueue received requests
	receivedReqsChan := make(chan utils.RemoteWriteRequest, 10)

	receivedReqs := utils.NewSyncSlice[utils.RemoteWriteRequest]()
	go func() {
		for req := range receivedReqsChan {
			receivedReqs.Append(req)
		}
	}()

	// Start a local HTTP server
	server := utils.NewRemoteWriteServer(receivedReqsChan, 200)
	require.NotNil(t, server)
	defer server.Close()

	// Get the URL at which the local test server is listening to
	serverURL := flagext.URLValue{}
	err := serverURL.Set(server.URL)
	require.NoError(t, err)

	// Instance the client
	cfg := Config{
		URL:            serverURL,
		BatchWait:      time.Millisecond * 50,
		BatchSize:      10,
		Client:         config.HTTPClientConfig{},
		BackoffConfig:  backoff.Config{MinBackoff: 5 * time.Second, MaxBackoff: 10 * time.Second, MaxRetries: 1},
		ExternalLabels: lokiflag.LabelSet{},
		Timeout:        1 * time.Second,
		TenantID:       "",
	}

	logger := log.NewLogfmtLogger(os.Stdout)

	m := NewMetrics(reg)
	qc, err := NewQueue(m, cfg, 0, 0, false, logger, QueueConfig{
		Capacity:     10,
		DrainTimeout: time.Second,
	})
	require.NoError(t, err)

	//labels := model.LabelSet{"app": "test"}
	lines := []string{
		"hola 1",
		"hola 2",
		"hola 3",
	}

	// Send all the input log entries
	for _, l := range lines {
		qc.StoreSeries([]record.RefSeries{
			{
				Labels: labels.Labels{{
					Name:  "app",
					Value: "test",
				}},
				Ref: chunks.HeadSeriesRef(1),
			},
		}, 0)

		_ = qc.AppendEntries(wal.RefEntries{
			Ref: chunks.HeadSeriesRef(1),
			Entries: []logproto.Entry{{
				Timestamp: time.Now(),
				Line:      l,
			}},
		}, 0)
	}

	require.Eventually(t, func() bool {
		return receivedReqs.Length() >= len(lines)
	}, time.Second*10, time.Second, "timed out waiting for messages to arrive")

	// Stop the client: it waits until the current batch is sent
	qc.Stop()
	close(receivedReqsChan)
}
