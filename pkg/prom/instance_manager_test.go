package prom

import (
	"testing"
	"time"

	"github.com/cortexproject/cortex/pkg/util/test"
	"github.com/go-kit/kit/log"
	"github.com/grafana/agent/pkg/prom/instance"
	"github.com/prometheus/prometheus/config"
)

func TestInstanceManager_ApplyConfig(t *testing.T) {
	fact := newMockInstanceFactory()
	spawner := mockInstanceSpawner(fact)

	cm := NewInstanceManager(DefaultInstanceManagerConfig, log.NewNopLogger(), spawner, nil)
	_ = cm.ApplyConfig(instance.Config{Name: "test"})

	test.Poll(t, time.Second, true, func() interface{} {
		return fact.created.Load() == 1
	})

	_ = cm.ApplyConfig(instance.Config{Name: "test", HostFilter: true})

	test.Poll(t, time.Second, true, func() interface{} {
		return fact.created.Load() == 2
	})

	test.Poll(t, time.Second, 1, func() interface{} {
		return len(cm.ListConfigs())
	})
}

func mockInstanceSpawner(fact *mockInstanceFactory) InstanceFactory {
	return func(c instance.Config) (Instance, error) {
		return fact.factory(config.DefaultGlobalConfig, c, "", nil)
	}
}
