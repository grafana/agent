package crow

import (
	"fmt"
	"math/rand"
	"time"

	"github.com/prometheus/client_golang/prometheus"
)

type sample struct {
	ScrapeTime time.Time
	Labels     prometheus.Labels
	Value      float64

	// How many times this sample has attempted to be validated. Starts at 0.
	ValidationAttempt int
}

// Ready checks if this sample is ready to be validated.
func (s *sample) Ready(now time.Time) bool {
	backoff := sampleBackoff(s.ValidationAttempt)
	return now.After(s.ScrapeTime.Add(backoff))
}

func sampleBackoff(attempt int) time.Duration {
	// Exponential backoff from 1s up to 1s + (250ms * 2^attempt).
	return time.Second + (250 * time.Millisecond * 1 << attempt)
}

type sampleGenerator struct {
	numSamples int
	sendCh     chan<- []*sample
	r          *rand.Rand
}

const validationSampleName = "crow_validation_sample"

func (sg *sampleGenerator) Describe(ch chan<- *prometheus.Desc) {
	ch <- prometheus.NewDesc(
		validationSampleName, "Sample to validate",
		[]string{"sample_num"},
		prometheus.Labels{},
	)
}

func (sg *sampleGenerator) Collect(ch chan<- prometheus.Metric) {
	var (
		scrapeTime = time.Now()

		sampleLabel = "sample_num"
		desc        = prometheus.NewDesc(
			validationSampleName, "Sample to validate",
			[]string{sampleLabel},
			prometheus.Labels{},
		)

		usedLabels = map[string]struct{}{}
		samples    = make([]*sample, sg.numSamples)
	)

	for s := 0; s < sg.numSamples; s++ {
	GenLabel:
		labelSuffix := make([]byte, 1)
		_, _ = sg.r.Read(labelSuffix)
		label := fmt.Sprintf("sample_%x", labelSuffix)
		if _, exist := usedLabels[label]; exist {
			goto GenLabel
		}
		usedLabels[label] = struct{}{}

		samples[s] = &sample{
			ScrapeTime: scrapeTime,
			Labels:     prometheus.Labels{sampleLabel: label},
			Value:      float64(sg.r.Int63n(1_000_000)),
		}
		ch <- prometheus.MustNewConstMetric(
			desc,
			prometheus.GaugeValue,
			samples[s].Value, samples[s].Labels[sampleLabel],
		)
	}

	sg.sendCh <- samples
}
