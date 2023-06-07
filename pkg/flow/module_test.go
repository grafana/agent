package flow

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestIDCollision(t *testing.T) {
	nc := newModuleController(&moduleControllerOptions{
		Logger:         nil,
		Tracer:         nil,
		Clusterer:      nil,
		Reg:            nil,
		DataPath:       "",
		HTTPListenAddr: "",
		HTTPPath:       "",
		DialFunc:       nil,
	})
	m, err := nc.NewModule("t1", nil)
	require.NoError(t, err)
	require.NotNil(t, m)
	m, err = nc.NewModule("t1", nil)
	require.Error(t, err)
	require.Nil(t, m)
}
