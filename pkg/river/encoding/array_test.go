package encoding

import (
	"encoding/json"
	"testing"

	"github.com/grafana/agent/pkg/river/internal/value"
	"github.com/stretchr/testify/require"
)

func TestSimpleArray(t *testing.T) {
	intArr := make([]int, 10)
	for i := 0; i < 10; i++ {
		intArr[i] = i
	}
	reqString := `
{
    "type": "array",
    "value": [
        {
            "type": "number",
            "value": 0
        },
        {
            "type": "number",
            "value": 1
        },
        {
            "type": "number",
            "value": 2
        },
        {
            "type": "number",
            "value": 3
        },
        {
            "type": "number",
            "value": 4
        },
        {
            "type": "number",
            "value": 5
        },
        {
            "type": "number",
            "value": 6
        },
        {
            "type": "number",
            "value": 7
        },
        {
            "type": "number",
            "value": 8
        },
        {
            "type": "number",
            "value": 9
        }
    ]
}
`
	arrF, err := newRiverArray(value.Encode(intArr))
	require.NoError(t, err)
	bb, err := json.Marshal(arrF)
	require.NoError(t, err)
	require.JSONEq(t, reqString, string(bb))
}

func TestStructArray(t *testing.T) {
	type sa struct {
		Age int `river:"age,attr"`
	}
	intArr := make([]*sa, 10)
	for i := 0; i < 10; i++ {
		intArr[i] = &sa{Age: i}
	}

	reqString := `
{
    "type": "array",
    "value": [
        {
            "type": "object",
            "value": [
                {
                    "value": {
                        "type": "number",
                        "value": 0
                    },
                    "key": "age"
                }
            ]
        },
        {
            "type": "object",
            "value": [
                {
                    "value": {
                        "type": "number",
                        "value": 1
                    },
                    "key": "age"
                }
            ]
        },
        {
            "type": "object",
            "value": [
                {
                    "value": {
                        "type": "number",
                        "value": 2
                    },
                    "key": "age"
                }
            ]
        },
        {
            "type": "object",
            "value": [
                {
                    "value": {
                        "type": "number",
                        "value": 3
                    },
                    "key": "age"
                }
            ]
        },
        {
            "type": "object",
            "value": [
                {
                    "value": {
                        "type": "number",
                        "value": 4
                    },
                    "key": "age"
                }
            ]
        },
        {
            "type": "object",
            "value": [
                {
                    "value": {
                        "type": "number",
                        "value": 5
                    },
                    "key": "age"
                }
            ]
        },
        {
            "type": "object",
            "value": [
                {
                    "value": {
                        "type": "number",
                        "value": 6
                    },
                    "key": "age"
                }
            ]
        },
        {
            "type": "object",
            "value": [
                {
                    "value": {
                        "type": "number",
                        "value": 7
                    },
                    "key": "age"
                }
            ]
        },
        {
            "type": "object",
            "value": [
                {
                    "value": {
                        "type": "number",
                        "value": 8
                    },
                    "key": "age"
                }
            ]
        },
        {
            "type": "object",
            "value": [
                {
                    "value": {
                        "type": "number",
                        "value": 9
                    },
                    "key": "age"
                }
            ]
        }
    ]
}
`
	arrF, err := newRiverArray(value.Encode(intArr))
	require.NoError(t, err)
	bb, err := json.Marshal(arrF)
	require.NoError(t, err)
	require.JSONEq(t, reqString, string(bb))
}
