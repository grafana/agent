package promsdprocessor

import (
	"testing"

	"github.com/go-kit/log"
	"github.com/prometheus/common/model"
	"github.com/prometheus/prometheus/discovery/targetgroup"
	"github.com/prometheus/prometheus/pkg/relabel"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.opentelemetry.io/collector/consumer/consumertest"
	"go.opentelemetry.io/collector/model/pdata"
	semconv "go.opentelemetry.io/collector/model/semconv/v1.6.1"
)

func TestSyncGroups(t *testing.T) {
	tests := []struct {
		name        string
		jobToSync   string
		relabelCfgs map[string][]*relabel.Config
		targets     []model.LabelSet
		expected    map[string]model.LabelSet
	}{
		{
			name:        "empty",
			jobToSync:   "",
			relabelCfgs: map[string][]*relabel.Config{},
			targets:     []model.LabelSet{},
			expected:    map[string]model.LabelSet{},
		},
		{
			name:      "no relabeling",
			jobToSync: "job",
			relabelCfgs: map[string][]*relabel.Config{
				"job": {},
			},
			targets: []model.LabelSet{
				{
					"__address__": "127.0.0.1",
				},
			},
			expected: map[string]model.LabelSet{
				"127.0.0.1": {},
			},
		},
		{
			name:      "strip port",
			jobToSync: "job",
			relabelCfgs: map[string][]*relabel.Config{
				"job": {},
			},
			targets: []model.LabelSet{
				{
					"__address__": "127.0.0.1:8888",
					"label":       "val",
				},
			},
			expected: map[string]model.LabelSet{
				"127.0.0.1": {
					"label": "val",
				},
			},
		},
		{
			name:      "passthrough",
			jobToSync: "job",
			relabelCfgs: map[string][]*relabel.Config{
				"job": {},
			},
			targets: []model.LabelSet{
				{
					"__address__": "127.0.0.1",
					"label":       "val",
				},
			},
			expected: map[string]model.LabelSet{
				"127.0.0.1": {
					"label": "val",
				},
			},
		},
		{
			name:      "ignore metadata",
			jobToSync: "job",
			relabelCfgs: map[string][]*relabel.Config{
				"job": {},
			},
			targets: []model.LabelSet{
				{
					"__address__": "127.0.0.1",
					"__ignore":    "ignore",
				},
			},
			expected: map[string]model.LabelSet{
				"127.0.0.1": {},
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			groups := []*targetgroup.Group{
				{
					Targets: tc.targets,
				},
			}

			p := &promServiceDiscoProcessor{
				logger:         log.NewNopLogger(),
				relabelConfigs: tc.relabelCfgs,
			}

			hostLabels := make(map[string]model.LabelSet)
			p.syncGroups(tc.jobToSync, groups, hostLabels)

			assert.Equal(t, tc.expected, hostLabels)
		})
	}
}

func TestOperationType(t *testing.T) {
	const (
		attrKey = "key"
		attrIP  = "1.1.1.1"
	)
	testCases := []struct {
		name            string
		operationType   string
		attributeExists bool
		newValue        string
		expectedValue   string
	}{
		{
			name:            "Upsert updates the attribute already exists",
			operationType:   OperationTypeUpsert,
			attributeExists: true,
			newValue:        "new-value",
			expectedValue:   "new-value",
		},
		{
			name:            "Update updates the attribute already exists",
			operationType:   OperationTypeUpdate,
			attributeExists: true,
			newValue:        "new-value",
			expectedValue:   "new-value",
		},
		{
			name:            "Insert does not update the attribute if it's already present",
			operationType:   OperationTypeInsert,
			attributeExists: true,
			newValue:        "new-value",
			expectedValue:   "old-value",
		},
		{
			name:            "Upsert updates the attribute if it isn't present",
			operationType:   OperationTypeUpsert,
			attributeExists: false,
			newValue:        "new-value",
			expectedValue:   "new-value",
		},
		{
			name:            "Update updates the attribute already exists",
			operationType:   OperationTypeUpdate,
			attributeExists: false,
			newValue:        "new-value",
			expectedValue:   "",
		},
		{
			name:            "Insert updates the attribute if it isn't present",
			operationType:   OperationTypeInsert,
			attributeExists: false,
			newValue:        "new-value",
			expectedValue:   "new-value",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			mockProcessor := new(consumertest.TracesSink)
			p, err := newTraceProcessor(mockProcessor, tc.operationType, nil)
			require.NoError(t, err)

			attrValue := pdata.NewAttributeValueString("old-value")
			ipAttrValue := pdata.NewAttributeValueString(attrIP)

			attrMap := pdata.NewAttributeMap()
			if tc.attributeExists {
				attrMap.Insert(attrKey, attrValue)
			}
			attrMap.Insert(semconv.AttributeNetHostIP, ipAttrValue)

			hostLabels := map[string]model.LabelSet{
				attrIP: {
					attrKey: model.LabelValue(tc.newValue),
				},
			}
			p.(*promServiceDiscoProcessor).hostLabels = hostLabels
			p.(*promServiceDiscoProcessor).processAttributes(attrMap)

			actualAttrValue, _ := attrMap.Get(attrKey)
			assert.Equal(t, tc.expectedValue, actualAttrValue.StringVal())
		})
	}
}
