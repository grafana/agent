package jfr

import (
	//"github.com/grafana/pyroscope/pkg/distributor/model"

	"fmt"
	"io"
	"time"

	"github.com/grafana/agent/component/discovery"
	"github.com/grafana/agent/component/pyroscope"
	jfrparser "github.com/grafana/jfr-parser/parser"
	"github.com/grafana/jfr-parser/parser/types"
	"github.com/prometheus/prometheus/model/labels"
)

const (
	sampleTypeCPU  = 0
	sampleTypeWall = 1

	sampleTypeInTLAB = 2

	sampleTypeOutTLAB = 3

	sampleTypeLock = 4

	sampleTypeThreadPark = 5

	sampleTypeLiveObject = 6
)

// labels labels.Labels, samples []*RawSample
type PushRequest struct {
	Labels  labels.Labels
	Samples []*pyroscope.RawSample
}

func ParseJFR(body []byte, metadata Metadata) (requests []PushRequest, err error) {
	defer func() {
		if r := recover(); r != nil {
			err = fmt.Errorf("jfr parser panic: %v", r)
		}
	}()
	parser := jfrparser.NewParser(body, jfrparser.Options{
		SymbolProcessor: processSymbols,
	})

	var event string

	builders := newJfrPprofBuilders(parser, metadata)

	var values = [2]int64{1, 0}

	for {
		typ, err := parser.ParseEvent()
		if err != nil {
			if err == io.EOF {
				break
			}
			return nil, fmt.Errorf("jfr parser ParseEvent error: %w", err)
		}

		switch typ {
		case parser.TypeMap.T_EXECUTION_SAMPLE:
			ts := parser.GetThreadState(parser.ExecutionSample.State)
			if ts != nil && ts.Name != "STATE_SLEEPING" {
				builders.addStacktrace(sampleTypeCPU, parser.ExecutionSample.StackTrace, values[:1])
			}
			if event == "wall" {
				builders.addStacktrace(sampleTypeWall, parser.ExecutionSample.StackTrace, values[:1])
			}
		case parser.TypeMap.T_ALLOC_IN_NEW_TLAB:
			values[1] = int64(parser.ObjectAllocationInNewTLAB.TlabSize)
			builders.addStacktrace(sampleTypeInTLAB, parser.ObjectAllocationInNewTLAB.StackTrace, values[:2])
		case parser.TypeMap.T_ALLOC_OUTSIDE_TLAB:
			values[1] = int64(parser.ObjectAllocationOutsideTLAB.AllocationSize)
			builders.addStacktrace(sampleTypeOutTLAB, parser.ObjectAllocationOutsideTLAB.StackTrace, values[:2])
		case parser.TypeMap.T_MONITOR_ENTER:
			values[1] = int64(parser.JavaMonitorEnter.Duration)
			builders.addStacktrace(sampleTypeLock, parser.JavaMonitorEnter.StackTrace, values[:2])
		case parser.TypeMap.T_THREAD_PARK:
			values[1] = int64(parser.ThreadPark.Duration)
			builders.addStacktrace(sampleTypeThreadPark, parser.ThreadPark.StackTrace, values[:2])
		case parser.TypeMap.T_LIVE_OBJECT:
			builders.addStacktrace(sampleTypeLiveObject, parser.LiveObject.StackTrace, values[:1])
		case parser.TypeMap.T_ACTIVE_SETTING:
			if parser.ActiveSetting.Name == "event" {
				event = parser.ActiveSetting.Value
			}

		}
	}

	requests, err = builders.build(event)

	return requests, err
}

type Metadata struct {
	StartTime  time.Time
	EndTime    time.Time
	SampleRate int
	Target     discovery.Target
}

func newJfrPprofBuilders(p *jfrparser.Parser, metadata Metadata) *jfrPprofBuilders {
	st := metadata.StartTime.UnixNano()
	et := metadata.EndTime.UnixNano()
	var period int64
	if metadata.SampleRate == 0 {
		period = 0
	} else {
		period = 1e9 / int64(metadata.SampleRate)
	}
	res := &jfrPprofBuilders{
		timeNanos:          st,
		durationNanos:      et - st,
		period:             period,
		parser:             p,
		sampleType2Builder: make(map[int64]*ProfileBuilder, 6),
		metadata:           metadata,
	}

	return res
}

type jfrPprofBuilders struct {
	timeNanos     int64
	durationNanos int64

	parser             *jfrparser.Parser
	sampleType2Builder map[int64]*ProfileBuilder

	period   int64
	metadata Metadata
}

