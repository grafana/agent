package config

import (
	"flag"
	"fmt"
	"reflect"

	"github.com/go-kit/log"
	"github.com/gorilla/mux"
	"github.com/grafana/agent/pkg/cluster"
	"github.com/grafana/agent/pkg/config/features"
	v1 "github.com/grafana/agent/pkg/metrics"
	"github.com/grafana/agent/pkg/metrics/instance"
	v2 "github.com/grafana/agent/pkg/metrics/v2"
	"github.com/grafana/agent/pkg/util"
	"github.com/prometheus/client_golang/prometheus"
	"google.golang.org/grpc"
	"gopkg.in/yaml.v2"
)

type metricsVersion int

const (
	metricsVersion1 metricsVersion = iota
	metricsVersion2
)

// DefaultVersionedMetrics is the default config for metrics.
var DefaultVersionedMetrics = VersionedMetrics{
	version:  metricsVersion1,
	configV1: v1.DefaultConfig,
}

// VersionedMetrics abstracts the subsystem configs for metrics v1
// and v2. VersionedMetrics can only be unmarshaled as part of Load.
type VersionedMetrics struct {
	version metricsVersion
	raw     util.RawYAML

	configV1  v1.Config
	configV2  v2.Config
	optionsV2 v2.Options
}

func (c *VersionedMetrics) RegisterFlags(fs *flag.FlagSet) {
	// Both of our v1 and v2 subsystems support flags that may overlap. We
	// register both of them to fake flag sets and merge. We'll defer parsing and
	// validation of flags to later once we know which version we intend to load.

	var (
		v1Flags = flag.NewFlagSet("metrics-v1", flag.PanicOnError)
		v2Flags = flag.NewFlagSet("metrics-v2", flag.PanicOnError)
	)
	c.configV1.RegisterFlags(v1Flags)
	c.optionsV2.RegisterFlags(v2Flags)

	deferredFlags := map[string]*deferredFlag{}
	getFlag := func(name string) *deferredFlag {
		if df, ok := deferredFlags[name]; ok {
			return df
		}
		df := &deferredFlag{name: name, vm: c}
		deferredFlags[name] = df
		return df
	}

	v1Flags.VisitAll(func(f *flag.Flag) {
		df := getFlag(f.Name)
		df.v1Flag = f
	})
	v2Flags.VisitAll(func(f *flag.Flag) {
		df := getFlag(f.Name)
		df.v2Flag = f
	})

	// Iterate through our deferred flags and register them to the real fs.
	for _, f := range deferredFlags {
		f.RegisterFlags(fs)
	}
}

type deferredFlag struct {
	v1Flag, v2Flag *flag.Flag

	name string
	raw  string
	set  bool

	vm *VersionedMetrics
}

func (df *deferredFlag) RegisterFlags(fs *flag.FlagSet) {
	var usage string

	switch {
	case df.v1Flag != nil && df.v2Flag != nil:
		usage = df.v1Flag.Usage
		if df.v1Flag.Usage != df.v2Flag.Usage {
			usage = fmt.Sprintf("%s (metrics-next: %s)", df.v1Flag.Usage, df.v2Flag.Usage)
		}
	case df.v1Flag != nil:
		usage = fmt.Sprintf("%s (invalid with metrics-next)", df.v1Flag.Usage)
	case df.v2Flag != nil:
		usage = fmt.Sprintf("%s (metrics-next only)", df.v2Flag.Usage)
	}

	fs.Var(df, df.name, usage)
}

// Set implements flag.Value.
func (s *deferredFlag) Set(v string) error {
	s.raw = v
	s.set = true
	return nil
}

// String implements flag.Value.
func (s *deferredFlag) String() string {
	switch {
	case s.raw != "":
		return s.raw
	case s.v1Flag != nil:
		return s.v1Flag.Value.String()
	case s.v2Flag != nil:
		return s.v2Flag.Value.String()
	default:
		return ""
	}
}

// Validate validates the flag and defers the parsing into the underlying flag
// for v. If the flag was set but isn't valid for v, an error is returned.
func (s *deferredFlag) Validate(v metricsVersion) error {
	switch v {
	case metricsVersion1:
		if s.v1Flag == nil {
			if s.set {
				return fmt.Errorf("flag %q cannot be used with metrics-next feature enabled", s.name)
			}
			return nil
		}
		return s.v1Flag.Value.Set(s.raw)
	case metricsVersion2:
		if s.v2Flag == nil {
			if s.set {
				return fmt.Errorf("flag %q cannot be used unless metrics-next feature is enabled", s.name)
			}
			return nil
		}
		return s.v2Flag.Value.Set(s.raw)
	default:
		panic("unexpected version")
	}
}

