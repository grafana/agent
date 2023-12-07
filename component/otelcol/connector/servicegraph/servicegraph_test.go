package servicegraph_test

import (
	"testing"
	"time"

	"github.com/grafana/agent/component/otelcol/connector/servicegraph"
	"github.com/grafana/river"
	"github.com/open-telemetry/opentelemetry-collector-contrib/processor/servicegraphprocessor"
	"github.com/stretchr/testify/require"
)

func TestArguments_UnmarshalRiver(t *testing.T) {
	tests := []struct {
		testName string
		cfg      string
		expected servicegraphprocessor.Config
		errorMsg string
	}{
		{
			testName: "Defaults",
			cfg: `
				output {}
			`,
			expected: servicegraphprocessor.Config{
				LatencyHistogramBuckets: []time.Duration{
					2 * time.Millisecond,
					4 * time.Millisecond,
					6 * time.Millisecond,
					8 * time.Millisecond,
					10 * time.Millisecond,
					50 * time.Millisecond,
					100 * time.Millisecond,
					200 * time.Millisecond,
					400 * time.Millisecond,
					800 * time.Millisecond,
					1 * time.Second,
					1400 * time.Millisecond,
					2 * time.Second,
					5 * time.Second,
					10 * time.Second,
					15 * time.Second,
				},
				Dimensions: []string{},
				Store: servicegraphprocessor.StoreConfig{
					MaxItems: 1000,
					TTL:      2 * time.Millisecond,
				},
				CacheLoop:           1 * time.Minute,
				StoreExpirationLoop: 2 * time.Second,
				//TODO: Add VirtualNodePeerAttributes when it's no longer controlled by
				// the "processor.servicegraph.virtualNode" feature gate.
				// VirtualNodePeerAttributes: []string{
				// 				"db.name",
				// 				"net.sock.peer.addr",
				// 				"net.peer.name",
				// 				"rpc.service",
				// 				"net.sock.peer.name",
				// 				"net.peer.name",
				// 				"http.url",
				// 				"http.target",
				// 			},
			},
		},
		{
			testName: "ExplicitValues",
			cfg: `
					dimensions = ["foo", "bar"]
					latency_histogram_buckets = ["2ms", "4s", "6h"]
					store {
						max_items = 333
						ttl = "12h"
					}
					cache_loop = "55m"
					store_expiration_loop = "77s"

					output {}
				`,
			expected: servicegraphprocessor.Config{
				LatencyHistogramBuckets: []time.Duration{
					2 * time.Millisecond,
					4 * time.Second,
					6 * time.Hour,
				},
				Dimensions: []string{"foo", "bar"},
				Store: servicegraphprocessor.StoreConfig{
					MaxItems: 333,
					TTL:      12 * time.Hour,
				},
				CacheLoop:           55 * time.Minute,
				StoreExpirationLoop: 77 * time.Second,
				//TODO: Ad VirtualNodePeerAttributes when it's no longer controlled by
				// the "processor.servicegraph.virtualNode" feature gate.
				// VirtualNodePeerAttributes: []string{"attr1", "attr2"},
			},
		},
		{
			testName: "InvalidCacheLoop",
			cfg: `
					cache_loop = "0s"
					output {}
				`,
			errorMsg: "cache_loop must be greater than 0",
		},
		{
			testName: "InvalidStoreExpirationLoop",
			cfg: `
					store_expiration_loop = "0s"
					output {}
				`,
			errorMsg: "store_expiration_loop must be greater than 0",
		},
		{
			testName: "InvalidStoreMaxItems",
			cfg: `
					store {
						max_items = 0
					}

					output {}
				`,
			errorMsg: "store.max_items must be greater than 0",
		},
		{
			testName: "InvalidStoreTTL",
			cfg: `
					store {
						ttl = "0s"
					}

					output {}
				`,
			errorMsg: "store.ttl must be greater than 0",
		},
	}

	for _, tc := range tests {
		t.Run(tc.testName, func(t *testing.T) {
			var args servicegraph.Arguments
			err := river.Unmarshal([]byte(tc.cfg), &args)
			if tc.errorMsg != "" {
				require.ErrorContains(t, err, tc.errorMsg)
				return
			}

			require.NoError(t, err)

			actualPtr, err := args.Convert()
			require.NoError(t, err)

			actual := actualPtr.(*servicegraphprocessor.Config)

			require.Equal(t, tc.expected, *actual)
		})
	}
}
