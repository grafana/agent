package metrics

import "github.com/prometheus/client_golang/prometheus"

type Metrics struct {
	PidCacheHit          prometheus.Counter
	PidCacheMiss         prometheus.Counter
	ContainerIDCacheHit  prometheus.Counter
	ContainerIDCacheMiss prometheus.Counter
}

func NewMetrics(reg prometheus.Registerer) *Metrics {
	m := &Metrics{
		PidCacheHit: prometheus.NewCounter(prometheus.CounterOpts{
			Name: "pyroscope_ebpf_pid_cache_hit_total",
			Help: "",
		}),
		PidCacheMiss: prometheus.NewCounter(prometheus.CounterOpts{
			Name: "pyroscope_ebpf_pid_cache_miss_total",
			Help: "",
		}),
		ContainerIDCacheHit: prometheus.NewCounter(prometheus.CounterOpts{
			Name: "pyroscope_ebpf_container_id_cache_hit_total",
			Help: "",
		}),
		ContainerIDCacheMiss: prometheus.NewCounter(prometheus.CounterOpts{
			Name: "pyroscope_ebpf_container_id_cache_miss_total",
			Help: "",
		}),
	}

	if reg != nil {
		reg.MustRegister(
			m.PidCacheHit,
			m.PidCacheMiss,
			m.ContainerIDCacheHit,
			m.ContainerIDCacheMiss,
		)
	}

	return m
}
