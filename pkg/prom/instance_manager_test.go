package prom

import (
	"context"
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/cortexproject/cortex/pkg/util/test"
	"github.com/go-kit/kit/log"
	"github.com/grafana/agent/pkg/prom/instance"
	"github.com/prometheus/prometheus/config"
	"github.com/prometheus/prometheus/scrape"
	"github.com/stretchr/testify/require"
)

func TestInstanceManager_ApplyConfig(t *testing.T) {
	fact := newFakeInstanceFactory()
	spawner := fakeInstanceSpawner(fact)

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

func TestInstanceManager_ApplyConfig_DynamicUpdates(t *testing.T) {
	logger := log.NewLogfmtLogger(log.NewSyncWriter(os.Stderr))

	baseMock := mockInstance{
		RunFunc: func(ctx context.Context) error {
			logger.Log("msg", "starting an instance")
			<-ctx.Done()
			return nil
		},
		UpdateFunc: func(c instance.Config) error {
			return nil
		},
		TargetsActiveFunc: func() map[string][]*scrape.Target {
			return nil
		},
	}

	t.Run("dynamic update successful", func(t *testing.T) {
		spawnedCount := 0
		spawner := func(c instance.Config) (Instance, error) {
			spawnedCount++
			return baseMock, nil
		}

		cm := NewInstanceManager(DefaultInstanceManagerConfig, logger, spawner, nil)

		for i := 0; i < 10; i++ {
			err := cm.ApplyConfig(instance.Config{Name: "test"})
			require.NoError(t, err)
		}

		require.Equal(t, 1, spawnedCount)
	})

	t.Run("dynamic update unsuccessful", func(t *testing.T) {
		spawnedCount := 0
		spawner := func(c instance.Config) (Instance, error) {
			spawnedCount++

			newMock := baseMock
			newMock.UpdateFunc = func(c instance.Config) error {
				return instance.ErrInvalidUpdate{
					Inner: fmt.Errorf("cannot dynamically update for testing reasons"),
				}
			}
			return newMock, nil
		}

		cm := NewInstanceManager(DefaultInstanceManagerConfig, logger, spawner, nil)

		for i := 0; i < 10; i++ {
			err := cm.ApplyConfig(instance.Config{Name: "test"})
			require.NoError(t, err)
		}

		require.Equal(t, 10, spawnedCount)
	})

	t.Run("dynamic update errored", func(t *testing.T) {
		spawnedCount := 0
		spawner := func(c instance.Config) (Instance, error) {
			spawnedCount++

			newMock := baseMock
			newMock.UpdateFunc = func(c instance.Config) error {
				return fmt.Errorf("something really bad happened")
			}
			return newMock, nil
		}

		cm := NewInstanceManager(DefaultInstanceManagerConfig, logger, spawner, nil)

		// Creation should succeed
		err := cm.ApplyConfig(instance.Config{Name: "test"})
		require.NoError(t, err)

		// ...but the update should fail
		err = cm.ApplyConfig(instance.Config{Name: "test"})
		require.Error(t, err, "something really bad happened")
		require.Equal(t, 1, spawnedCount)
	})
}

func fakeInstanceSpawner(fact *fakeInstanceFactory) InstanceFactory {
	return func(c instance.Config) (Instance, error) {
		return fact.factory(config.DefaultGlobalConfig, c, "", nil)
	}
}

type mockInstance struct {
	RunFunc           func(ctx context.Context) error
	UpdateFunc        func(c instance.Config) error
	TargetsActiveFunc func() map[string][]*scrape.Target
}

func (m mockInstance) Run(ctx context.Context) error {
	if m.RunFunc != nil {
		return m.RunFunc(ctx)
	}
	panic("RunFunc not provided")
}

func (m mockInstance) Update(c instance.Config) error {
	if m.UpdateFunc != nil {
		return m.UpdateFunc(c)
	}
	panic("UpdateFunc not provided")
}

func (m mockInstance) TargetsActive() map[string][]*scrape.Target {
	if m.TargetsActiveFunc != nil {
		return m.TargetsActiveFunc()
	}
	panic("TargetsActiveFunc not provided")
}
