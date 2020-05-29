package prom

import (
	"context"
	"testing"
	"time"

	"github.com/cortexproject/cortex/pkg/util/test"
	"github.com/grafana/agent/pkg/prom/instance"
	"github.com/prometheus/prometheus/config"
)

func TestInstanceManager_ApplyConfig(t *testing.T) {
	fact := newMockInstanceFactory()
	spawner := mockInstanceSpawner(fact)

	cm := NewInstanceManager(spawner)
	cm.ApplyConfig(instance.Config{Name: "test"})

	test.Poll(t, time.Second, true, func() interface{} {
		return fact.created.Load() == 1
	})

	cm.ApplyConfig(instance.Config{Name: "test", HostFilter: true})

	test.Poll(t, time.Second, true, func() interface{} {
		return fact.created.Load() == 2
	})

	test.Poll(t, time.Second, 1, func() interface{} {
		return len(cm.ListConfigs())
	})
}

func mockInstanceSpawner(fact *mockInstanceFactory) func(context.Context, instance.Config) {
	return func(ctx context.Context, c instance.Config) {
		inst, err := fact.factory(config.DefaultGlobalConfig, c, "", nil)
		if err != nil {
			return
		}

		_ = inst.Run(ctx)
	}
}
