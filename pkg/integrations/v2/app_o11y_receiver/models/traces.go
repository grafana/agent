package models

import (
	"fmt"

	otlp "go.opentelemetry.io/collector/model/otlp"
	otelpdata "go.opentelemetry.io/collector/model/pdata"

	"github.com/grafana/agent/pkg/integrations/v2/app_o11y_receiver/utils"
)

// TraceContext holds trace id and span id associated to an entity (log, exception, measurement...).
type TraceContext struct {
	TraceID string `json:"trace_id"`
	SpanID  string `json:"span_id"`
}

// KeyVal representation of the trace context object.
func (tc TraceContext) KeyVal() *utils.KeyVal {
	retv := utils.NewKeyVal()
	utils.KeyValAdd(retv, "traceID", tc.TraceID)
	utils.KeyValAdd(retv, "spanID", tc.SpanID)
	return retv
}

// Traces wraps the otel traces model.
type Traces struct {
	otelpdata.Traces
}

// UnmarshalJSON unmarshals Traces model.
func (t *Traces) UnmarshalJSON(b []byte) error {
	unmarshaler := otlp.NewJSONTracesUnmarshaler()
	td, err := unmarshaler.UnmarshalTraces(b)
	if err != nil {
		return err
	}
	*t = Traces{td}
	return nil
}

// MarshalJSON marshals Traces model to json.
func (t Traces) MarshalJSON() ([]byte, error) {
	marshaler := otlp.NewJSONTracesMarshaler()
	return marshaler.MarshalTraces(t.Traces)
}

// SpanSlice unpacks Traces entity into a slice of Spans.
func (t Traces) SpanSlice() []otelpdata.Span {
	spans := make([]otelpdata.Span, 0)
	rss := t.ResourceSpans()
	for i := 0; i < rss.Len(); i++ {
		rs := rss.At(i)
		ilss := rs.InstrumentationLibrarySpans()
		for j := 0; j < ilss.Len(); j++ {
			s := ilss.At(j).Spans()
			for si := 0; si < s.Len(); si++ {
				spans = append(spans, s.At(si))
			}
		}
	}
	return spans
}

// SpanToKeyVal returns KeyVal representation of a Span.
func SpanToKeyVal(s otelpdata.Span) *utils.KeyVal {
	kv := utils.NewKeyVal()
	if s.StartTimestamp() > 0 {
		utils.KeyValAdd(kv, "timestamp", s.StartTimestamp().AsTime().String())
	}
	if s.EndTimestamp() > 0 {
		utils.KeyValAdd(kv, "end_timestamp", s.StartTimestamp().AsTime().String())
	}
	utils.KeyValAdd(kv, "kind", "span")
	utils.KeyValAdd(kv, "traceID", s.TraceID().HexString())
	utils.KeyValAdd(kv, "spanID", s.SpanID().HexString())
	utils.KeyValAdd(kv, "span_kind", s.Kind().String())
	utils.KeyValAdd(kv, "name", s.Name())
	utils.KeyValAdd(kv, "parent_spanID", s.ParentSpanID().HexString())
	s.Attributes().Range(func(k string, v otelpdata.AttributeValue) bool {
		utils.KeyValAdd(kv, "attr_"+k, fmt.Sprintf("%v", v))
		return true
	})

	return kv
}
