package scrape

import (
	"context"

	"github.com/grafana/agent/component/pyroscope"
	"github.com/prometheus/prometheus/model/labels"
)

type godeltaprofProbe struct {
	godeltaprof bool
	path        string
}

func godeltaprofProbes(profileType string, path string) []godeltaprofProbe {
	if profileType == pprofMemory && path == "/debug/pprof/allocs" {
		return []godeltaprofProbe{
			{true, "/debug/pprof/delta_heap"},
			{false, "/debug/pprof/allocs"},
		}
	}
	if profileType == pprofBlock && path == "/debug/pprof/block" {
		return []godeltaprofProbe{
			{true, "/debug/pprof/delta_block"},
			{false, "/debug/pprof/block"},
		}
	}
	if profileType == pprofMutex && path == "/debug/pprof/mutex" {
		return []godeltaprofProbe{
			{true, "/debug/pprof/delta_mutex"},
			{false, "/debug/pprof/mutex"},
		}
	}
	return []godeltaprofProbe{
		{false, path},
	}
}

func newAppender(probe godeltaprofProbe, t *scrapeLoop) pyroscope.Appender {
	appender := t.appendable.Appender()
	if probe.godeltaprof {
		return newGodeltaprofAppender(appender)
	} else {
		return NewDeltaAppender(appender, t.allLabels)
	}
}

type godeltaprofAppender struct {
	appender pyroscope.Appender
}

func newGodeltaprofAppender(appender pyroscope.Appender) *godeltaprofAppender {
	return &godeltaprofAppender{
		appender: appender,
	}
}

func (d *godeltaprofAppender) Append(ctx context.Context, lbs labels.Labels, samples []*pyroscope.RawSample) error {
	// Notify the server that this profile is a delta profile and we don't need to compute the delta again.
	lbsBuilder := labels.NewBuilder(lbs)
	lbsBuilder.Set(pyroscope.LabelNameDelta, "false")
	return d.appender.Append(ctx, lbsBuilder.Labels(), samples)
}
