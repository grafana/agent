package otelcol_test

import (
	"testing"

	"github.com/grafana/agent/component/otelcol"
	"github.com/stretchr/testify/require"
)

func TestConvertAttrAction(t *testing.T) {
	inputActions := otelcol.AttrActionKeyValueSlice{
		{
			Action: "insert",
			Value:  123,
			Key:    "attribute1",
		},
		{
			Action: "delete",
			Key:    "attribute2",
		},
		{
			Action: "upsert",
			Value:  true,
			Key:    "attribute3",
		},
	}

	expectedActions := []interface{}{
		map[string]interface{}{
			"action": "insert",
			"key":    "attribute1",
			"value":  123,
		},
		map[string]interface{}{
			"action": "delete",
			"key":    "attribute2",
			"value":  interface{}(nil),
		},
		map[string]interface{}{
			"action": "upsert",
			"key":    "attribute3",
			"value":  true,
		},
	}

	result := inputActions.Convert()
	require.Equal(t, expectedActions, result)
}
