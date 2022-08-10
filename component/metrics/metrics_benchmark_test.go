package metrics

import (
	"fmt"
	"testing"

	"github.com/prometheus/prometheus/model/labels"
)

func BenchmarkLabels(b *testing.B) {
	fms := generateFlowMetrics()
	for i := 0; i < b.N; i++ {
		for j := 0; j < len(fms); j++ {
			GlobalRefMapping.getGlobalRefIDByLabels(fms[j].labels)
		}
	}
}

func BenchmarkFlowMetrics(b *testing.B) {
	fms := generateFlowMetrics()
	for i := 0; i < b.N; i++ {
		for j := 0; j < len(fms); j++ {
			GlobalRefMapping.getGlobalRefIDByLabels(fms[j].LabelsCopy())
		}
	}
}

func generateFlowMetrics() []*FlowMetric {
	fms := make([]*FlowMetric, 100)
	lbls := make(labels.Labels, 10)
	for i := 0; i < len(lbls); i++ {
		lbls[i] = labels.Label{
			Name:  fmt.Sprintf("label_%d", i),
			Value: fmt.Sprintf("value_%d", i),
		}
	}
	for i := 0; i < len(fms); i++ {
		fms[i] = NewFlowMetric(0, lbls, 0)
	}
	return fms
}
