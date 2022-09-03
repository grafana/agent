package encoding

import (
	"encoding/json"
	"strings"
	"testing"

	"github.com/grafana/agent/pkg/river/internal/value"
	"github.com/stretchr/testify/require"
)

const number = "number"
const age = "age"

func TestSimpleArray(t *testing.T) {
	intArr := make([]int, 10)
	for i := 0; i < 10; i++ {
		intArr[i] = i
	}

	arrF, err := newArray(value.Encode(intArr))
	require.NoError(t, err)
	require.Len(t, arrF.valueFields, 10)
	require.True(t, arrF.valueFields[0].Value == 0)
	require.True(t, arrF.valueFields[9].Value == 9)
	bb, err := json.Marshal(arrF)
	require.NoError(t, err)
	require.True(t, strings.Contains(string(bb), "9"))
}

func TestStructArray(t *testing.T) {
	type sa struct {
		Age int `river:"age,attr"`
	}
	intArr := make([]*sa, 10)
	for i := 0; i < 10; i++ {
		intArr[i] = &sa{Age: i}
	}

	arrF, err := newArray(value.Encode(intArr))
	require.NoError(t, err)
	require.Len(t, arrF.structFields, 10)
	require.True(t, arrF.structFields[0].Type == object)
	require.Len(t, arrF.structFields[0].Value, 1)
	require.True(t, arrF.structFields[0].Value[0].Key == age)
	require.True(t, arrF.structFields[0].Value[0].Value.(*ValueField).Type == number)
	require.True(t, arrF.structFields[0].Value[0].Value.(*ValueField).Value == 0)
	bb, err := json.Marshal(arrF)
	require.NoError(t, err)
	require.True(t, strings.Contains(string(bb), "9"))
}
