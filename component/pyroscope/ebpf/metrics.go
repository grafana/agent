package ebpf

import "github.com/prometheus/client_golang/prometheus"

type Metrics struct {
	TargetsActive          prometheus.Gauge
	ProfilingSessionsTotal prometheus.Counter
	PprofsTotal            prometheus.Counter
}

func NewMetrics(reg prometheus.Registerer) *Metrics {
	m := &Metrics{
		TargetsActive: prometheus.NewGauge(prometheus.GaugeOpts{
			Name: "pyroscope_ebpf_active_targets",
			Help: "Current number of active targets being tracked by the ebpf component",
		}),
		ProfilingSessionsTotal: prometheus.NewCounter(prometheus.CounterOpts{
			Name: "pyroscope_ebpf_profiling_sessions_total",
			Help: "Total number of profiling sessions started by the ebpf component",
		}),
		PprofsTotal: prometheus.NewCounter(prometheus.CounterOpts{
			Name: "pyroscope_ebpf_pprofs_total",
			Help: "Total number of pprof profiles collected by the ebpf component",
		}),
	}

	if reg != nil {
		reg.MustRegister(
			m.TargetsActive,
			m.ProfilingSessionsTotal,
			m.PprofsTotal,
		)
	}

	return m
}