// Validate ensures that no flag was improperly set. Validate will internally
// set the version based on the feature flag.
func (c *VersionedMetrics) Validate(fs *flag.FlagSet) error {
	switch features.Enabled(fs, featMetricsNext) {
	case true:
		if err := c.setVersion(metricsVersion2); err != nil {
			return err
		}
	default:
		if err := c.setVersion(metricsVersion1); err != nil {
			return err
		}
	}

	var firstErr error
	fs.Visit(func(f *flag.Flag) {
		if firstErr != nil {
			return
		}
		df, ok := f.Value.(*deferredFlag)
		if !ok || df.vm != c {
			return
		}
		firstErr = df.Validate(c.version)
	})

	return firstErr
}

var (
	_ yaml.Unmarshaler = (*VersionedMetrics)(nil)
	_ yaml.Marshaler   = (*VersionedMetrics)(nil)
)

// UnmarshalYAML implements yaml.Unmarshaler. Full unmarshaling is deferred until
// setVersion is invoked.
func (c *VersionedMetrics) UnmarshalYAML(unmarshal func(interface{}) error) error {
	return unmarshal(&c.raw)
}

// MarshalYAML implements yaml.Marshaler.
func (c VersionedMetrics) MarshalYAML() (interface{}, error) {
	switch c.version {
	case metricsVersion1:
		return c.configV1, nil
	case metricsVersion2:
		return c.configV2, nil
	default:
		return c.raw, nil
	}
}

// IsZero implements yaml.IsZeroer.
func (c VersionedMetrics) IsZero() bool {
	switch c.version {
	case metricsVersion1:
		return reflect.ValueOf(c.configV1).IsZero()
	case metricsVersion2:
		return reflect.ValueOf(c.configV2).IsZero()
	default:
		return len(c.raw) == 0
	}
}

// ApplyDefaults applies defaults to the subsystem.
func (c *VersionedMetrics) ApplyDefaults() error {
	if c.version != metricsVersion2 {
		return c.configV1.ApplyDefaults()
	}
	return c.configV2.ApplyDefaults(c.optionsV2)
}

// setVersion completes the deferred unmarshal and unmarshals the raw YAML into
// the subsystem config for version v.
func (c *VersionedMetrics) setVersion(v metricsVersion) error {
	c.version = v

	switch c.version {
	case metricsVersion1:
		c.configV1 = v1.DefaultConfig
		return yaml.UnmarshalStrict(c.raw, &c.configV1)
	case metricsVersion2:
		c.configV2 = v2.DefaultConfig
		return yaml.UnmarshalStrict(c.raw, &c.configV2)
	default:
		panic(fmt.Sprintf("unknown metrics version %d", c.version))
	}
}

// Metrics is an abstraction over both the v1 and v2 systems.
type Metrics interface {
	InstanceManager() instance.Manager
	Validate(*instance.Config) error
	ApplyConfig(*VersionedMetrics) error
	WireAPI(*mux.Router)
	WireGRPC(*grpc.Server)
	Stop()
}

// NewMetrics creates a new Metrics instance. cfg must have already internally
// had a specific version set before calling.
func NewMetrics(l log.Logger, reg prometheus.Registerer, cfg *VersionedMetrics, cluster *cluster.Node) (Metrics, error) {
	if cfg.version != metricsVersion2 {
		inst, err := v1.New(reg, cfg.configV1, l)
		if err != nil {
			return nil, err
		}
		return &v1Metrics{Agent: inst}, nil
	}

	cfg.optionsV2.Cluster = cluster
	inst, err := v2.New(l, reg, cfg.optionsV2)
	if err != nil {
		return nil, err
	}
	if err := inst.ApplyConfig(cfg.configV2); err != nil {
		return nil, err
	}
	return &v2Metrics{Metrics: inst}, nil
}

type v1Metrics struct{ *v1.Agent }

func (m *v1Metrics) ApplyConfig(c *VersionedMetrics) error {
	return m.Agent.ApplyConfig(c.configV1)
}

type v2Metrics struct{ *v2.Metrics }

func (m *v2Metrics) ApplyConfig(c *VersionedMetrics) error {
	return m.Metrics.ApplyConfig(c.configV2)
}

func (m *v2Metrics) InstanceManager() instance.Manager {
	return fakeInstanceManager{}
}

func (m *v2Metrics) Validate(*instance.Config) error {
	return errDynInstanceUnsupported
}

var errDynInstanceUnsupported = fmt.Errorf("dynamic instances not supported with metrics-next")

type fakeInstanceManager struct{}

func (im fakeInstanceManager) GetInstance(name string) (instance.ManagedInstance, error) {
	// TODO(rfratto): gets should be OK, but the interface here would need to change.
	return nil, errDynInstanceUnsupported
}

func (im fakeInstanceManager) ListInstances() map[string]instance.ManagedInstance {
	return nil
}

func (im fakeInstanceManager) ListConfigs() map[string]instance.Config {
	return nil
}

func (im fakeInstanceManager) ApplyConfig(instance.Config) error {
	return errDynInstanceUnsupported
}

func (im fakeInstanceManager) DeleteConfig(name string) error {
	return errDynInstanceUnsupported
}

func (im fakeInstanceManager) Stop() {}
