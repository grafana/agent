package queue

import (
	"arena"
	"bytes"
	"github.com/stretchr/testify/require"
	"testing"
	"time"
)

func TestMapMarshal(t *testing.T) {
	input := make(map[string]string)
	input["t1"] = "test1"
	input["t2"] = "test2"

	mem := arena.NewArena()
	defer mem.Free()
	buf := bytes.NewBuffer(nil)
	pushMap(input, buf, mem)
	output := make(map[string]string)
	buf2 := bytes.NewBuffer(buf.Bytes())
	fetchMap(output, buf2, mem)
	require.True(t, input["t1"] == output["t1"])
	require.True(t, input["t2"] == output["t2"])
}

func TestSampleMarshal(t *testing.T) {
	mem := arena.NewArena()
	defer mem.Free()
	input := make([]string, 0)
	input = append(input, "t1", "test1")
	s1 := &sample{
		L:         input,
		TimeStamp: time.Now().Unix(),
		Value:     1.0,
	}
	buf := bytes.NewBuffer(nil)
	s1.Marshal(buf, mem)
	buf2 := bytes.NewBuffer(buf.Bytes())
	s2 := &sample{}
	err := Unmarshal(s2, buf2, mem)
	require.NoError(t, err)
	require.True(t, s1.Value == s2.Value)
	require.True(t, s2.Value == s1.Value)
	require.True(t, len(s1.L) == len(s2.L))
	require.True(t, s1.L[0] == s2.L[0])
	require.True(t, s1.L[1] == s2.L[1])
}
