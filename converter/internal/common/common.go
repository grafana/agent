package common

import (
	"bytes"
	"fmt"
	"strings"

	"github.com/grafana/river/parser"
	"github.com/grafana/river/printer"
	"github.com/grafana/river/scanner"

	"github.com/grafana/agent/component"
	flow_relabel "github.com/grafana/agent/component/common/relabel"
	"github.com/grafana/agent/component/discovery"
	"github.com/grafana/agent/converter/diag"
	"github.com/grafana/river/rivertypes"
	"github.com/grafana/river/token/builder"
)

// NewBlockWithOverride generates a new [*builder.Block] using a hook to
// override specific types.
func NewBlockWithOverride(name []string, label string, args component.Arguments) *builder.Block {
	return NewBlockWithOverrideFn(name, label, args, getValueOverrideHook())
}

// NewBlockWithOverrideFn generates a new [*builder.Block] using a hook fn to
// override specific types.
func NewBlockWithOverrideFn(name []string, label string, args component.Arguments, fn builder.ValueOverrideHook) *builder.Block {
	block := builder.NewBlock(name, label)
	block.Body().SetValueOverrideHook(fn)
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
		case []rivertypes.Secret:
			secrets := make([]string, 0, len(value))
			for _, secret := range value {
				secrets = append(secrets, string(secret))
			}
			return secrets
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

// LabelForParts generates a consistent component label for a set of parts
// delimited with an underscore.
func LabelForParts(parts ...interface{}) string {
	var sParts []string
	for _, part := range parts {
		if part != "" {
			sParts = append(sParts, fmt.Sprintf("%v", part))
		}
	}
	return strings.Join(sParts, "_")
}

// LabelWithIndex generates a consistent component label for a set of parts
// delimited with an underscore and suffixed with the provided index. If the
// index is 0, the label is generated without the index.
func LabelWithIndex(index int, parts ...interface{}) string {
	if index == 0 {
		return LabelForParts(parts...)
	}

	appendedIndex := index + 1
	return LabelForParts(append(parts, appendedIndex)...)
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

func SanitizeIdentifierPanics(in string) string {
	out, err := scanner.SanitizeIdentifier(in)
	if err != nil {
		panic(err)
	}
	return out
}
