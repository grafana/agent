package automaticloggingprocessor

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.opentelemetry.io/collector/consumer/pdata"
)

func TestSpanKeyVals(t *testing.T) {
	tests := []struct {
		spanName  string
		spanAttrs map[string]pdata.AttributeValue
		spanStart uint64
		spanEnd   uint64
		svc       string
		cfg       AutomaticLoggingConfig
		expected  []interface{}
	}{
		{
			expected: []interface{}{
				"name", "",
				"dur", "0ns",
				"svc", "",
				"status", pdata.StatusCode(0),
			},
		},
		{
			spanName: "test",
			expected: []interface{}{
				"name", "test",
				"dur", "0ns",
				"svc", "",
				"status", pdata.StatusCode(0),
			},
		},
		{
			svc: "test",
			expected: []interface{}{
				"name", "",
				"dur", "0ns",
				"svc", "test",
				"status", pdata.StatusCode(0),
			},
		},
		{
			spanEnd: 10,
			expected: []interface{}{
				"name", "",
				"dur", "10ns",
				"svc", "",
				"status", pdata.StatusCode(0),
			},
		},
		{
			spanStart: 10,
			spanEnd:   100,
			expected: []interface{}{
				"name", "",
				"dur", "90ns",
				"svc", "",
				"status", pdata.StatusCode(0),
			},
		},
		{
			spanAttrs: map[string]pdata.AttributeValue{
				"xstr": pdata.NewAttributeValueString("test"),
			},
			expected: []interface{}{
				"name", "",
				"dur", "0ns",
				"svc", "",
				"status", pdata.StatusCode(0),
			},
		},
		{
			spanAttrs: map[string]pdata.AttributeValue{
				"xstr": pdata.NewAttributeValueString("test"),
			},
			cfg: AutomaticLoggingConfig{
				SpanAttributes: []string{"xstr"},
			},
			expected: []interface{}{
				"name", "",
				"dur", "0ns",
				"svc", "",
				"status", pdata.StatusCode(0),
				"xstr", "test",
			},
		},
	}

	for _, tc := range tests {
		p, err := newTraceProcessor(&automaticLoggingProcessor{}, &tc.cfg)
		require.NoError(t, err)

		span := pdata.NewSpan()
		span.SetName(tc.spanName)
		span.Attributes().InitFromMap(tc.spanAttrs).Sort()
		span.SetStartTime(pdata.TimestampUnixNano(tc.spanStart))
		span.SetEndTime(pdata.TimestampUnixNano(tc.spanEnd))

		actual := p.(*automaticLoggingProcessor).spanKeyVals(span, tc.svc)
		assert.Equal(t, tc.expected, actual)
	}
}

func TestProcessKeyVals(t *testing.T) {
	tests := []struct {
		processAttrs map[string]pdata.AttributeValue
		svc          string
		cfg          AutomaticLoggingConfig
		expected     []interface{}
	}{
		{
			expected: []interface{}{
				"svc", "",
			},
		},
		{
			processAttrs: map[string]pdata.AttributeValue{
				"xstr": pdata.NewAttributeValueString("test"),
			},
			expected: []interface{}{
				"svc", "",
			},
		},
		{
			processAttrs: map[string]pdata.AttributeValue{
				"xstr": pdata.NewAttributeValueString("test"),
			},
			cfg: AutomaticLoggingConfig{
				ProcessAttributes: []string{"xstr"},
			},
			expected: []interface{}{
				"svc", "",
				"xstr", "test",
			},
		},
	}

	for _, tc := range tests {
		p, err := newTraceProcessor(&automaticLoggingProcessor{}, &tc.cfg)
		require.NoError(t, err)

		process := pdata.NewResource()
		process.Attributes().InitFromMap(tc.processAttrs).Sort()

		actual := p.(*automaticLoggingProcessor).processKeyVals(process, tc.svc)
		assert.Equal(t, tc.expected, actual)
	}
}
