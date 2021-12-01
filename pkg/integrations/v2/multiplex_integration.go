package integrations

import (
	"context"
	"fmt"
	"reflect"
)

func NewMultiplexIntegration(config MultiplexConfig, opts IntegrationOptions) (Integration, error) {
	var mux multiplexIntegration
	cc, err := mux.getControllerConfig(config)
	if err != nil {
		return nil, err
	}

	ctrl, err := NewController(cc, opts)
	if err != nil {
		return nil, err
	}

	mux.Controller = ctrl
	return &mux, nil
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
	mc, ok := c.(MultiplexConfig)
	if !ok {
		return fmt.Errorf("invalid type %T", c)
	}
	cc, err := mi.getControllerConfig(mc)
	if err != nil {
		return err
	}
	return mi.UpdateController(cc, mi.opts)
}

func (mu *multiplexIntegration) getControllerConfig(c MultiplexConfig) (ControllerConfig, error) {
	val := reflect.ValueOf(c)

	if val.Kind() == reflect.Ptr {
		val = val.Elem()
	}
	if val.Kind() != reflect.Slice {
		return nil, fmt.Errorf("invalid config type %T: must be slice of Config", c)
	}

	cc := make(ControllerConfig, 0, val.Len())
	for i := 0; i < val.Len(); i++ {
		v := val.Index(i).Interface()
		item, ok := v.(Config)
		if !ok {
			return nil, fmt.Errorf("invlaid config type %T: slice element %T does not implement Config", c, v)
		}
		cc = append(cc, item)
	}

	return cc, nil
}
