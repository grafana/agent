package exporters

import (
	"github.com/grafana/agent/pkg/integrations/v2/app_o11y_exporter/config"
	"github.com/grafana/agent/pkg/integrations/v2/app_o11y_exporter/models"
	"github.com/prometheus/client_golang/prometheus"
)

var WEB_VITALS = map[string]string{
	"lcp":  "Largest Contentful Paint",
	"fid":  "First Input Delay",
	"cls":  "Cumulative Layout Shift",
	"ttfb": "Time To First Byte",
	"fcp":  "First Contentful Paint",
}

type PrometheusMetricsExporter struct {
	wv     map[string]prometheus.Summary
	custom map[string]prometheus.Summary
}

func NewPrometheusMetricsExporter(reg *prometheus.Registry, metrics []config.Measurement) AppReceiverExporter {
	wv := make(map[string]prometheus.Summary, len(WEB_VITALS))
	for k, v := range WEB_VITALS {
		wv[k] = prometheus.NewSummary(prometheus.SummaryOpts{
			Name: k,
			Help: v,
		})
		reg.MustRegister(wv[k])
	}

	cMetrics := make(map[string]prometheus.Summary, len(metrics))
	for _, m := range metrics {
		cMetrics[m.Name] = prometheus.NewSummary(prometheus.SummaryOpts{
			Name: m.Name,
			Help: m.Description,
		})
	}

	return &PrometheusMetricsExporter{
		wv:     wv,
		custom: cMetrics,
	}
}

func (pe *PrometheusMetricsExporter) Init() error {
	return nil
}

func (pe *PrometheusMetricsExporter) Process(payload models.Payload) error {
	for _, m := range payload.Measurements {
		err := pe.processMesaurement(m)
		if err != nil {
			return err
		}
	}
	return nil
}

func (pe *PrometheusMetricsExporter) processMesaurement(m models.Measurement) error {
	switch m.Type {
	default:
		return nil
	case models.MTYPE_WEBVITALS:
		for k := range WEB_VITALS {
			if v, ok := m.Values[k]; ok {
				pe.wv[k].Observe(v)
			}
		}
	case models.MTYPE_CUSTOM:
		for k := range pe.custom {
			if v, ok := m.Values[k]; ok {
				pe.custom[k].Observe(v)
			}
		}
	}
	return nil
}

// Static typecheck tests
var (
	_ AppReceiverExporter = (*PrometheusMetricsExporter)(nil)
	_ AppMetricsExporter  = (*PrometheusMetricsExporter)(nil)
)
