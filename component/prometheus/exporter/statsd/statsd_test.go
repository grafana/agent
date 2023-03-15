package statsd

import (
	"testing"
	"time"

	"github.com/grafana/agent/pkg/river"
	"github.com/prometheus/statsd_exporter/pkg/mapper"
	"github.com/stretchr/testify/require"
)

var (
	exampleRiverConfig = `
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
					quantiles {
						quantile			= 0.95
						error				= 0.05
					}
					quantiles {
						quantile			= 0.98
						error				= 0.02
					}
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
			mappings {
				match						= "match"
				name						= "name1"
				observer_type				= "summary"
				timer_type					= "summary"
				match_type					= "regex"
				help						= "help"
				action						= "map"
				match_metric_type			= "glob"
				ttl							= "1m"
				summary_options {
					quantiles {
						quantile			= 0.95
						error				= 0.05
					}
					quantiles {
						quantile			= 0.98
						error				= 0.02
					}
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
		}
		read_buffer						= 1
		cache_size						= 2
		cache_type						= "any"
		event_queue_size				= 1000
		event_flush_interval			= "1m"
		parse_dogstatsd_tags			= true
		parse_influxdb_tags				= false
		parse_librato_tags				= false
		parse_signalfx_tags				= false
		`
	duration1s, _ = time.ParseDuration("1s")
	duration1m, _ = time.ParseDuration("1m")
)

func TestRiverUnmarshall(t *testing.T) {

	var args Config
	err := river.Unmarshal([]byte(exampleRiverConfig), &args)
	require.NoError(t, err)

	require.Equal(t, "1010", string(args.ListenUDP))
	require.Equal(t, "1011", string(args.ListenTCP))
	require.Equal(t, "unix", string(args.ListenUnixgram))
	require.Equal(t, "prom", string(args.UnixSocketMode))

	require.Equal(t, "histogram", string(args.MappingConfig.Defaults.ObserverType))
	require.Equal(t, "glob", string(args.MappingConfig.Defaults.MatchType))
	require.Equal(t, false, bool(args.MappingConfig.Defaults.GlobDisableOrdering))
	require.Equal(t, duration1s, args.MappingConfig.Defaults.Ttl)
	require.Equal(t, []MetricObjective{{Quantile: 0.95, Error: 0.05}, {Quantile: 0.98, Error: 0.02}}, []MetricObjective(args.MappingConfig.Defaults.SummaryOptions.Quantiles))
	require.Equal(t, duration1s, args.MappingConfig.Defaults.SummaryOptions.MaxAge)
	require.Equal(t, 1, int(args.MappingConfig.Defaults.SummaryOptions.AgeBuckets))
	require.Equal(t, 1, int(args.MappingConfig.Defaults.SummaryOptions.BufCap))
	require.Equal(t, []float64{0.25, 0.5, 0.75, 1}, []float64(args.MappingConfig.Defaults.HistogramOptions.Buckets))
	require.Equal(t, 0.1, float64(args.MappingConfig.Defaults.HistogramOptions.NativeHistogramBucketFactor))
	require.Equal(t, 4, int(args.MappingConfig.Defaults.HistogramOptions.NativeHistogramMaxBuckets))

	require.Equal(t, "match", string(args.MappingConfig.Mappings[0].Match))
	require.Equal(t, "name1", string(args.MappingConfig.Mappings[0].Name))
	require.Equal(t, "summary", string(args.MappingConfig.Mappings[0].ObserverType))
	require.Equal(t, "summary", string(args.MappingConfig.Mappings[0].TimerType))
	require.Equal(t, "regex", string(args.MappingConfig.Mappings[0].MatchType))
	require.Equal(t, "help", string(args.MappingConfig.Mappings[0].HelpText))
	require.Equal(t, "map", string(args.MappingConfig.Mappings[0].Action))
	require.Equal(t, "glob", string(args.MappingConfig.Mappings[0].MatchMetricType))
	require.Equal(t, duration1m, args.MappingConfig.Mappings[0].Ttl)
	require.Equal(t, []MetricObjective{{Quantile: 0.95, Error: 0.05}, {Quantile: 0.98, Error: 0.02}}, []MetricObjective(args.MappingConfig.Mappings[0].SummaryOptions.Quantiles))
	require.Equal(t, duration1s, args.MappingConfig.Mappings[0].SummaryOptions.MaxAge)
	require.Equal(t, 1, int(args.MappingConfig.Mappings[0].SummaryOptions.AgeBuckets))
	require.Equal(t, 1, int(args.MappingConfig.Mappings[0].SummaryOptions.BufCap))
	require.Equal(t, []float64{0.25, 0.5, 0.75, 1}, []float64(args.MappingConfig.Mappings[0].HistogramOptions.Buckets))
	require.Equal(t, 0.1, float64(args.MappingConfig.Mappings[0].HistogramOptions.NativeHistogramBucketFactor))
	require.Equal(t, 4, int(args.MappingConfig.Mappings[0].HistogramOptions.NativeHistogramMaxBuckets))

	require.Equal(t, 1, int(args.ReadBuffer))
	require.Equal(t, 2, int(args.CacheSize))
	require.Equal(t, "any", string(args.CacheType))
	require.Equal(t, 1000, int(args.EventQueueSize))
	require.Equal(t, duration1m, args.EventFlushInterval)
	require.Equal(t, true, args.ParseDogStatsd)
	require.Equal(t, false, args.ParseInfluxDB)
	require.Equal(t, false, args.ParseLibrato)
	require.Equal(t, false, args.ParseSignalFX)
}

