package controller

import (
	"testing"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/stretchr/testify/require"
)

func TestRegister(t *testing.T) {
	r := newRegister(prometheus.NewRegistry())
	err := r.RegisterComponent(prometheus.NewCounter(prometheus.CounterOpts{
		Name: "test_counter",
		Help: "testing a counter",
	}))
	require.NoError(t, err)
	require.Len(t, r.internalCollectors, 1)
	require.True(t, r.UnregisterComponent())
	require.Len(t, r.internalCollectors, 0)
}

func TestRegisterNormal(t *testing.T) {
	r := newRegister(prometheus.NewRegistry())
	err := r.Register(prometheus.NewCounter(prometheus.CounterOpts{
		Name: "test_counter",
		Help: "testing a counter",
	}))
	require.NoError(t, err)
	require.Len(t, r.internalCollectors, 1)
	require.True(t, r.UnregisterComponent())
	require.Len(t, r.internalCollectors, 0)
}

func TestRegisterNormalUnregisterNormal(t *testing.T) {
	r := newRegister(prometheus.NewRegistry())
	testCounter := prometheus.NewCounter(prometheus.CounterOpts{
		Name: "test_counter",
		Help: "testing a counter",
	})
	err := r.Register(testCounter)
	require.NoError(t, err)
	require.Len(t, r.internalCollectors, 1)
	success := r.Unregister(testCounter)
	require.True(t, success)
	require.Len(t, r.internalCollectors, 0)
}
