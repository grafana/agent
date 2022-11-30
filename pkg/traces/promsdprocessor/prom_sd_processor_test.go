package promsdprocessor

import (
	"context"
	"net"
	"testing"

	"github.com/go-kit/log"
	"github.com/prometheus/common/model"
	"github.com/prometheus/prometheus/discovery/targetgroup"
	"github.com/prometheus/prometheus/model/relabel"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.opentelemetry.io/collector/client"
	"go.opentelemetry.io/collector/consumer/consumertest"
	"go.opentelemetry.io/collector/pdata/pcommon"
	semconv "go.opentelemetry.io/collector/semconv/v1.6.1"
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
			p, err := newTraceProcessor(mockProcessor, tc.operationType, nil, nil)
			require.NoError(t, err)

			attrMap := pcommon.NewMap()
			if tc.attributeExists {
				attrMap.PutStr(attrKey, "old-value")
			}
			attrMap.PutStr(semconv.AttributeNetHostIP, attrIP)

			hostLabels := map[string]model.LabelSet{
				attrIP: {
					attrKey: model.LabelValue(tc.newValue),
				},
			}
			p.(*promServiceDiscoProcessor).hostLabels = hostLabels
			p.(*promServiceDiscoProcessor).processAttributes(context.TODO(), attrMap)

			actualAttrValue, _ := attrMap.Get(attrKey)
			assert.Equal(t, tc.expectedValue, actualAttrValue.Str())
		})
	}
}

func TestPodAssociation(t *testing.T) {
	const ipStr = "1.1.1.1"

	testCases := []struct {
		name            string
		podAssociations []string
		ctxFn           func(t *testing.T) context.Context
		attrMapFn       func(t *testing.T) pcommon.Map
		expectedIP      string
	}{
		{
			name: "connection IP (HTTP & gRPC)",
			ctxFn: func(t *testing.T) context.Context {
				info := client.Info{
					Addr: &net.IPAddr{
						IP: net.ParseIP(net.ParseIP(ipStr).String()),
					},
				}
				return client.NewContext(context.Background(), info)
			},
			attrMapFn:  func(*testing.T) pcommon.Map { return pcommon.NewMap() },
			expectedIP: ipStr,
		},
		{
			name: "connection IP that includes a port number",
			ctxFn: func(t *testing.T) context.Context {
				info := client.Info{
					Addr: &net.TCPAddr{
						IP:   net.ParseIP(net.ParseIP(ipStr).String()),
						Port: 1234,
					},
				}
				return client.NewContext(context.Background(), info)
			},
			attrMapFn:  func(*testing.T) pcommon.Map { return pcommon.NewMap() },
			expectedIP: ipStr,
		},
		{
			name:            "connection IP is empty",
			podAssociations: []string{podAssociationConnectionIP},
			ctxFn: func(t *testing.T) context.Context {
				c := client.FromContext(context.Background())
				return client.NewContext(context.Background(), c)
			},
			attrMapFn:  func(*testing.T) pcommon.Map { return pcommon.NewMap() },
			expectedIP: "",
		},
		{
			name:  "ip attribute",
			ctxFn: func(t *testing.T) context.Context { return context.Background() },
			attrMapFn: func(*testing.T) pcommon.Map {
				attrMap := pcommon.NewMap()
				attrMap.PutStr("ip", ipStr)
				return attrMap
			},
			expectedIP: ipStr,
		},
		{
			name:  "net.host.ip attribute",
			ctxFn: func(t *testing.T) context.Context { return context.Background() },
			attrMapFn: func(*testing.T) pcommon.Map {
				attrMap := pcommon.NewMap()
				attrMap.PutStr(semconv.AttributeNetHostIP, ipStr)
				return attrMap
			},
			expectedIP: ipStr,
		},
		{
			name:  "k8s ip attribute",
			ctxFn: func(t *testing.T) context.Context { return context.Background() },
			attrMapFn: func(*testing.T) pcommon.Map {
				attrMap := pcommon.NewMap()
				attrMap.PutStr("k8s.pod.ip", ipStr)
				return attrMap
			},
			expectedIP: ipStr,
		},
		{
			name:  "ip from hostname",
			ctxFn: func(t *testing.T) context.Context { return context.Background() },
			attrMapFn: func(*testing.T) pcommon.Map {
				attrMap := pcommon.NewMap()
				attrMap.PutStr(semconv.AttributeHostName, ipStr)
				return attrMap
			},
			expectedIP: ipStr,
		},
		{
			name: "uses attr before context (default associations)",
			ctxFn: func(t *testing.T) context.Context {
				info := client.Info{
					Addr: &net.IPAddr{
						IP: net.ParseIP("2.2.2.2"),
					},
				}
				return client.NewContext(context.Background(), info)
			},
			attrMapFn: func(*testing.T) pcommon.Map {
				attrMap := pcommon.NewMap()
				attrMap.PutStr(semconv.AttributeNetHostIP, ipStr)
				return attrMap
			},
			expectedIP: ipStr,
		},
		{
			name:  "uses attr before hostname (default associations)",
			ctxFn: func(t *testing.T) context.Context { return context.Background() },
			attrMapFn: func(*testing.T) pcommon.Map {
				attrMap := pcommon.NewMap()
				attrMap.PutStr(semconv.AttributeNetHostIP, ipStr)
				attrMap.PutStr(semconv.AttributeHostName, "3.3.3.3")
				return attrMap
			},
			expectedIP: ipStr,
		},
		{
			name:            "ip attribute but not as pod association",
			podAssociations: []string{podAssociationk8sIPLabel},
			ctxFn:           func(t *testing.T) context.Context { return context.Background() },
			attrMapFn: func(*testing.T) pcommon.Map {
				attrMap := pcommon.NewMap()
				attrMap.PutStr("ip", ipStr)
				return attrMap
			},
			expectedIP: "",
		},
		{
			name:            "uses hostname before attribute (reverse order from default)",
			podAssociations: []string{podAssociationHostnameLabel, podAssociationOTelIPLabel},
			ctxFn:           func(t *testing.T) context.Context { return context.Background() },
			attrMapFn: func(*testing.T) pcommon.Map {
				attrMap := pcommon.NewMap()
				attrMap.PutStr(semconv.AttributeNetHostIP, "3.3.3.3")
				attrMap.PutStr(semconv.AttributeHostName, ipStr)
				return attrMap
			},
			expectedIP: ipStr,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			mockProcessor := new(consumertest.TracesSink)
			p, err := newTraceProcessor(mockProcessor, "", tc.podAssociations, nil)
			require.NoError(t, err)

			ip := p.(*promServiceDiscoProcessor).getPodIP(tc.ctxFn(t), tc.attrMapFn(t))
			assert.Equal(t, tc.expectedIP, ip)
		})
	}
}
