package instance

import (
	"errors"
	"fmt"
	"strings"
	"testing"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/testutil"
	"github.com/stretchr/testify/require"
)

func TestCountingManager(t *testing.T) {
	mockConfigs := make(map[string]Config)

	mock := &MockManager{
		ListInstancesFunc: func() map[string]ManagedInstance { return nil },
		ListConfigsFunc: func() map[string]Config {
			return mockConfigs
		},
		ApplyConfigFunc: func(c Config) error {
			mockConfigs[c.Name] = c
			return nil
		},
		DeleteConfigFunc: func(name string) error {
			if _, ok := mockConfigs[name]; !ok {
				return errors.New("config does not exist")
			}
			delete(mockConfigs, name)
			return nil
		},
	}

	reg := prometheus.NewRegistry()
	cm := NewCountingManager(reg, mock)

	// Pull values from the registry and assert that our gauge has the
	// expected value.
	requireGaugeValue := func(value int) {
		expect := fmt.Sprintf(`
		# HELP agent_prometheus_active_configs Current number of active configs being used by the agent.
		# TYPE agent_prometheus_active_configs gauge
		agent_prometheus_active_configs %d
		`, value)

		r := strings.NewReader(expect)
		require.NoError(t, testutil.GatherAndCompare(reg, r))
	}

	requireGaugeValue(0)

	// Apply two diferent configs, but each config twice. The gauge should
	// only be set to 2.
	_ = cm.ApplyConfig(Config{Name: "config-a"})
	_ = cm.ApplyConfig(Config{Name: "config-a"})
	_ = cm.ApplyConfig(Config{Name: "config-b"})
	_ = cm.ApplyConfig(Config{Name: "config-b"})
	requireGaugeValue(2)

	// Deleting a config that doesn't exist shouldn't change the gauge.
	_ = cm.DeleteConfig("config-nil")
	requireGaugeValue(2)

	// Deleting a config that does exist _should_ change the gauge.
	_ = cm.DeleteConfig("config-b")
	requireGaugeValue(1)
}
