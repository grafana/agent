package stages

// NOTE: This code is copied from Promtail (07cbef92268aecc0f20d1791a6df390c2df5c072) with changes kept to the minimum.

import (
	"testing"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/stretchr/testify/assert"

	util_log "github.com/grafana/loki/pkg/util/log"
)

var testDecolorizePipeline = `
stage.decolorize {}
`

func TestPipeline_Decolorize(t *testing.T) {
	t.Parallel()

	tests := map[string]struct {
		config        string
		entry         string
		expectedEntry string
	}{
		"successfully run pipeline on non-colored text": {
			testDecolorizePipeline,
			"sample text",
			"sample text",
		},
		"successfully run pipeline on colored text": {
			testDecolorizePipeline,
			"\033[0;32mgreen\033[0m \033[0;31mred\033[0m",
			"green red",
		},
	}

	for testName, testData := range tests {
		testData := testData

		t.Run(testName, func(t *testing.T) {
			t.Parallel()

			pl, err := NewPipeline(util_log.Logger, loadConfig(testData.config), nil, prometheus.DefaultRegisterer)
			if err != nil {
				t.Fatal(err)
			}
			out := processEntries(pl, newEntry(nil, nil, testData.entry, time.Now()))[0]
			assert.Equal(t, testData.expectedEntry, out.Line)
		})
	}
}
