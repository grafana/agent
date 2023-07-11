package common

import (
	"bytes"
	"fmt"

	"github.com/grafana/agent/pkg/river/parser"
	"github.com/grafana/agent/pkg/river/printer"

	"github.com/grafana/agent/component"
	flow_relabel "github.com/grafana/agent/component/common/relabel"
	"github.com/grafana/agent/component/discovery"
	"github.com/grafana/agent/converter/diag"
	"github.com/grafana/agent/pkg/river/rivertypes"
	"github.com/grafana/agent/pkg/river/token/builder"
)

// NewBlockWithOverride generates a new [*builder.Block] using a hook to
// override specific types.
func NewBlockWithOverride(name []string, label string, args component.Arguments) *builder.Block {
	block := builder.NewBlock(name, label)
	block.Body().SetValueOverrideHook(getValueOverrideHook())
	block.Body().AppendFrom(args)
	return block
}

// GetValueOverrideHook returns a hook for overriding the go value of
// specific go types for converting configs from one type to another.
func getValueOverrideHook() builder.ValueOverrideHook {
	return func(val interface{}) interface{} {
		switch value := val.(type) {
		case rivertypes.Secret:
			return string(value)
		case flow_relabel.Regexp:
			return value.String()
		case []discovery.Target:
			return ConvertTargets{
				Targets: value,
			}
		default:
			return val
		}
	}
}

// GetUniqueLabel appends a counter to the input label if it is not the first.
// The input label is 1-indexed.
func GetUniqueLabel(label string, currentCount int) string {
	if currentCount == 1 {
		return label
	}

	return fmt.Sprintf("%s_%d", label, currentCount)
}

// PrettyPrint parses river config and returns it in a standardize format.
// If PrettyPrint fails, the input is returned unmodified.
func PrettyPrint(in []byte) ([]byte, diag.Diagnostics) {
	var diags diag.Diagnostics

	// Return early if there was no file.
	if len(in) == 0 {
		return in, diags
	}

	f, err := parser.ParseFile("", in)
	if err != nil {
		diags.Add(diag.SeverityLevelError, err.Error())
		return in, diags
	}

	var buf bytes.Buffer
	if err = printer.Fprint(&buf, f); err != nil {
		diags.Add(diag.SeverityLevelError, err.Error())
		return in, diags
	}

	// Add a trailing newline at the end of the file, which is omitted by Fprint.
	_, _ = buf.WriteString("\n")
	return buf.Bytes(), nil
}
