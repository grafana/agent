package automaticloggingprocessor

import (
	"testing"
	"time"

	"github.com/grafana/agent/pkg/logs"
	"github.com/grafana/agent/pkg/util"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.opentelemetry.io/collector/consumer/pdata"
	"gopkg.in/yaml.v3"
)

func TestSpanKeyVals(t *testing.T) {
	tests := []struct {
		spanName  string
		spanAttrs map[string]pdata.AttributeValue
		spanStart time.Time
		spanEnd   time.Time
		cfg       AutomaticLoggingConfig
		expected  []interface{}
	}{
		{
			expected: []interface{}{
				"span", "",
				"dur", "0ns",
				"status", pdata.StatusCode(0),
			},
		},
		{
			spanName: "test",
			expected: []interface{}{
				"span", "test",
				"dur", "0ns",
				"status", pdata.StatusCode(0),
			},
		},
		{
			expected: []interface{}{
				"span", "",
				"dur", "0ns",
				"status", pdata.StatusCode(0),
			},
		},
		{
			spanStart: time.Unix(0, 0),
			spanEnd:   time.Unix(0, 10),
			expected: []interface{}{
				"span", "",
				"dur", "10ns",
				"status", pdata.StatusCode(0),
			},
		},
		{
			spanStart: time.Unix(0, 10),
			spanEnd:   time.Unix(0, 100),
			expected: []interface{}{
				"span", "",
				"dur", "90ns",
				"status", pdata.StatusCode(0),
			},
		},
		{
			spanAttrs: map[string]pdata.AttributeValue{
				"xstr": pdata.NewAttributeValueString("test"),
			},
			expected: []interface{}{
				"span", "",
				"dur", "0ns",
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
				"span", "",
				"dur", "0ns",
				"status", pdata.StatusCode(0),
				"xstr", "test",
			},
		},
		{
			cfg: AutomaticLoggingConfig{
				Overrides: OverrideConfig{
					SpanNameKey: "a",
					DurationKey: "c",
					StatusKey:   "d",
				},
			},
			expected: []interface{}{
				"a", "",
				"c", "0ns",
				"d", pdata.StatusCode(0),
			},
		},
	}

	for _, tc := range tests {
		tc.cfg.Backend = BackendStdout
		tc.cfg.Spans = true
		p, err := newTraceProcessor(&automaticLoggingProcessor{}, &tc.cfg)
		require.NoError(t, err)

		span := pdata.NewSpan()
		span.SetName(tc.spanName)
		span.Attributes().InitFromMap(tc.spanAttrs).Sort()
		span.SetStartTimestamp(pdata.TimestampFromTime(tc.spanStart))
		span.SetEndTimestamp(pdata.TimestampFromTime(tc.spanEnd))

		actual := p.(*automaticLoggingProcessor).spanKeyVals(span)
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
		tc.cfg.Backend = BackendStdout
		tc.cfg.Spans = true
		p, err := newTraceProcessor(&automaticLoggingProcessor{}, &tc.cfg)
		require.NoError(t, err)

		process := pdata.NewResource()
		process.Attributes().InitFromMap(tc.processAttrs).Sort()

		actual := p.(*automaticLoggingProcessor).processKeyVals(process, tc.svc)
		assert.Equal(t, tc.expected, actual)
	}
}

func TestBadConfigs(t *testing.T) {
	tests := []struct {
		cfg *AutomaticLoggingConfig
	}{
		{
			cfg: &AutomaticLoggingConfig{},
		},
		{
			cfg: &AutomaticLoggingConfig{
				Backend: "blarg",
				Spans:   true,
			},
		},
		{
			cfg: &AutomaticLoggingConfig{
				Backend: "logs",
			},
		},
		{
			cfg: &AutomaticLoggingConfig{
				Backend: "loki",
			},
		},
		{
			cfg: &AutomaticLoggingConfig{
				Backend: "stdout",
			},
		},
	}

	for _, tc := range tests {
		p, err := newTraceProcessor(&automaticLoggingProcessor{}, tc.cfg)
		require.Error(t, err)
		require.Nil(t, p)
	}
}

func TestLogToStdoutSet(t *testing.T) {
	cfg := &AutomaticLoggingConfig{
		Backend: BackendStdout,
		Spans:   true,
	}

	p, err := newTraceProcessor(&automaticLoggingProcessor{}, cfg)
	require.NoError(t, err)
	require.True(t, p.(*automaticLoggingProcessor).logToStdout)

	cfg = &AutomaticLoggingConfig{
		Backend: BackendLogs,
		Spans:   true,
	}

	p, err = newTraceProcessor(&automaticLoggingProcessor{}, cfg)
	require.NoError(t, err)
	require.False(t, p.(*automaticLoggingProcessor).logToStdout)
}

func TestDefaults(t *testing.T) {
	cfg := &AutomaticLoggingConfig{
		Spans: true,
	}

	p, err := newTraceProcessor(&automaticLoggingProcessor{}, cfg)
	require.NoError(t, err)
	require.Equal(t, BackendStdout, p.(*automaticLoggingProcessor).cfg.Backend)
	require.Equal(t, defaultTimeout, p.(*automaticLoggingProcessor).cfg.Timeout)
	require.True(t, p.(*automaticLoggingProcessor).logToStdout)

	require.Equal(t, defaultLogsTag, p.(*automaticLoggingProcessor).cfg.Overrides.LogsTag)
	require.Equal(t, defaultServiceKey, p.(*automaticLoggingProcessor).cfg.Overrides.ServiceKey)
	require.Equal(t, defaultSpanNameKey, p.(*automaticLoggingProcessor).cfg.Overrides.SpanNameKey)
	require.Equal(t, defaultStatusKey, p.(*automaticLoggingProcessor).cfg.Overrides.StatusKey)
	require.Equal(t, defaultDurationKey, p.(*automaticLoggingProcessor).cfg.Overrides.DurationKey)
	require.Equal(t, defaultTraceIDKey, p.(*automaticLoggingProcessor).cfg.Overrides.TraceIDKey)
}

func TestLokiNameMigration(t *testing.T) {
	logsConfig := &logs.Config{
		Configs: []*logs.InstanceConfig{{Name: "default"}},
	}

	input := util.Untab(`
		backend: loki
		loki_name: default
		overrides:
			loki_tag: tempo
	`)
	expect := util.Untab(`
		backend: logs_instance
		logs_instance_name: default
		overrides:
			logs_instance_tag: tempo
	`)

	var cfg AutomaticLoggingConfig
	require.NoError(t, yaml.Unmarshal([]byte(input), &cfg))
	require.NoError(t, cfg.Validate(logsConfig))

	bb, err := yaml.Marshal(cfg)
	require.NoError(t, err)
	require.YAMLEq(t, expect, string(bb))
}