func TestConvert(t *testing.T) {
	var args Config
	err := river.Unmarshal([]byte(exampleRiverConfig), &args)
	require.NoError(t, err)

	configStatsd := args.Convert()

	require.Equal(t, "1010", string(configStatsd.ListenUDP))
	require.Equal(t, "1011", string(configStatsd.ListenTCP))
	require.Equal(t, "unix", string(configStatsd.ListenUnixgram))
	require.Equal(t, "prom", string(configStatsd.UnixSocketMode))

	require.Equal(t, mapper.ObserverTypeHistogram, configStatsd.MappingConfig.Defaults.ObserverType)
	require.Equal(t, mapper.MatchTypeGlob, configStatsd.MappingConfig.Defaults.MatchType)
	require.Equal(t, false, bool(configStatsd.MappingConfig.Defaults.GlobDisableOrdering))
	require.Equal(t, duration1s, configStatsd.MappingConfig.Defaults.Ttl)
	require.Equal(t, []mapper.MetricObjective{{Quantile: 0.95, Error: 0.05}, {Quantile: 0.98, Error: 0.02}}, []mapper.MetricObjective(configStatsd.MappingConfig.Defaults.SummaryOptions.Quantiles))
	require.Equal(t, duration1s, configStatsd.MappingConfig.Defaults.SummaryOptions.MaxAge)
	require.Equal(t, 1, int(configStatsd.MappingConfig.Defaults.SummaryOptions.AgeBuckets))
	require.Equal(t, 1, int(configStatsd.MappingConfig.Defaults.SummaryOptions.BufCap))
	require.Equal(t, []float64{0.25, 0.5, 0.75, 1}, []float64(configStatsd.MappingConfig.Defaults.HistogramOptions.Buckets))
	require.Equal(t, 0.1, float64(configStatsd.MappingConfig.Defaults.HistogramOptions.NativeHistogramBucketFactor))
	require.Equal(t, 4, int(configStatsd.MappingConfig.Defaults.HistogramOptions.NativeHistogramMaxBuckets))

	require.Equal(t, "match", string(configStatsd.MappingConfig.Mappings[0].Match))
	require.Equal(t, "name1", string(configStatsd.MappingConfig.Mappings[0].Name))
	require.Equal(t, "summary", string(configStatsd.MappingConfig.Mappings[0].ObserverType))
	require.Equal(t, "summary", string(configStatsd.MappingConfig.Mappings[0].TimerType))
	require.Equal(t, "regex", string(configStatsd.MappingConfig.Mappings[0].MatchType))
	require.Equal(t, "help", string(configStatsd.MappingConfig.Mappings[0].HelpText))
	require.Equal(t, "map", string(configStatsd.MappingConfig.Mappings[0].Action))
	require.Equal(t, "glob", string(configStatsd.MappingConfig.Mappings[0].MatchMetricType))
	require.Equal(t, duration1m, configStatsd.MappingConfig.Mappings[0].Ttl)
	require.Equal(t, []mapper.MetricObjective{{Quantile: 0.95, Error: 0.05}, {Quantile: 0.98, Error: 0.02}}, []mapper.MetricObjective(configStatsd.MappingConfig.Mappings[0].SummaryOptions.Quantiles))
	require.Equal(t, duration1s, configStatsd.MappingConfig.Mappings[0].SummaryOptions.MaxAge)
	require.Equal(t, 1, int(configStatsd.MappingConfig.Mappings[0].SummaryOptions.AgeBuckets))
	require.Equal(t, 1, int(configStatsd.MappingConfig.Mappings[0].SummaryOptions.BufCap))
	require.Equal(t, []float64{0.25, 0.5, 0.75, 1}, []float64(configStatsd.MappingConfig.Mappings[0].HistogramOptions.Buckets))
	require.Equal(t, 0.1, float64(configStatsd.MappingConfig.Mappings[0].HistogramOptions.NativeHistogramBucketFactor))
	require.Equal(t, 4, int(configStatsd.MappingConfig.Mappings[0].HistogramOptions.NativeHistogramMaxBuckets))

	require.Equal(t, 1, int(configStatsd.ReadBuffer))
	require.Equal(t, 2, int(configStatsd.CacheSize))
	require.Equal(t, "any", string(configStatsd.CacheType))
	require.Equal(t, 1000, int(configStatsd.EventQueueSize))
	require.Equal(t, duration1m, configStatsd.EventFlushInterval)
	require.Equal(t, true, configStatsd.ParseDogStatsd)
	require.Equal(t, false, configStatsd.ParseInfluxDB)
	require.Equal(t, false, configStatsd.ParseLibrato)
	require.Equal(t, false, configStatsd.ParseSignalFX)
}
