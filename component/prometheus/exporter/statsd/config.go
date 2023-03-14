package statsd

import (
	"time"

	"github.com/grafana/agent/pkg/integrations/statsd_exporter"
	"github.com/prometheus/statsd_exporter/pkg/mapper"
)

type Config struct {
	ListenUDP      string       `river:"listen_udp,attr,optional"`
	ListenTCP      string       `river:"listen_tcp,attr,optional"`
	ListenUnixgram string       `river:"listen_unixgram,attr,optional"`
	UnixSocketMode string       `river:"unix_socket_mode,attr,optional"`
	MappingConfig  MetricMapper `river:"mapping_config,block,optional"`

	ReadBuffer          int           `river:"read_buffer,attr,optional"`
	CacheSize           int           `river:"cache_size,attr,optional"`
	CacheType           string        `river:"cache_type,attr,optional"`
	EventQueueSize      int           `river:"event_queue_size,attr,optional"`
	EventFlushThreshold int           `river:"event_flush_threshold,attr,optional"`
	EventFlushInterval  time.Duration `river:"event_flush_interval,attr,optional"`

	ParseDogStatsd bool `river:"parse_dogstatsd_tags,attr,optional"`
	ParseInfluxDB  bool `river:"parse_influxdb_tags,attr,optional"`
	ParseLibrato   bool `river:"parse_librato_tags,attr,optional"`
	ParseSignalFX  bool `river:"parse_signalfx_tags,attr,optional"`
}

type MetricMapper struct {
	Defaults MapperConfigDefaults `river:"defaults,block,optional"`
	Mappings []MetricMapping      `river:"mappings,block,optional"`
}

type MapperConfigDefaults struct {
	ObserverType        string           `river:"observer_type,attr,optional"`
	MatchType           string           `river:"match_type,attr,optional"`
	GlobDisableOrdering bool             `river:"glob_disable_ordering,attr,optional"`
	Ttl                 time.Duration    `river:"ttl,attr,optional"`
	SummaryOptions      SummaryOptions   `river:"summary_options,block,optional"`
	HistogramOptions    HistogramOptions `river:"histogram_options,block,optional"`
}

type SummaryOptions struct {
	Quantiles  []MetricObjective `river:"quantiles,block,optional"`
	MaxAge     time.Duration     `river:"max_age,attr,optional"`
	AgeBuckets uint32            `river:"age_buckets,attr,optional"`
	BufCap     uint32            `river:"buf_cap,attr,optional"`
}

type HistogramOptions struct {
	Buckets                     []float64 `river:"buckets,attr,optional"`
	NativeHistogramBucketFactor float64   `river:"native_histogram_bucket_factor,attr,optional"`
	NativeHistogramMaxBuckets   uint32    `river:"native_histogram_max_buckets,attr,optional"`
}

type MetricObjective struct {
	Quantile float64 `river:"quantile,attr,optional"`
	Error    float64 `river:"error,attr,optional"`
}

type MetricMapping struct {
	Match            string            `river:"match,attr,optional"`
	Name             string            `river:"name,attr,optional"`
	Labels           map[string]string `river:"labels,attr,optional"`
	ObserverType     string            `river:"observer_type,attr,optional"`
	TimerType        string            `river:"timer_type,attr,optional"`
	LegacyBuckets    []float64         `river:"buckets,attr,optional"`
	LegacyQuantiles  []MetricObjective `river:"quantiles,block,optional"`
	MatchType        string            `river:"match_type,attr,optional"`
	HelpText         string            `river:"help,attr,optional"`
	Action           string            `river:"action,attr,optional"`
	MatchMetricType  string            `river:"match_metric_type,attr,optional"`
	Ttl              time.Duration     `river:"ttl,attr,optional"`
	SummaryOptions   SummaryOptions    `river:"summary_options,block,optional"`
	HistogramOptions HistogramOptions  `river:"histogram_options,block,optional"`
}

// DefaultConfig holds non-zero default options for the Config when it is
// unmarshaled from YAML.
//
// Some defaults are populated from init functions in the github.com/grafana/agent/pkg/integrations/statsd_exporter package.
var DefaultConfig = Config{

	ListenUDP:      statsd_exporter.DefaultConfig.ListenUDP,
	ListenTCP:      statsd_exporter.DefaultConfig.ListenTCP,
	UnixSocketMode: statsd_exporter.DefaultConfig.UnixSocketMode,

	CacheSize:           statsd_exporter.DefaultConfig.CacheSize,
	CacheType:           statsd_exporter.DefaultConfig.CacheType,
	EventQueueSize:      statsd_exporter.DefaultConfig.EventQueueSize,
	EventFlushThreshold: statsd_exporter.DefaultConfig.EventFlushThreshold,
	EventFlushInterval:  statsd_exporter.DefaultConfig.EventFlushInterval,

	ParseDogStatsd: statsd_exporter.DefaultConfig.ParseDogStatsd,
	ParseInfluxDB:  statsd_exporter.DefaultConfig.ParseInfluxDB,
	ParseLibrato:   statsd_exporter.DefaultConfig.ParseLibrato,
	ParseSignalFX:  statsd_exporter.DefaultConfig.ParseSignalFX,
}

