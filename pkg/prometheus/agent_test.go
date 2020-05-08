package prometheus

import (
	"errors"
	"io"
	"sync"
	"testing"
	"time"

	"github.com/cortexproject/cortex/pkg/util/test"
	"github.com/go-kit/kit/log"
	"github.com/grafana/agent/pkg/prometheus/instance"
	"github.com/prometheus/prometheus/config"
	"github.com/stretchr/testify/require"
	"go.uber.org/atomic"
	"gopkg.in/yaml.v2"
)

func TestNew_ValidatesConfig(t *testing.T) {
	// Zero value of Config is invalid; it needs at least a
	// WAL dir defined
	invalidConfig := Config{}
	_, err := New(invalidConfig, nil)
	require.Error(t, err)
}

func TestConfig_Validate(t *testing.T) {
	valid := Config{
		WALDir: "/tmp/data",
		Configs: []instance.Config{
			{Name: "instance"},
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
			expect:  errors.New("error validating instance 0: missing instance name"),
		},
		{
			name: "duplicate config name",
			mutator: func(c *Config) {
				c.Configs = append(c.Configs,
					instance.Config{Name: "newinstance"},
					instance.Config{Name: "instance"},
				)
			},
			expect: errors.New("prometheus instance names must be unique. found multiple instances with name instance"),
		},
	}

	for _, tc := range tt {
		t.Run(tc.name, func(t *testing.T) {
			cfg := copyConfig(t, valid)
			tc.mutator(&cfg)

			err := cfg.Validate()
			require.Equal(t, tc.expect, err)
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
			{Name: "instance_a"},
			{Name: "instance_b"},
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

	t.Run("wait should be called on each instance", func(t *testing.T) {
		fact.mut.Lock()
		defer fact.mut.Unlock()

		for _, mi := range fact.mocks {
			// Each instance should have wait called on it
			test.Poll(t, time.Millisecond*500, true, func() interface{} {
				return mi.waitCalled.Load()
			})
		}
	})

	t.Run("instances should be restarted when stopped", func(t *testing.T) {
		oldInstances := fact.created.Load()

		fact.mut.Lock()
		for _, mi := range fact.mocks {
			mi.exitErr = io.EOF
			mi.Stop()
		}
		fact.mut.Unlock()

		test.Poll(t, time.Millisecond*500, oldInstances*2, func() interface{} {
			return fact.created.Load()
		})
	})

	t.Run("instances should not be restarted when stopped normally", func(t *testing.T) {
		oldInstances := fact.created.Load()

		fact.mut.Lock()
		for _, mi := range fact.mocks {
			mi.Stop()
		}
		fact.mut.Unlock()

		time.Sleep(time.Millisecond * 100)
		require.Equal(t, oldInstances, fact.created.Load())
	})
}

func TestAgent_Stop(t *testing.T) {
	// Lanch two instances
	cfg := Config{
		WALDir: "/tmp/wal",
		Configs: []instance.Config{
			{Name: "instance_a"},
			{Name: "instance_b"},
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

	oldInstances := fact.created.Load()

	a.Stop()

	time.Sleep(time.Millisecond * 100)
	require.Equal(t, oldInstances, fact.created.Load(), "new instances shuold not have been created")

	fact.mut.Lock()
	for _, mi := range fact.mocks {
		require.True(t, mi.exitCalled.Load())
	}
	fact.mut.Unlock()
}

type mockInstance struct {
	cfg instance.Config

	waitCalled *atomic.Bool
	exitCalled *atomic.Bool

	exited  chan bool
	exitErr error
}

func (i *mockInstance) Wait() error {
	i.waitCalled.Store(true)
	<-i.exited
	return i.exitErr
}

func (i *mockInstance) Config() instance.Config {
	return i.cfg
}

func (i *mockInstance) Stop() {
	if !i.exitCalled.Load() {
		i.exitCalled.Store(true)
		if i.exitErr == nil {
			i.exitErr = instance.ErrInstanceStoppedNormally
		}
		close(i.exited)
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

func (f *mockInstanceFactory) factory(_ config.GlobalConfig, cfg instance.Config, _ string, _ log.Logger) (inst, error) {
	f.created.Add(1)

	f.mut.Lock()
	defer f.mut.Unlock()

	inst := &mockInstance{
		cfg:        cfg,
		exited:     make(chan bool),
		waitCalled: atomic.NewBool(false),
		exitCalled: atomic.NewBool(false),
	}

	f.mocks = append(f.mocks, inst)
	return inst, nil
}
