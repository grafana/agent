package kafkatarget

// This code is copied from Promtail. The kafkatarget package is used to
// configure and run the targets that can read kafka entries and forward them
// to other loki components.

import (
	"strings"

	"github.com/prometheus/common/model"
	"github.com/prometheus/prometheus/model/labels"
	"github.com/prometheus/prometheus/model/relabel"

	"github.com/grafana/loki/pkg/util"
)

func format(lbs labels.Labels, cfg []*relabel.Config) model.LabelSet {
	if len(lbs) == 0 {
		return nil
	}
	processed := relabel.Process(lbs, cfg...)
	labelOut := model.LabelSet(util.LabelsToMetric(processed))
	for k := range labelOut {
		if strings.HasPrefix(string(k), "__") {
			delete(labelOut, k)
		}
	}
	return labelOut
}
