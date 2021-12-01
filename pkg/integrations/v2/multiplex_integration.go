package integrations

import (
	"context"
	"fmt"
	"reflect"
)

// NewMultiplexConfig will return a Config that generates a multiplexing
// integration. The integration will handle multiple instances of inner.
//
// name must be different from inner.Name or NewMultiplexConfig will panic.
func NewMultiplexConfig(name string, inner Config) Config {
	if name == inner.Name() {
		panic("bug: multiplex integration name must be different than integration name. e.g., add _configs as a suffix")
	}

	return &multiplexConfig{name: name, reference: inner}
}

// TODO(rfratto): verify that registered integrations don't reset the value of
// the registered Config when unmarshaling, otherwise this will break

type multiplexConfig struct {
	name      string
	reference Config

	// Unmarshaled configs of type reference.
	configs []Config
}

// UnmarshalYAML implements yaml.Unmarshaler and will unmarshal a slice of
// mt.reference, filling mt.configs.
func (mt *multiplexConfig) UnmarshalYAML(unmarshal func(interface{}) error) error {
	sliceTy := reflect.SliceOf(reflect.TypeOf(mt.reference))
	slicePtr := reflect.New(sliceTy)

	if err := unmarshal(slicePtr.Interface()); err != nil {
		return err
	}

	mt.configs = mt.configs[:0]
	slice := slicePtr.Elem()
	for i := 0; i < slice.Len(); i++ {
		val := slice.Index(i).Interface().(Config)
		mt.configs = append(mt.configs, val)
	}
	return nil
}

// UnmarshalYAML implements yaml.Marshaler and will marshal the slice of mt.configs.
func (mt *multiplexConfig) MarshalYAML() (interface{}, error) {
	return mt.configs, nil
}

func (mt *multiplexConfig) Name() string {
	return mt.name
}

func (mt *multiplexConfig) Identifier(IntegrationOptions) (string, error) {
	return mt.name, nil
}

func (mt *multiplexConfig) NewIntegration(opts IntegrationOptions) (Integration, error) {
	ctrl, err := NewController(ControllerConfig(mt.configs), opts)
	if err != nil {
		return nil, err
	}

	return &multiplexIntegration{Controller: ctrl}, nil
}

// multiplexIntegration implements Integration and all known extensions using a
// controller.
type multiplexIntegration struct {
	*Controller
}

// Interface assertions for multiplexIntegration.
var (
	_ Integration        = (*multiplexIntegration)(nil)
	_ UpdateIntegration  = (*multiplexIntegration)(nil)
	_ HTTPIntegration    = (*multiplexIntegration)(nil)
	_ MetricsIntegration = (*multiplexIntegration)(nil)
)

func (mi *multiplexIntegration) RunIntegration(ctx context.Context) error {
	return mi.Run(ctx)
}

func (mi *multiplexIntegration) ApplyConfig(c Config) error {
	mc, ok := c.(*multiplexConfig)
	if !ok {
		return fmt.Errorf("unvalid config type %T", c)
	}
	return mi.UpdateController(mc.configs, mi.opts)
}
