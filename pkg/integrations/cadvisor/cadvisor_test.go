package cadvisor

import (
	"context"
	"testing"

	"github.com/go-kit/log"
	"github.com/stretchr/testify/assert"
	"gopkg.in/yaml.v3"
)

func run_integration(t *testing.T, cfgStr string) error {
	var cfg Config

	err := yaml.Unmarshal([]byte(cfgStr), &cfg)
	assert.NoError(t, err)
	ig, err := cfg.NewIntegration(log.NewNopLogger())
	assert.NoError(t, err)
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	return ig.Run(ctx)
}

func TestConfig_DockerOnly(t *testing.T) {
	t.Run("docker_only with default configuration is successful", func(t *testing.T) {
		// Run it once with the default config, expecting success.
		defaultCfg := `
docker_only: true
`
		var err error

		assert.NotPanics(t, func() { err = run_integration(t, defaultCfg) })
		assert.ErrorIs(t, err, context.Canceled)
	})

	// 	t.Run("docker_only with empty raw_cgroup_prefix_allowlist panics", func(t *testing.T) {
	// 		// then again when docker_only is true, and raw_cgroup_prefix_allowlist is an empty array,
	// 		// expecting the cadvisor collectors to panic. If this suddenly starts hanging, or does not panic, the default
	// 		// value for raw_cgroup_prefix_allowlist should be returned to a zero value string slice.
	// 		//
	// 		var err error
	// 		panicCfgStr := `
	// docker_only: true
	// raw_cgroup_prefix_allowlist: []
	// `
	// 		assert.Panics(t, func() { err = run_integration(t, panicCfgStr) })
	// 		assert.NoError(t, err)
	// 	})
}
