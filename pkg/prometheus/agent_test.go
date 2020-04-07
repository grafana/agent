package prometheus

import (
	"errors"
	"io"
	"testing"
	"time"

	"github.com/cortexproject/cortex/pkg/util/test"
	"github.com/go-kit/kit/log"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/common/model"
	"github.com/prometheus/prometheus/config"
	"github.com/stretchr/testify/require"
	"go.uber.org/atomic"
	"gopkg.in/yaml.v2"
)

func TestConfig_UnmarshalYAML_Defaults(t *testing.T) {
	var c Config

	// Static compilation test: *Config should be an unmarshaller
	var _ yaml.Unmarshaler = &c

	err := yaml.Unmarshal([]byte("{}"), &c)
	require.NoError(t, err)
	require.Equal(t, config.DefaultGlobalConfig, c.Global)

	t.Run("Invalid YAML", func(t *testing.T) {
		var c Config
		err := yaml.Unmarshal([]byte("<h1>I'm pretty sure this is XML.</h1>"), &c)
		require.Error(t, err)
	})
}

func TestConfig_UnmarshalYAML_DefaultsOverride(t *testing.T) {
	data := `global: { scrape_timeout: '33s' }`

	expect := config.GlobalConfig{
		ScrapeInterval:     model.Duration(1 * time.Minute),
		ScrapeTimeout:      model.Duration(33 * time.Second),
		EvaluationInterval: model.Duration(1 * time.Minute),
	}

	var c Config
	err := yaml.Unmarshal([]byte(data), &c)
	require.NoError(t, err)
	require.Equal(t, expect, c.Global)
}

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
		Configs: []InstanceConfig{
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
					InstanceConfig{Name: "newinstance"},
					InstanceConfig{Name: "instance"},
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
		Configs: []InstanceConfig{
			{Name: "instance_a"},
			{Name: "instance_b"},
		},
		InstanceRestartBackoff: time.Duration(0),
	}

	var fact mockInstanceFactory

	a, err := newAgent(cfg, log.NewNopLogger(), fact.factory)
	require.NoError(t, err)
	require.Equal(t, fact.created.Load(), int64(2))
	require.Len(t, a.instances, 2)

	t.Run("wait should be called on each instance", func(t *testing.T) {
		a.forAllInstances(func(_ int, i instance) {
			mi := i.(*mockInstance)

			// Each instance should have wait called on it
			test.Poll(t, time.Millisecond*500, true, func() interface{} {
				return mi.waitCalled.Load()
			})
		})
	})

	t.Run("instances should be restarted when stopped", func(t *testing.T) {
		oldInstances := fact.created.Load()

		a.forAllInstances(func(_ int, i instance) {
			// Set abnormal error so the instance is restarted
			mi := i.(*mockInstance)
			mi.exitErr = io.EOF

			i.Stop()
		})

		test.Poll(t, time.Millisecond*500, oldInstances*2, func() interface{} {
			return fact.created.Load()
		})
	})

	t.Run("instances should not be restarted when stopped normally", func(t *testing.T) {
		oldInstances := fact.created.Load()

		a.forAllInstances(func(_ int, i instance) {
			i.Stop()
		})

		time.Sleep(time.Millisecond * 100)
		require.Equal(t, oldInstances, fact.created.Load())
	})
}

func TestAgent_Stop(t *testing.T) {
	// Lanch two instances
	cfg := Config{
		WALDir: "/tmp/wal",
		Configs: []InstanceConfig{
			{Name: "instance_a"},
			{Name: "instance_b"},
		},
		InstanceRestartBackoff: time.Duration(0),
	}

	var fact mockInstanceFactory

	a, err := newAgent(cfg, log.NewNopLogger(), fact.factory)
	require.NoError(t, err)
	require.Equal(t, fact.created.Load(), int64(2))
	require.Len(t, a.instances, 2)

	oldInstances := fact.created.Load()

	a.Stop()

	time.Sleep(time.Millisecond * 100)
	require.Equal(t, oldInstances, fact.created.Load(), "new instances shuold not have been created")

	a.forAllInstances(func(_ int, inst instance) {
		mi := inst.(*mockInstance)
		require.True(t, mi.exitCalled.Load())
	})
}

type mockInstance struct {
	cfg InstanceConfig

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

func (i *mockInstance) Config() InstanceConfig {
	return i.cfg
}

func (i *mockInstance) Stop() {
	if !i.exitCalled.Load() {
		i.exitCalled.Store(true)
		if i.exitErr == nil {
			i.exitErr = errInstanceStoppedNormally
		}
		close(i.exited)
	}
}

type mockInstanceFactory struct {
	created *atomic.Int64
}

func (f *mockInstanceFactory) factory(_ config.GlobalConfig, cfg InstanceConfig, _ string, _ log.Logger) (instance, error) {

	if f.created == nil {
		f.created = atomic.NewInt64(0)
	}
	f.created.Add(1)

	return &mockInstance{
		cfg:        cfg,
		exited:     make(chan bool),
		waitCalled: atomic.NewBool(false),
		exitCalled: atomic.NewBool(false),
	}, nil
}

func TestMetricValueCollector(t *testing.T) {
	r := prometheus.NewRegistry()
	vc := NewMetricValueCollector(r, "this_should_be_tracked")

	shouldTrack := prometheus.NewGauge(prometheus.GaugeOpts{
		Name: "this_should_be_tracked",
		ConstLabels: prometheus.Labels{
			"foo": "bar",
		},
	})

	shouldTrack.Set(12345)

	shouldNotTrack := prometheus.NewCounter(prometheus.CounterOpts{
		Name: "this_should_not_be_tracked",
	})

	r.MustRegister(shouldTrack, shouldNotTrack)

	vals, err := vc.GetValues("foo", "bar")
	require.NoError(t, err)
	require.Equal(t, []float64{12345}, vals)
}

func TestRemoteWriteMetricInterceptor_AllValues(t *testing.T) {
	r := prometheus.NewRegistry()
	vc := NewMetricValueCollector(r, "track")

	valueA := prometheus.NewGauge(prometheus.GaugeOpts{
		Name: "this_should_be_tracked",
		ConstLabels: prometheus.Labels{
			"foo": "bar",
		},
	})
	valueA.Set(12345)

	valueB := prometheus.NewGauge(prometheus.GaugeOpts{
		Name: "track_this_too",
		ConstLabels: prometheus.Labels{
			"foo": "bar",
		},
	})
	valueB.Set(67890)

	shouldNotReturn := prometheus.NewGauge(prometheus.GaugeOpts{
		Name: "track_this_but_label_does_not_match",
		ConstLabels: prometheus.Labels{
			"foo": "nope",
		},
	})

	r.MustRegister(valueA, valueB, shouldNotReturn)

	vals, err := vc.GetValues("foo", "bar")
	require.NoError(t, err)
	require.Equal(t, []float64{12345, 67890}, vals)
}
