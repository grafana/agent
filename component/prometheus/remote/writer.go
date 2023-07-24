package remote

import "github.com/grafana/agent/component/prometheus"

type RemoteWrite interface {
	Name() string
	AppendMetadata(metadata []prometheus.Metadata) bool
	Append(samples []prometheus.Sample) bool
	AppendExemplars(exemplars []prometheus.Exemplar) bool
	AppendHistograms(histograms []prometheus.Histogram) bool
	AppendFloatHistograms(floatHistograms []prometheus.FloatHistogram) bool
}
