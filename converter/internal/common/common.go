package common

import (
	"github.com/grafana/agent/component"
	"github.com/grafana/agent/pkg/river/rivertypes"
	"github.com/grafana/agent/pkg/river/token/builder"
)

// AppendBlock appends Flow component arguments using the convert value
// override hook to the file we are building.
func AppendBlock(f *builder.File, name []string, label string, args component.Arguments) {
	block := builder.NewBlock(name, label)
	block.Body().SetValueOverrideHook(getValueOverrideHook())
	block.Body().AppendFrom(args)
	f.Body().AppendBlock(block)
}

// GetValueOverrideHook returns a hook for overriding the go value of
// specific go types for converting configs from one type to another.
func getValueOverrideHook() builder.ValueOverrideHook {
	return func(val interface{}) interface{} {
		switch value := val.(type) {
		case rivertypes.Secret:
			return string(value)
		default:
			return val
		}
	}
}
