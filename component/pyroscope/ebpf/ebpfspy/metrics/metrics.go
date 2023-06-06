package metrics

import "github.com/prometheus/client_golang/prometheus"

type Metrics struct {
	PidCacheHit          prometheus.Counter
	PidCacheMiss         prometheus.Counter
	ElfCacheBuildIDHit   prometheus.Counter
	ElfCacheBuildIDMiss  prometheus.Counter
	ElfCacheStatHit      prometheus.Counter
	ElfCacheStatMiss     prometheus.Counter
	ContainerIDCacheHit  prometheus.Counter
	ContainerIDCacheMiss prometheus.Counter
}

func NewMetrics(reg prometheus.Registerer) *Metrics {
	m := &Metrics{
		PidCacheHit: prometheus.NewCounter(prometheus.CounterOpts{
			Name: "pyroscope_ebpf_pid_cache_hit_total",
			Help: "Total number of ebpf symbolizer pid cache hit.",
		}),
		PidCacheMiss: prometheus.NewCounter(prometheus.CounterOpts{
			Name: "pyroscope_ebpf_pid_cache_miss_total",
			Help: "Total number of ebpf symbolizer pid cache miss.",
		}),
		ElfCacheBuildIDHit: prometheus.NewCounter(prometheus.CounterOpts{
			Name: "pyroscope_ebpf_elf_cache_build_id_hit_total",
			Help: "Total number of ebpf symbolizer elf cache (build-id) hit.",
		}),
		ElfCacheBuildIDMiss: prometheus.NewCounter(prometheus.CounterOpts{
			Name: "pyroscope_ebpf_elf_cache_build_id_miss_total",
			Help: "Total number of ebpf symbolizer elf cache (build-id) miss.",
		}),
		ElfCacheStatHit: prometheus.NewCounter(prometheus.CounterOpts{
			Name: "pyroscope_ebpf_elf_cache_stat_hit_total",
			Help: "Total number of ebpf symbolizer elf cache (stat) hit.",
		}),
		ElfCacheStatMiss: prometheus.NewCounter(prometheus.CounterOpts{
			Name: "pyroscope_ebpf_elf_cache_stat_miss_total",
			Help: "Total number of ebpf symbolizer elf cache (stat) miss.",
		}),
		ContainerIDCacheHit: prometheus.NewCounter(prometheus.CounterOpts{
			Name: "pyroscope_ebpf_container_id_cache_hit_total",
			Help: "Total number of ebpf target finder container id cache hit.",
		}),
		ContainerIDCacheMiss: prometheus.NewCounter(prometheus.CounterOpts{
			Name: "pyroscope_ebpf_container_id_cache_miss_total",
			Help: "Total number of ebpf target finder container id cache miss.",
		}),
	}

	if reg != nil {
		reg.MustRegister(
			m.PidCacheHit,
			m.PidCacheMiss,
			m.ElfCacheBuildIDHit,
			m.ElfCacheBuildIDMiss,
			m.ElfCacheStatHit,
			m.ElfCacheStatMiss,
			m.ContainerIDCacheHit,
			m.ContainerIDCacheMiss,
		)
	}

	return m
}
