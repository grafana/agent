package river

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/grafana/agent/component/discovery"
)

func TestMap(t *testing.T) {
	type test struct {
		Targets []discovery.Target `river:"targets,attr"`
	}
	newTest := test{
		Targets: make([]discovery.Target, 0),
	}
	newTest.Targets = append(newTest.Targets, map[string]string{
		"__address__": "localhost",
	})
	fields, err := convertComponentChild(newTest)
	require.NoError(t, err)
	require.Len(t, fields, 1)
	require.True(t, fields[0].Type == attr)
	require.True(t, fields[0].Value.(*Field).Type == "array")
	require.Len(t, fields[0].Value.(*Field).Value, 1)
	objField := fields[0].Value.(*Field).Value.([]interface{})[0].(*Field)
	require.True(t, objField.Type == "object")
	require.True(t, objField.Value.([]*Field)[0].Key == "__address__")
	require.True(t, objField.Value.([]*Field)[0].Value.(*Field).Type == "string")
	require.True(t, objField.Value.([]*Field)[0].Value.(*Field).Value == "localhost")
}

type TCapsule struct {
	I int
}

// RiverCapsule marks receivers as a capsule.
func (r TCapsule) RiverCapsule() {}

func TestCapsule(t *testing.T) {
	type test struct {
		Receiver *TCapsule `river:"receiver,attr"`
	}

	newTest := test{
		Receiver: &TCapsule{
			I: 1,
		},
	}

	fields, err := convertComponentChild(newTest)
	require.NoError(t, err)
	require.Len(t, fields, 1)
	require.True(t, fields[0].Type == attr)
	require.True(t, fields[0].Name == "receiver")
	require.True(t, fields[0].Value.(*Field).Value == "capsule(\"river.TCapsule\")")
	require.True(t, fields[0].Value.(*Field).Type == "capsule")
}

func TestTime(t *testing.T) {
	type test struct {
		Time time.Time `river:"time,attr"`
	}

	n := time.Now()
	newTest := test{
		Time: n,
	}

	fields, err := convertComponentChild(newTest)
	require.NoError(t, err)
	require.Len(t, fields, 1)
	require.True(t, fields[0].Type == attr)
	require.True(t, fields[0].Name == "time")
	tf := fields[0].Value.(*Field)
	require.True(t, tf.Value == n.Format(time.RFC3339Nano))
	require.True(t, tf.Type == "string")
}
