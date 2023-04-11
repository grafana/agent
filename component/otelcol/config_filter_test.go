package otelcol_test

import (
	"testing"

	"github.com/grafana/agent/component/otelcol"
	"github.com/stretchr/testify/require"
	"k8s.io/utils/pointer"
)

func TestConvertMatchProperties(t *testing.T) {
	inputMatchProps := otelcol.MatchProperties{
		MatchType: "strict",
		RegexpConfig: &otelcol.RegexpConfig{
			CacheEnabled:       false,
			CacheMaxNumEntries: 2,
		},
		Services:         []string{"svcA", "svcB"},
		SpanNames:        []string{"login.*"},
		LogBodies:        []string{"AUTH.*"},
		LogSeverityTexts: []string{"debug.*"},
		LogSeverityNumber: &otelcol.LogSeverityNumberMatchProperties{
			Min:            2,
			MatchUndefined: true,
		},
		MetricNames: []string{"metric1"},
		Attributes: []otelcol.Attribute{
			{
				Key:   "attr1",
				Value: 5,
			},
			{
				Key:   "attr2",
				Value: "asdf",
			},
			{
				Key:   "attr2",
				Value: false,
			},
		},
		Resources: []otelcol.Attribute{
			{
				Key:   "attr1",
				Value: 5,
			},
		},
		Libraries: []otelcol.InstrumentationLibrary{
			{
				Name:    "mongo-java-driver",
				Version: pointer.String("3.8.0"),
			},
		},
		SpanKinds: []string{"span1"},
	}

	expectedMatchProps := map[string]interface{}{
		"attributes": []interface{}{
			map[string]interface{}{
				"key":   "attr1",
				"value": 5,
			},
			map[string]interface{}{
				"key":   "attr2",
				"value": "asdf",
			},
			map[string]interface{}{
				"key":   "attr2",
				"value": false,
			},
		},
		"libraries": []interface{}{
			map[string]interface{}{
				"name":    "mongo-java-driver",
				"version": "3.8.0",
			},
		},
		"log_bodies": []string{"AUTH.*"},
		"log_severity_number": map[string]interface{}{
			"min":             int32(2),
			"match_undefined": true,
		},
		"log_severity_texts": []string{
			"debug.*",
		},
		"match_type":   "strict",
		"metric_names": []string{"metric1"},
		"regexp": map[string]interface{}{
			"cacheenabled":       false,
			"cachemaxnumentries": 2,
		},
		"resources": []interface{}{
			map[string]interface{}{
				"key":   "attr1",
				"value": 5,
			},
		},
		"services":   []string{"svcA", "svcB"},
		"span_kinds": []string{"span1"},
		"span_names": []string{"login.*"},
	}

	tests := []struct {
		testName            string
		inputMatchConfig    otelcol.MatchConfig
		expectedMatchConfig map[string]interface{}
	}{
		{
			testName:            "TestConvertEmpty",
			inputMatchConfig:    otelcol.MatchConfig{},
			expectedMatchConfig: make(map[string]interface{}),
		},
		{
			testName: "TestConvertMandatory",
			inputMatchConfig: otelcol.MatchConfig{
				Include: &otelcol.MatchProperties{
					MatchType: "strict",
				},
			},
			expectedMatchConfig: map[string]interface{}{
				"include": map[string]interface{}{
					"match_type": "strict",
				},
			},
		},
		{
			testName: "TestAllOptsInclExcl",
			inputMatchConfig: otelcol.MatchConfig{
				Include: &inputMatchProps,
				Exclude: &inputMatchProps,
			},
			expectedMatchConfig: map[string]interface{}{
				"include": expectedMatchProps,
				"exclude": expectedMatchProps,
			},
		},
		{
			testName: "TestAllOptsIncl",
			inputMatchConfig: otelcol.MatchConfig{
				Include: &inputMatchProps,
			},
			expectedMatchConfig: map[string]interface{}{
				"include": expectedMatchProps,
			},
		},
		{
			testName: "TestAllOptsExcl",
			inputMatchConfig: otelcol.MatchConfig{
				Exclude: &inputMatchProps,
			},
			expectedMatchConfig: map[string]interface{}{
				"exclude": expectedMatchProps,
			},
		},
	}

	for _, tt := range tests {
		if matchConf := tt.inputMatchConfig.Exclude; matchConf != nil {
			result := matchConf.Convert()
			require.Equal(t, tt.expectedMatchConfig["exclude"], result)
		} else {
			require.Empty(t, tt.expectedMatchConfig["exclude"])
		}

		if matchConf := tt.inputMatchConfig.Include; matchConf != nil {
			result := matchConf.Convert()
			require.Equal(t, tt.expectedMatchConfig["include"], result)
		} else {
			require.Empty(t, tt.expectedMatchConfig["include"])
		}
	}
}
