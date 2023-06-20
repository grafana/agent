package common

import (
	"fmt"

	"github.com/grafana/agent/component"
	"github.com/grafana/agent/component/discovery"
	"github.com/grafana/agent/pkg/river/rivertypes"
	"github.com/grafana/agent/pkg/river/token/builder"
)

// AppendBlockWithOverride appends Flow component arguments using the convert
// value override hook to the file we are building.
func AppendBlockWithOverride(f *builder.File, name []string, label string, args component.Arguments) {
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
		case []discovery.Target:
			return ConvertTargets{
				Targets: value,
			}
		default:
			return val
		}
	}
}

func GetUniqueLabel(label string, currentCount int) string {
	if currentCount == 1 {
		return label
	}

	return fmt.Sprintf("%s_%d", label, currentCount)
}
