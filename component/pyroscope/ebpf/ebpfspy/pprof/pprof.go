//go:build linux

package pprof

import (
	"bytes"
	"compress/gzip"
	"fmt"
	"io"
	"sync"
	"time"

	"github.com/google/pprof/profile"
	"github.com/prometheus/prometheus/model/labels"
	"go.uber.org/atomic"
)

var (
	gzipWriterPoolCounter atomic.Int64
	gzipWriterPool        = sync.Pool{
		New: func() any {
			gzipWriterPoolCounter.Inc()
			res, err := gzip.NewWriterLevel(io.Discard, gzip.BestSpeed)
			if err != nil {
				panic(err)
			}
			return res
		},
	}
)

type ProfileBuilders struct {
	Builders   map[uint64]*ProfileBuilder
	SampleRate int
}

func NewProfileBuilders(sampleRate int) *ProfileBuilders {
	return &ProfileBuilders{Builders: make(map[uint64]*ProfileBuilder), SampleRate: sampleRate}
}

func (b ProfileBuilders) BuilderForTarget(hash uint64, labels labels.Labels) *ProfileBuilder {
	res := b.Builders[hash]
	if res != nil {
		return res
	}
	buf := bytes.NewBuffer(nil)
	prevCounter := gzipWriterPoolCounter.Load()
	gzipWriter := gzipWriterPool.Get().(*gzip.Writer)
	newCounter := gzipWriterPoolCounter.Load()
	if prevCounter != newCounter {
		fmt.Printf("------------>>> gzipWriterPoolCounter: %d %d\n", newCounter, prevCounter)
	}
	gzipWriter.Reset(buf)
	builder := &ProfileBuilder{
		buf:         buf,
		gzipWritter: gzipWriter,
		locations:   make(map[string]*profile.Location),
		functions:   make(map[string]*profile.Function),
		Labels:      labels,
		profile: &profile.Profile{
			Mapping: []*profile.Mapping{
				{
					ID: 1,
				},
			},
			SampleType: []*profile.ValueType{{Type: "cpu", Unit: "nanoseconds"}},
			Period:     time.Second.Nanoseconds() / int64(b.SampleRate),
			PeriodType: &profile.ValueType{Type: "cpu", Unit: "nanoseconds"},
		},
	}
	res = builder
	b.Builders[hash] = res
	return res
}

type ProfileBuilder struct {
	locations   map[string]*profile.Location
	functions   map[string]*profile.Function
	profile     *profile.Profile
	Labels      labels.Labels
	buf         *bytes.Buffer
	gzipWritter *gzip.Writer
}

func (p *ProfileBuilder) AddSample(stacktrace []string, value uint64) {
	sample := &profile.Sample{
		Value: []int64{int64(value) * p.profile.Period},
	}
	for _, s := range stacktrace {
		loc := p.addLocation(s)
		sample.Location = append(sample.Location, loc)
	}
	p.profile.Sample = append(p.profile.Sample, sample)
}

func (p *ProfileBuilder) addLocation(function string) *profile.Location {
	loc, ok := p.locations[function]
	if ok {
		return loc
	}

	id := uint64(len(p.profile.Location) + 1)
	loc = &profile.Location{
		ID:      id,
		Mapping: p.profile.Mapping[0],
		Line: []profile.Line{
			{
				Function: p.addFunction(function),
			},
		},
	}
	p.profile.Location = append(p.profile.Location, loc)
	p.locations[function] = loc
	return loc
}

func (p *ProfileBuilder) addFunction(function string) *profile.Function {
	f, ok := p.functions[function]
	if ok {
		return f
	}

	id := uint64(len(p.profile.Function) + 1)
	f = &profile.Function{
		ID:   id,
		Name: function,
	}
	p.profile.Function = append(p.profile.Function, f)
	p.functions[function] = f
	return f
}

func (p *ProfileBuilder) Build() ([]byte, error) {
	defer func() {
		p.gzipWritter.Reset(io.Discard)
		gzipWriterPool.Put(p.gzipWritter)
	}()
	err := p.profile.WriteUncompressed(p.gzipWritter)
	if err != nil {
		return nil, fmt.Errorf("ebpf profile encode %w", err)
	}
	err = p.gzipWritter.Close()
	if err != nil {
		return nil, fmt.Errorf("ebpf profile encode %w", err)
	}
	return p.buf.Bytes(), nil
}