func (b *jfrPprofBuilders) addStacktrace(sampleType int64, ref types.StackTraceRef, values []int64) {
	e := b.sampleType2Builder[sampleType]
	if e == nil {
		e = NewProfileBuilder(b.timeNanos)
		b.sampleType2Builder[sampleType] = e
	}
	st := b.parser.GetStacktrace(ref)
	if st == nil {
		return
	}

	addValues := func(dst []int64) {
		mul := 1
		if sampleType == sampleTypeCPU || sampleType == sampleTypeWall {
			mul = int(b.period)
		}
		for i, value := range values {
			dst[i] += value * int64(mul)
		}
	}

	sample := e.FindExternalSample(uint32(ref))
	if sample != nil {
		addValues(sample.Value)
		return
	}

	locations := make([]uint64, 0, len(st.Frames))
	for i := 0; i < len(st.Frames); i++ {
		f := st.Frames[i]
		loc, found := e.FindLocationByExternalID(uint32(f.Method))
		if found {
			locations = append(locations, loc)
			continue
		}
		m := b.parser.GetMethod(f.Method)
		if m != nil {

			cls := b.parser.GetClass(m.Type)
			if cls != nil {
				clsName := b.parser.GetSymbolString(cls.Name)
				methodName := b.parser.GetSymbolString(m.Name)
				frame := clsName + "." + methodName
				loc = e.AddExternalFunction(frame, uint32(f.Method))
				locations = append(locations, loc)
			}
			//todo remove Scratch field from the Method
		}
	}
	vs := make([]int64, len(values))
	addValues(vs)
	e.AddExternalSample(locations, vs, uint32(ref))
}

func (b *jfrPprofBuilders) build(event string) ([]PushRequest, error) {
	defer func() {
		for _, builder := range b.sampleType2Builder {
			builder.Profile.ReturnToVTPool()
		}
		b.sampleType2Builder = nil
	}()
	profiles := make([]PushRequest, 0, len(b.sampleType2Builder))

	for sampleType, e := range b.sampleType2Builder {
		//for _, e := range entries {
		e.TimeNanos = b.timeNanos
		e.DurationNanos = b.durationNanos
		metric := ""
		switch sampleType {
		case sampleTypeCPU:
			e.AddSampleType("cpu", "nanoseconds")
			e.PeriodType("cpu", "nanoseconds")
			metric = "process_cpu"
		case sampleTypeWall:
			e.AddSampleType("wall", "nanoseconds")
			e.PeriodType("wall", "nanoseconds")
			metric = "wall"
		case sampleTypeInTLAB:
			e.AddSampleType("alloc_in_new_tlab_objects", "count")
			e.AddSampleType("alloc_in_new_tlab_bytes", "bytes")
			e.PeriodType("space", "bytes")
			metric = "memory"
		case sampleTypeOutTLAB:
			e.AddSampleType("alloc_outside_tlab_objects", "count")
			e.AddSampleType("alloc_outside_tlab_bytes", "bytes")
			e.PeriodType("space", "bytes")
			metric = "memory"
		case sampleTypeLock:
			e.AddSampleType("contentions", "count")
			e.AddSampleType("delay", "nanoseconds")
			e.PeriodType("mutex", "count")
			metric = "mutex"
		case sampleTypeThreadPark:
			e.AddSampleType("contentions", "count")
			e.AddSampleType("delay", "nanoseconds")
			e.PeriodType("block", "count")
			metric = "block"
		case sampleTypeLiveObject:
			e.AddSampleType("live", "count")
			e.PeriodType("objects", "count")
			metric = "memory"
		}
		ls := labels.NewBuilder(make(labels.Labels, 0, len(b.metadata.Target)+5))
		ls.Set(labels.MetricName, metric)
		ls.Set("__delta__", "false")
		ls.Set("jfr_event", event)
		ls.Set("pyroscope_spy", "grafana-agent.java")
		for k, v := range b.metadata.Target {
			ls.Set(k, v)
		}
		prof, err := e.Profile.MarshalVT()
		if err != nil {
			return nil, fmt.Errorf("marshal profile error: %w", err)
		}
		profiles = append(profiles, PushRequest{
			Labels: ls.Labels(),
			Samples: []*pyroscope.RawSample{
				{
					RawProfile: prof,
				},
			},
		})
		//}
	}
	return profiles, nil
}
