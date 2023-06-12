package common_test

import (
	"testing"

	"github.com/grafana/agent/component/discovery"
	"github.com/grafana/agent/converter/internal/common"
	"github.com/grafana/agent/pkg/river/token/builder"
	"github.com/stretchr/testify/require"
)

func TestOptionalSecret_Write(t *testing.T) {
	tt := []struct {
		name   string
		value  common.ConvertTargets
		expect string
	}{
		{
			name: "nil",
			value: common.ConvertTargets{
				Targets: nil,
			},
			expect: `[]`,
		},
		{
			name: "empty",
			value: common.ConvertTargets{
				Targets: []discovery.Target{{}},
			},
			expect: `[]`,
		},
		{
			name: "__address__ key",
			value: common.ConvertTargets{
				Targets: []discovery.Target{{"__address__": "testing"}},
			},
			expect: `[{
	__address__ = "testing",
}]`,
		},
		{
			name: "multiple __address__ key",
			value: common.ConvertTargets{
				Targets: []discovery.Target{{"__address__": "testing"}, {"__address__": "testing2"}},
			},
			expect: `concat([{
	__address__ = "testing",
}],
	[{
		__address__ = "testing2",
	}])`,
		},
		{
			name: "non __address__ key",
			value: common.ConvertTargets{
				Targets: []discovery.Target{{"key": ""}},
			},
			expect: `key`,
		},
		{
			name: "multiple non __address__ key",
			value: common.ConvertTargets{
				Targets: []discovery.Target{{"key": ""}, {"key2": ""}},
			},
			expect: `concat(key,
	key2)`,
		},
		{
			name: "both key types",
			value: common.ConvertTargets{
				Targets: []discovery.Target{{"__address__": "testing"}, {"key": ""}},
			},
			expect: `concat([{
	__address__ = "testing",
}],
	key)`,
		},
	}

	for _, tc := range tt {
		t.Run(tc.name, func(t *testing.T) {
			be := builder.NewExpr()
			be.SetValue(tc.value)
			require.Equal(t, tc.expect, string(be.Bytes()))
		})
	}
}
