package prom

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"testing"
	"time"

	"github.com/cortexproject/cortex/pkg/util/test"
	"github.com/go-kit/kit/log"
	"github.com/grafana/agent/pkg/prom/instance"
	"github.com/prometheus/prometheus/config"
	"github.com/stretchr/testify/require"
	"go.uber.org/atomic"
	"gopkg.in/yaml.v2"
)

func TestConfig_Validate(t *testing.T) {
	valid := Config{
		WALDir: "/tmp/data",
		Configs: []instance.Config{
			makeInstanceConfig("instance"),
		},
	}

	tt := []struct {
		name    string
		mutator func(c *Config)
		expect  error
	}{
		{
			name:    "complete config should be valid",
			mutator: func(c *Config) {},
			expect:  nil,
		},
		{
			name:    "no wal dir",
			mutator: func(c *Config) { c.WALDir = "" },
			expect:  errors.New("no wal_directory configured"),
		},
		{
			name:    "missing instance name",
			mutator: func(c *Config) { c.Configs[0].Name = "" },
			expect:  errors.New("error validating instance at index 0: missing instance name"),
		},
		{
			name: "duplicate config name",
			mutator: func(c *Config) {
				c.Configs = append(c.Configs,
					makeInstanceConfig("newinstance"),
					makeInstanceConfig("instance"),
				)
			},
			expect: errors.New("prometheus instance names must be unique. found multiple instances with name instance"),
		},
	}

	for _, tc := range tt {
		t.Run(tc.name, func(t *testing.T) {
			cfg := copyConfig(t, valid)
			tc.mutator(&cfg)

			err := cfg.ApplyDefaults()
			if tc.expect == nil {
				require.NoError(t, err)
			} else {
				require.EqualError(t, err, tc.expect.Error())
			}
		})
	}
}

func copyConfig(t *testing.T, c Config) Config {
	bb, err := yaml.Marshal(c)
	require.NoError(t, err)

	var cp Config
	err = yaml.Unmarshal(bb, &cp)
	require.NoError(t, err)
	return cp
}

func TestAgent(t *testing.T) {
	// Lanch two instances
	cfg := Config{
		WALDir: "/tmp/wal",
		Configs: []instance.Config{
			makeInstanceConfig("instance_a"),
			makeInstanceConfig("instance_b"),
		},
		InstanceRestartBackoff: time.Duration(0),
	}

	fact := newMockInstanceFactory()

	a, err := newAgent(cfg, log.NewNopLogger(), fact.factory)
	require.NoError(t, err)

	test.Poll(t, time.Second*30, true, func() interface{} {
		if fact.created == nil {
			return false
		}
		return fact.created.Load() == 2 && len(a.cm.processes) == 2
	})

	t.Run("instances should be running", func(t *testing.T) {
		for _, mi := range fact.Mocks() {
			// Each instance should have wait called on it
			test.Poll(t, time.Millisecond*500, true, func() interface{} {
				return mi.running.Load()
			})
		}
	})

	t.Run("instances should be restarted when stopped", func(t *testing.T) {
		for _, mi := range fact.Mocks() {
			test.Poll(t, time.Millisecond*500, int64(1), func() interface{} {
				return mi.startedCount.Load()
			})
		}

		for _, mi := range fact.Mocks() {
			mi.err <- fmt.Errorf("really bad error")
		}

		for _, mi := range fact.Mocks() {
			test.Poll(t, time.Millisecond*500, int64(2), func() interface{} {
				return mi.startedCount.Load()
			})
		}
	})

	t.Run("instances should not be restarted when stopped normally", func(t *testing.T) {
		oldInstances := fact.created.Load()

		for _, mi := range fact.Mocks() {
			// Simulate a stop by saying the context was cancelled
			mi.err <- context.Canceled
		}

		time.Sleep(time.Millisecond * 100)
		require.Equal(t, oldInstances, fact.created.Load())
	})
}

func TestAgent_Stop(t *testing.T) {
	// Lanch two instances
	cfg := Config{
		WALDir: "/tmp/wal",
		Configs: []instance.Config{
			makeInstanceConfig("instance_a"),
			makeInstanceConfig("instance_b"),
		},
		InstanceRestartBackoff: time.Duration(0),
	}

	fact := newMockInstanceFactory()

	a, err := newAgent(cfg, log.NewNopLogger(), fact.factory)
	require.NoError(t, err)

	test.Poll(t, time.Second*30, true, func() interface{} {
		if fact.created == nil {
			return false
		}
		return fact.created.Load() == 2 && len(a.cm.processes) == 2
	})

	a.Stop()

	time.Sleep(time.Millisecond * 100)

	for _, mi := range fact.Mocks() {
		require.False(t, mi.running.Load(), "instance should not have been restarted")
	}
}

type mockInstance struct {
	cfg instance.Config

	err          chan error
	startedCount *atomic.Int64
	running      *atomic.Bool
}

func (i *mockInstance) Run(ctx context.Context) error {
	i.startedCount.Inc()
	i.running.Store(true)
	defer i.running.Store(false)

	select {
	case <-ctx.Done():
		return ctx.Err()
	case err := <-i.err:
		return err
	}
}

type mockInstanceFactory struct {
	mut   sync.Mutex
	mocks []*mockInstance

	created *atomic.Int64
}

func newMockInstanceFactory() *mockInstanceFactory {
	return &mockInstanceFactory{created: atomic.NewInt64(0)}
}

func (f *mockInstanceFactory) Mocks() []*mockInstance {
	f.mut.Lock()
	defer f.mut.Unlock()
	return f.mocks
}

func (f *mockInstanceFactory) factory(_ config.GlobalConfig, cfg instance.Config, _ string, _ log.Logger) (inst, error) {
	f.created.Add(1)

	f.mut.Lock()
	defer f.mut.Unlock()

	inst := &mockInstance{
		cfg:          cfg,
		running:      atomic.NewBool(false),
		startedCount: atomic.NewInt64(0),
		err:          make(chan error),
	}

	f.mocks = append(f.mocks, inst)
	return inst, nil
}

func makeInstanceConfig(name string) instance.Config {
	cfg := instance.DefaultConfig
	cfg.Name = name
	return cfg
}