// Convert gives a config suitable for use with github.com/grafana/agent/pkg/integrations/statsd_exporter.
func (c *Config) Convert() *statsd_exporter.Config {

	return &statsd_exporter.Config{
		ListenUDP:           c.ListenUDP,
		ListenTCP:           c.ListenTCP,
		ListenUnixgram:      c.ListenUnixgram,
		UnixSocketMode:      c.UnixSocketMode,
		ReadBuffer:          c.ReadBuffer,
		CacheSize:           c.CacheSize,
		CacheType:           c.CacheType,
		EventQueueSize:      c.EventQueueSize,
		EventFlushThreshold: c.EventFlushThreshold,
		EventFlushInterval:  c.EventFlushInterval,
		ParseDogStatsd:      c.ParseDogStatsd,
		ParseInfluxDB:       c.ParseInfluxDB,
		ParseLibrato:        c.ParseLibrato,
		ParseSignalFX:       c.ParseSignalFX,
		MappingConfig: &mapper.MetricMapper{
			Defaults: mapper.MapperConfigDefaults{
				ObserverType:        mapper.ObserverType(c.MappingConfig.Defaults.ObserverType),
				MatchType:           mapper.MatchType(c.MappingConfig.Defaults.MatchType),
				GlobDisableOrdering: c.MappingConfig.Defaults.GlobDisableOrdering,
				Ttl:                 c.MappingConfig.Defaults.Ttl,
				SummaryOptions:      *convertSummaryOptions(c.MappingConfig.Defaults.SummaryOptions),
				HistogramOptions:    *convertHistogramOptions(c.MappingConfig.Defaults.HistogramOptions),
			},
			Mappings: convertMappings(c.MappingConfig.Mappings),
		},
	}
}

func convertMappings(m []MetricMapping) []mapper.MetricMapping {
	var out []mapper.MetricMapping
	for _, v := range m {
		mapping := mapper.MetricMapping{
			Match:            v.Match,
			Name:             v.Name,
			Labels:           v.Labels,
			ObserverType:     mapper.ObserverType(v.ObserverType),
			TimerType:        mapper.ObserverType(v.TimerType),
			LegacyBuckets:    v.LegacyBuckets,
			LegacyQuantiles:  convertMetricObjective(v.LegacyQuantiles),
			MatchType:        mapper.MatchType(v.MatchType),
			HelpText:         v.HelpText,
			Action:           mapper.ActionType(v.Action),
			MatchMetricType:  mapper.MetricType(v.MatchMetricType),
			Ttl:              v.Ttl,
			SummaryOptions:   convertSummaryOptions(v.SummaryOptions),
			HistogramOptions: convertHistogramOptions(v.HistogramOptions),
		}
		out = append(out, mapping)
	}
	return out
}

func convertSummaryOptions(s SummaryOptions) *mapper.SummaryOptions {
	return &mapper.SummaryOptions{
		Quantiles:  convertMetricObjective(s.Quantiles),
		MaxAge:     s.MaxAge,
		AgeBuckets: s.AgeBuckets,
		BufCap:     s.BufCap,
	}
}

func convertHistogramOptions(h HistogramOptions) *mapper.HistogramOptions {
	return &mapper.HistogramOptions{
		Buckets:                     h.Buckets,
		NativeHistogramBucketFactor: h.NativeHistogramBucketFactor,
		NativeHistogramMaxBuckets:   h.NativeHistogramMaxBuckets,
	}
}

func convertMetricObjective(m []MetricObjective) []mapper.MetricObjective {
	var out []mapper.MetricObjective
	for _, v := range m {
		mo := mapper.MetricObjective{
			Quantile: v.Quantile,
			Error:    v.Error,
		}
		out = append(out, mo)
	}
	return out
}

// UnmarshalRiver implements River unmarshalling for Config.
func (c *Config) UnmarshalRiver(f func(interface{}) error) error {
	*c = DefaultConfig

	type cfg Config
	return f((*cfg)(c))
}
