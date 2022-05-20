package ssl_exporter

import (
	"context"
	"fmt"
	"sort"
	"sync"

	"github.com/go-kit/log"
	"github.com/go-kit/log/level"
	sslv1 "github.com/grafana/agent/pkg/integrations/ssl_exporter"
	"github.com/grafana/agent/pkg/integrations/v2"
	"github.com/prometheus/client_golang/prometheus"
	ssl_config "github.com/ribbybibby/ssl_exporter/v2/config"
	"github.com/ribbybibby/ssl_exporter/v2/prober"
)

func init() {
	integrations.Register(&Config{}, integrations.TypeMultiplex)
}

type Exporter struct {
	sync.Mutex
	probeSuccess prometheus.Gauge
	proberType   *prometheus.GaugeVec

	options   Options
	namespace string
}

func (e *Exporter) RunIntegration(ctx context.Context) error {
	<-ctx.Done()
	return nil
}

type Options struct {
	Namespace   string
	MetricsPath string
	ProbePath   string
	Registry    *prometheus.Registry
	SSLTargets  []SSLTarget
	SSLConfig   *ssl_config.Config
	Logger      log.Logger
}

func NewSSLExporter(opts Options) (*Exporter, error) {
	e := &Exporter{
		options:   opts,
		namespace: opts.Namespace,
		probeSuccess: prometheus.NewGauge(
			prometheus.GaugeOpts{
				Name: prometheus.BuildFQName(opts.Namespace, "", "probe_success"),
				Help: "If the probe was a success",
			},
		),
		proberType: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: prometheus.BuildFQName(opts.Namespace, "", "prober"),
				Help: "The prober used by the exporter to connect to the target",
			},
			[]string{"prober"},
		),
	}

	return e, nil
}

func (e *Exporter) Describe(ch chan<- *prometheus.Desc) {
	for _, desc := range sslv1.Descs {
		ch <- desc
	}
}

func (e *Exporter) Collect(ch chan<- prometheus.Metric) {
	e.Lock()
	defer e.Unlock()

	logger := e.options.Logger

	for _, target := range e.options.SSLTargets {
		ctx := context.Background()

		var moduleName string
		if target.Module != "" {
			moduleName = e.options.SSLConfig.DefaultModule
			if moduleName == "" {
				level.Error(logger).Log("msg", "Module parameter must be set")
				continue
			}
		}

		module, ok := e.options.SSLConfig.Modules[target.Module]
		if !ok {
			level.Error(logger).Log("msg", fmt.Sprintf("Unknown module '%s'", target.Module))
			continue
		}

		probeFunc, ok := prober.Probers[module.Prober]
		if !ok {
			level.Error(logger).Log("msg", fmt.Sprintf("Unknown prober %q", module.Prober))
			continue
		}

		e.options.Registry = prometheus.NewRegistry()
		e.options.Registry.MustRegister(e.probeSuccess, e.proberType)
		e.proberType.WithLabelValues(module.Prober).Set(1)

		// set high-level metric not collected in the prober
		err := probeFunc(ctx, logger, target.Target, module, e.options.Registry)
		if err != nil {
			level.Error(logger).Log("msg", err)
			e.probeSuccess.Set(0)
		} else {
			e.probeSuccess.Set(1)
		}

		// gather all the metrics we've collected in the prober
		metricFams, err := e.options.Registry.Gather()
		if err != nil {
			level.Error(logger).Log("msg", err)
			continue
		}
		for _, mf := range metricFams {
			for _, m := range mf.Metric {
				// get desc from name
				desc, ok := sslv1.Descs[*mf.Name]
				if !ok {
					level.Error(logger).Log("msg", fmt.Sprintf("Unknown metric %q", *mf.Name))
					continue
				}

				// ensure label order
				sort.Slice(m.Label, func(i, j int) bool {
					iPrec := sslv1.LabelOrder[*m.Label[i].Name]
					jPrec := sslv1.LabelOrder[*m.Label[j].Name]
					return iPrec < jPrec
				})
				labelValues := []string{}
				for _, l := range m.Label {
					labelValues = append(labelValues, *l.Value)
				}

				// create prometheus metric
				metric, err := prometheus.NewConstMetric(desc, prometheus.GaugeValue, *m.Gauge.Value, labelValues...)
				if err != nil {
					level.Error(logger).Log("msg", err)
					continue
				}
				ch <- metric
			}
		}
	}
}
