package client

import (
	"fmt"
	"os"
	"sync/atomic"
	"testing"
	"time"

	"github.com/alecthomas/units"
	"github.com/go-kit/log"
	"github.com/grafana/agent/component/common/loki/utils"
	"github.com/grafana/dskit/backoff"
	"github.com/grafana/dskit/flagext"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/common/config"
	"github.com/prometheus/prometheus/model/labels"
	"github.com/prometheus/prometheus/tsdb/chunks"
	"github.com/prometheus/prometheus/tsdb/record"
	"github.com/stretchr/testify/require"

	"github.com/grafana/loki/pkg/ingester/wal"
	"github.com/grafana/loki/pkg/logproto"
	lokiflag "github.com/grafana/loki/pkg/util/flagext"
)

type testCase struct {
	// numLines is the total number of lines sent through the client in the benchmark.
	numLines int

	// numSeries is the different number of series to use in entries. Series are dynamically generated for each entry, but
	// would be numSeries in total, and evenly distributed.
	numSeries int

	// configs
	batchSize   int
	batchWait   time.Duration
	queueConfig QueueConfig

	// expects
	expectedRWReqsCount int64
}

func TestQueueClient(t *testing.T) {
	for name, tc := range map[string]testCase{
		"small test": {
			numLines:  3,
			numSeries: 1,
			batchSize: 10,
			batchWait: time.Millisecond * 50,
			queueConfig: QueueConfig{
				Capacity:     100,
				DrainTimeout: time.Second,
			},
		},
		"many lines and series, immediate delivery": {
			numLines:  1000,
			numSeries: 10,
			batchSize: 10,
			batchWait: time.Millisecond * 50,
			queueConfig: QueueConfig{
				Capacity:     100,
				DrainTimeout: time.Second,
			},
		},
		"many lines and series, delivery because of batch age": {
			numLines:  100,
			numSeries: 10,
			batchSize: int(1 * units.MiB), // make batch size big enough so that all batches should be delivered because of batch age
			batchWait: time.Millisecond * 50,
			queueConfig: QueueConfig{
				Capacity:     int(100 * units.MiB), // keep buffered channel size on 100
				DrainTimeout: 10 * time.Second,
			},
			expectedRWReqsCount: 1,
		},
	} {
		t.Run(name, func(t *testing.T) {
			reg := prometheus.NewRegistry()

			// Create a buffer channel where we do enqueue received requests
			receivedReqsChan := make(chan utils.RemoteWriteRequest, 10)
			// count the number for remote-write requests received (which should correlated with the number of sent batches),
			// and the total number of entries.
			var receivedRWsCount atomic.Int64
			var receivedEntriesCount atomic.Int64

			receivedReqs := utils.NewSyncSlice[utils.RemoteWriteRequest]()
			go func() {
				for req := range receivedReqsChan {
					receivedReqs.Append(req)
					receivedRWsCount.Add(1)
					for _, s := range req.Request.Streams {
						receivedEntriesCount.Add(int64(len(s.Entries)))
					}
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
				BatchWait:      tc.batchWait,
				BatchSize:      tc.batchSize,
				Client:         config.HTTPClientConfig{},
				BackoffConfig:  backoff.Config{MinBackoff: 5 * time.Second, MaxBackoff: 10 * time.Second, MaxRetries: 1},
				ExternalLabels: lokiflag.LabelSet{},
				Timeout:        1 * time.Second,
				TenantID:       "",
				Queue:          tc.queueConfig,
			}

			logger := log.NewLogfmtLogger(os.Stdout)

			m := NewMetrics(reg)
			qc, err := NewQueue(m, cfg, 0, 0, false, logger)
			require.NoError(t, err)

			//labels := model.LabelSet{"app": "test"}
			lines := make([]string, 0, tc.numLines)
			for i := 0; i < tc.numLines; i++ {
				lines = append(lines, fmt.Sprintf("hola %d", i))
			}

			// Send all the input log entries
			for i, l := range lines {
				mod := i % tc.numSeries
				qc.StoreSeries([]record.RefSeries{
					{
						Labels: labels.Labels{{
							Name:  "app",
							Value: fmt.Sprintf("test-%d", mod),
						}},
						Ref: chunks.HeadSeriesRef(mod),
					},
				}, 0)

				_ = qc.AppendEntries(wal.RefEntries{
					Ref: chunks.HeadSeriesRef(mod),
					Entries: []logproto.Entry{{
						Timestamp: time.Now(),
						Line:      l,
					}},
				}, 0)
			}

			require.Eventually(t, func() bool {
				return receivedEntriesCount.Load() == int64(len(lines))
			}, time.Second*10, time.Second, "timed out waiting for entries to arrive")

			if tc.expectedRWReqsCount != 0 {
				require.Equal(t, tc.expectedRWReqsCount, receivedRWsCount.Load(), "number for remote write request not expected")
			}

			// Stop the client: it waits until the current batch is sent
			qc.Stop()
			close(receivedReqsChan)
		})
	}
}

func BenchmarkQueueClient(b *testing.B) {
	for name, bc := range map[string]testCase{
		"100 entries single series": {
			numLines:  100,
			numSeries: 1,
		},
		"100k entries, 100 series": {
			numLines:  100_000,
			numSeries: 100,
		},
	} {
		b.Run(name, func(b *testing.B) {
			runSingleBenchCase(b, bc)
		})
	}
}

func runSingleBenchCase(b *testing.B, bc testCase) {
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
	require.NotNil(b, server)
	defer server.Close()

	// Get the URL at which the local test server is listening to
	serverURL := flagext.URLValue{}
	err := serverURL.Set(server.URL)
	require.NoError(b, err)

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
		Queue: QueueConfig{
			Capacity:     1000, // queue size of 100
			DrainTimeout: time.Second * 10,
		},
	}

	logger := log.NewLogfmtLogger(os.Stdout)

	m := NewMetrics(reg)
	qc, err := NewQueue(m, cfg, 0, 0, false, logger)
	require.NoError(b, err)

	//labels := model.LabelSet{"app": "test"}
	var lines []string
	for i := 0; i < bc.numLines; i++ {
		lines = append(lines, fmt.Sprintf("hola %d", i))
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		// Send all the input log entries
		for j, l := range lines {
			seriesId := j % bc.numSeries
			qc.StoreSeries([]record.RefSeries{
				{
					Labels: labels.Labels{{
						Name: "app",
						// take j module bc.numSeries to evenly distribute those numSeries across all sent entries
						Value: fmt.Sprintf("series-%d", seriesId),
					}},
					Ref: chunks.HeadSeriesRef(seriesId),
				},
			}, 0)

			_ = qc.AppendEntries(wal.RefEntries{
				Ref: chunks.HeadSeriesRef(seriesId),
				Entries: []logproto.Entry{{
					Timestamp: time.Now(),
					Line:      l,
				}},
			}, 0)
		}

		require.Eventually(b, func() bool {
			return receivedReqs.Length() == len(lines)
		}, time.Second*10, time.Second, "timed out waiting for messages to arrive")

		// reset receiving slice
		receivedReqs.Reset()
	}

	// Stop the client: it waits until the current batch is sent
	qc.Stop()
	close(receivedReqsChan)
}
