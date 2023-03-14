package statsd

import (
	"testing"
	"time"

	"github.com/grafana/agent/pkg/river"
	"github.com/stretchr/testify/require"
)

func TestRiverUnmarshall(t *testing.T) {
	var exampleRiverConfig = `
	listen_udp							= "1010"
	listen_tcp							= "1011"
	listen_unixgram						= "unix"
	unix_socket_mode					= "prom"
	mapping_config {
		defaults {
			observer_type				= "histogram"
			match_type					= "glob"
			glob_disable_ordering 		= false
			ttl							= "1s"
			summary_options {
				quantiles [{
						quantile		= 0.95,
						error			= 0.05,
					},
					{
						quantile		= 0.98,
						error			= 0.02,
					}]
				max_age					= "1s"
				age_buckets				= 1
				buf_cap					= 1
			}
			histogram_options {
				buckets					= [0.25, 0.5, 0.75, 1]
				native_histogram_bucket_factor = 0.1
				native_histogram_max_buckets = 4
			}
		}
		mappings = [
			{
				match					= "match",
				name					= "name1",
				observer_type			= "summary",
				timer_type				= "summary",
				buckets					= [0.25, 0.5, 0.75, 1],
				quantiles [
					{
						quantile		= 0.98,
						error			= 0.005,
					},
					{
						quantile		= 0.99,
						error			= 0.005,
					},
				],
				match_type				= "regex",
				help					= "help",
				action					= "map",
				match_metric_type		= "glob",
				ttl						= "1m",
				summary_options = {
					quantiles = [
						{
							quantile		= 0.95,
							error			= 0.05,
						},
						{
							quantile		= 0.98,
							error			= 0.02,
						},
					],
					max_age					= "1s",
					age_buckets				= 1,
					buf_cap					= 1,
				},
				histogram_options = {
					buckets					= [0.25, 0.5, 0.75, 1],
					native_histogram_bucket_factor = 0.1,
					native_histogram_max_buckets = 4,
				},
			},
		]
	}
	`

	var args Config
	err := river.Unmarshal([]byte(exampleRiverConfig), &args)
	require.NoError(t, err)

	require.Equal(t, "1010", string(args.ListenTCP))
	require.Equal(t, "1011", string(args.ListenTCP))
	require.Equal(t, "unix", string(args.ListenUnixgram))
	require.Equal(t, "prom", string(args.UnixSocketMode))
	require.Equal(t, "histogram", string(args.MappingConfig.Defaults.ObserverType))
	require.Equal(t, "glob", string(args.MappingConfig.Defaults.MatchType))
	require.Equal(t, false, bool(args.MappingConfig.Defaults.GlobDisableOrdering))
	require.Equal(t, time.Duration.Seconds(1), time.Duration(args.MappingConfig.Defaults.Ttl))
	require.Equal(t, []MetricObjective{{Quantile: 0.95, Error: 0.05}, {Quantile: 0.98, Error: 0.05}}, []MetricObjective(args.MappingConfig.Defaults.SummaryOptions.Quantiles))
	require.Equal(t, time.Duration.Seconds(1), time.Duration(args.MappingConfig.Defaults.SummaryOptions.MaxAge))
	require.Equal(t, 1, int(args.MappingConfig.Defaults.SummaryOptions.AgeBuckets))
	require.Equal(t, 1, int(args.MappingConfig.Defaults.SummaryOptions.BufCap))
	require.Equal(t, []float64{0.25, 0.5, 0.75, 1}, []float64(args.MappingConfig.Defaults.HistogramOptions.Buckets))
	require.Equal(t, 0.1, float64(args.MappingConfig.Defaults.HistogramOptions.NativeHistogramBucketFactor))
	require.Equal(t, 4, int(args.MappingConfig.Defaults.HistogramOptions.NativeHistogramMaxBuckets))
}
