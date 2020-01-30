package prometheus

import (
	"encoding/json"
	"errors"
	"testing"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/stretchr/testify/require"
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
	bb, err := json.Marshal(c)
	require.NoError(t, err)

	var cp Config
	err = json.Unmarshal(bb, &cp)
	require.NoError(t, err)
	return cp
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

func TestParseDescName(t *testing.T) {
	d := prometheus.NewDesc(
		"this_is_my_metric_name",
		"the metric for testing",
		[]string{"app", "cluster"},
		prometheus.Labels{"foo": "bar"},
	)

	name := parseDescName(d)
	require.Equal(t, "this_is_my_metric_name", name)
}
