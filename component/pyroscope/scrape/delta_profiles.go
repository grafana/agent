package scrape

import (
	"bytes"
	"compress/gzip"
	"context"
	"fmt"
	"io"

	"github.com/grafana/agent/component/pyroscope"
	"github.com/grafana/agent/component/pyroscope/scrape/internal/fastdelta"
	"github.com/prometheus/common/model"
	"github.com/prometheus/prometheus/model/labels"
)

var deltaProfiles map[string][]fastdelta.ValueType = map[string][]fastdelta.ValueType{
	pprofMemory: {
		{Type: "alloc_objects", Unit: "count"},
		{Type: "alloc_space", Unit: "bytes"},
	},
	pprofMutex: {
		{Type: "contentions", Unit: "count"},
		{Type: "delay", Unit: "nanoseconds"},
	},
	pprofBlock: {
		{Type: "contentions", Unit: "count"},
		{Type: "delay", Unit: "nanoseconds"},
	},
}

type DeltaProfiler interface {
	Delta(p []byte, out io.Writer) error
}

func NewDeltaAppender(appender pyroscope.Appender, labels labels.Labels) pyroscope.Appender {
	types, ok := deltaProfiles[labels.Get(model.MetricNameLabel)]
	if !ok {
		// for profiles that we don't need to produce delta, just return the appender
		return appender
	}
	delta := &deltaAppender{
		appender: appender,
		delta:    fastdelta.NewDeltaComputer(types...),
		gzw:      gzip.NewWriter(nil),
	}
	delta.reset()
	return delta
}

type deltaAppender struct {
	appender pyroscope.Appender
	delta    DeltaProfiler

	buf bytes.Buffer
	gzr gzip.Reader
	gzw *gzip.Writer

	// true if we have seen at least one sample
	initialized bool
}

func (d *deltaAppender) reset() {
	d.buf.Reset()
	d.gzw.Reset(&d.buf)
}

func (d *deltaAppender) Append(ctx context.Context, lbs labels.Labels, samples []*pyroscope.RawSample) error {
	// Notify the server that this profile is a delta profile and we don't need to compute the delta again.
	lbsBuilder := labels.NewBuilder(lbs)
	lbsBuilder.Set(pyroscope.LabelNameDelta, "false")
	for _, sample := range samples {
		data, err := d.computeDelta(sample.RawProfile)
		if err != nil {
			return err
		}
		// The first sample should be skipped because we don't have a previous sample to compute delta with.
		if !d.initialized {
			d.initialized = true
			continue
		}
		if err := d.appender.Append(ctx, lbsBuilder.Labels(), []*pyroscope.RawSample{{RawProfile: data}}); err != nil {
			return err
		}
	}
	return nil
}

// computeDelta computes the delta between the given profile and the last
// data is uncompressed if it is gzip compressed.
// The returned data is always gzip compressed.
func (d *deltaAppender) computeDelta(data []byte) (b []byte, err error) {
	if isGzipData(data) {
		if err := d.gzr.Reset(bytes.NewReader(data)); err != nil {
			return nil, err
		}
		data, err = io.ReadAll(&d.gzr)
		if err != nil {
			return nil, fmt.Errorf("decompressing profile: %v", err)
		}
	}

	d.reset()

	if err = d.delta.Delta(data, d.gzw); err != nil {
		return nil, fmt.Errorf("computing delta: %v", err)
	}
	if err := d.gzw.Close(); err != nil {
		return nil, fmt.Errorf("closing gzip writer: %v", err)
	}
	// The returned slice will be retained in case the profile upload fails,
	// so we need to return a copy of the buffer's bytes to avoid a data
	// race.
	b = make([]byte, len(d.buf.Bytes()))
	copy(b, d.buf.Bytes())
	return b, nil
}

func isGzipData(data []byte) bool {
	return bytes.HasPrefix(data, []byte{0x1f, 0x8b})
}
