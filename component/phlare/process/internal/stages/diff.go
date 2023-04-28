package stages

// This package is ported over from grafana/loki/clients/pkg/logentry/stages.
// We aim to port the stages in steps, to avoid introducing huge amounts of
// new code without being able to slowly review, examine and test them.

import (
	"github.com/go-kit/log"
	"github.com/jmespath/go-jmespath"
)

// DiffConfig represents a JSON Stage configuration
type DiffConfig struct {
	// Allow to override settings
}

// diffStage sets extracted data using JMESPath expressions
type diffStage struct {
	cfg         *DiffConfig
	expressions map[string]*jmespath.JMESPath
	logger      log.Logger
}

func validateDiffConfig(*DiffConfig) error {
	return nil
}

// newJSONStage creates a new json pipeline stage from a config.
func newDiffStage(logger log.Logger, cfg DiffConfig) (Stage, error) {
	return &diffStage{
		cfg:    &cfg,
		logger: log.With(logger, "component", "stage", "type", "diff"),
	}, nil
}

func (j *diffStage) Run(in chan Entry) chan Entry {
	out := make(chan Entry)

	go func() {
		defer close(out)
		for e := range in {
			err := j.processEntry(&e)
			if err != nil {
				continue
			}
			out <- e
		}
	}()
	return out
}

func (j *diffStage) processEntry(entry *Entry) error {
	// If a source key is provided, the json stage should process it
	// from the extracted map, otherwise should fallback to the entry
	//input := entry

	// TODO: DO THE MAGIC

	return nil
}

// Name implements Stage
func (j *diffStage) Name() string {
	return StageTypeDiff
}

/*

const (
	memoryProfileName   = "memory"
	allocObjectTypeName = "alloc_objects"
	allocSpaceTypeName  = "alloc_space"
	blockProfileName    = "block"
	contentionsTypeName = "contentions"
	delayTypeName       = "delay"
)

type samples struct {
	timeNanos int64
	samples   []*profile.Sample
}

// deltaProfiles is a helper to compute delta of profiles.
type deltaProfiles struct {
	mtx sync.Mutex
	// todo cleanup sample profiles that are not used anymore using a cleanup goroutine.
	highestSamples map[model.Fingerprint]*samples
}

func newDeltaProfiles() *deltaProfiles {
	return &deltaProfiles{
		highestSamples: make(map[model.Fingerprint]*samples),
	}
}

func (d *deltaProfiles) computeDelta(ps *profile.Profile, lbs model.LabelSet) *profile.Profile {
	// there's no delta to compute for those profile.
	if !isDelta(lbs) {
		return ps
	}

	fingerprint := lbs.Fingerprint()

	d.mtx.Lock()
	defer d.mtx.Unlock()

	// we store all series ref so fetching with one work.
	lastSamples, ok := d.highestSamples[lbs.Fingerprint()]
	if !ok {
		// if we don't have the last profile, we can't compute the delta.
		// so we remove the delta from the list of labels and profiles.
		d.highestSamples[fingerprint] = &samples{
			timeNanos: ps.TimeNanos,
			samples:   copySampleSlice(ps.Sample),
		}

		return nil
	}

	// we have the last profile, we can compute the delta.
	// samples are sorted by stacktrace id.
	// we need to compute the delta for each stacktrace.
	if len(lastSamples.samples) == 0 {
		return ps
	}

	highestSamples, reset := deltaSamples(lastSamples.samples, ps.Sample)
	if reset {
		// if we reset the delta, we can't compute the delta anymore.
		// so we remove the delta from the list of labels and profiles.
		d.highestSamples[fingerprint].samples = copySampleSlice(ps.Sample)
		d.highestSamples[fingerprint].timeNanos = ps.TimeNanos
		return nil
	}

	// remove samples that are all zero
	i := 0
	for _, x := range ps.Samples {
		if x.Value != 0 {
			ps.Samples[i] = x
			i++
		}
	}
	ps.Samples = copySlice(ps.Samples[:i])
	samples := d.highestSamples[ps.SeriesFingerprint]
	samples.samples = highestSamples
	samples.timeNanos = ps.TimeNanos

	return ps
}

func copySampleSlice(s []*profile.Sample) []*profile.Sample {
	if s == nil {
		return nil
	}
	r := make([]*schemav1.Sample, len(s))
	for i := range s {
		r[i] = copySample(s[i])
	}
	return r
}

func copySample(s *schemav1.Sample) *schemav1.Sample {
	if s == nil {
		return nil
	}
	return &schemav1.Sample{
		StacktraceID: s.StacktraceID,
		Value:        s.Value,
	}
}

func isDelta(lbs model.LabelSet) bool {

	switch lbs.Get(phlaremodel.LabelNameDelta) {
	case "false":
		return false
	case "true":
		return true
	}
	if lbs.Get(model.MetricNameLabel) == memoryProfileName {
		ty := lbs.Get(phlaremodel.LabelNameType)
		if ty == allocObjectTypeName || ty == allocSpaceTypeName {
			return true
		}
	}
	return false
}

func deltaSamples(highest, new []*profile.Sample) ([]*profile.Sample, bool) {
	stacktraces := make(map[uint64]*schemav1.Sample)
	for _, h := range highest {
		stacktraces[h.StacktraceID] = h
	}
	for _, n := range new {
		if s, ok := stacktraces[n.StacktraceID]; ok {
			if s.Value <= n.Value {
				newMax := n.Value
				n.Value -= s.Value
				s.Value = newMax
			} else {
				// this is a reset, we can't compute the delta anymore.
				return nil, true
			}
			continue
		}
		highest = append(highest, copySample(n))
	}
	return highest, false
}
*/
