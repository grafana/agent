package stages

// NOTE: This code is copied from Promtail (07cbef92268aecc0f20d1791a6df390c2df5c072) with changes kept to the minimum.

import (
	"github.com/grafana/loki/pkg/logql/log"
)

type DecolorizeConfig struct{}

type decolorizeStage struct{}

func newDecolorizeStage(_ DecolorizeConfig) (Stage, error) {
	return &decolorizeStage{}, nil
}

// Run implements Stage
func (m *decolorizeStage) Run(in chan Entry) chan Entry {
	decolorizer, _ := log.NewDecolorizer()
	out := make(chan Entry)
	go func() {
		defer close(out)
		for e := range in {
			decolorizedLine, _ := decolorizer.Process(
				e.Timestamp.Unix(),
				[]byte(e.Entry.Line),
				nil,
			)
			e.Entry.Line = string(decolorizedLine)
			out <- e
		}
	}()
	return out
}

// Name implements Stage
func (m *decolorizeStage) Name() string {
	return StageTypeDecolorize
}
