package redis_test

import (
	"fmt"
	"testing"

	"github.com/grafana/agent/pkg/autodiscovery/redis"
	"github.com/stretchr/testify/require"
)

func TestRun(t *testing.T) {
	m, err := redis.New()
	require.NoError(t, err)
	res, err := m.Run()
	require.NoError(t, err)
	// fmt.Println(res, err)
	for _, target := range res.MetricsTargets {
		fmt.Println(target.RiverString())
		// fmt.Println(target.Labels())
	}
	for _, target := range res.LogfileTargets {
		fmt.Println(target.RiverString())
		// fmt.Println(target.Labels())
	}
}
