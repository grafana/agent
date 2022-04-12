package models

import (
	"encoding/json"
	"fmt"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

type TestCase struct {
	values     map[string]float64
	expected   Measurement
	shouldFail bool
}

func testMeasurement(t *testing.T, tcs TestCase) {
	var valuesb strings.Builder
	idx := 0
	for k, v := range tcs.values {
		valuesb.WriteString(fmt.Sprintf("\"%s\": %f", k, v))
		if idx < len(tcs.values)-1 {
			valuesb.WriteByte(',')
		}
		idx++
	}

	payload := fmt.Sprintf(`
{
	"timestamp": "2021-09-30T10:46:17.680Z",
	"values": { %s }
}`, valuesb.String())

	var m Measurement
	err := json.Unmarshal([]byte(payload), &m)

	if !tcs.shouldFail {
		assert.Nil(t, err)
	} else {
		assert.NotNil(t, err)
		return
	}

	assert.Equal(t, len(tcs.expected.Values), len(m.Values))
	for k, v := range m.Values {
		assert.NotEmpty(t, tcs.expected.Values[k])
		assert.Equal(t, v, tcs.expected.Values[k])
	}
}

func TestMeasurements(t *testing.T) {
	testcases := []TestCase{
		{
			values: map[string]float64{
				"lcp": 2500.0,
				"fid": 200.0,
				"cls": 0.15,
			},
			expected: Measurement{
				Values: map[string]float64{
					"lcp": 2500.0,
					"fid": 200.0,
					"cls": 0.15,
				},
			},
		},
		{
			values: map[string]float64{},
			expected: Measurement{
				Values: map[string]float64{},
			},
		},
	}
	for _, tcs := range testcases {
		testMeasurement(t, tcs)
	}
}
