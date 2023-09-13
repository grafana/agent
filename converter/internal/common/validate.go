package common

import (
	"fmt"
	"reflect"
	"strings"

	"github.com/grafana/agent/converter/diag"
	"github.com/grafana/river/token/builder"
)

func UnsupportedNotDeepEquals(a any, b any, name string) diag.Diagnostics {
	return UnsupportedNotDeepEqualsMessage(a, b, name, "")
}

func UnsupportedNotDeepEqualsMessage(a any, b any, name string, message string) diag.Diagnostics {
	var diags diag.Diagnostics
	if !reflect.DeepEqual(a, b) {
		if message != "" {
			diags.Add(diag.SeverityLevelError, fmt.Sprintf("unsupported %s config was provided: %s", name, message))
		} else {
			diags.Add(diag.SeverityLevelError, fmt.Sprintf("unsupported %s config was provided.", name))
		}
	}

	return diags
}

func UnsupportedNotEquals(a any, b any, name string) diag.Diagnostics {
	var diags diag.Diagnostics
	if a != b {
		diags.Add(diag.SeverityLevelError, fmt.Sprintf("unsupported %s config was provided.", name))
	}

	return diags
}

// ValidateNodes will look at the final nodes and ensure that there are no
// duplicate labels.
func ValidateNodes(f *builder.File) diag.Diagnostics {
	var diags diag.Diagnostics

	nodes := f.Body().Nodes()
	labels := make(map[string]string, len(nodes))
	for _, node := range nodes {
		switch n := node.(type) {
		case *builder.Block:
			label := strings.Join(n.Name, ".")
			if n.Label != "" {
				label += "." + n.Label
			}
			if _, ok := labels[label]; ok {
				diags.Add(diag.SeverityLevelCritical, fmt.Sprintf("duplicate label after conversion %q. this is due to how valid flow labels are assembled and can be avoided by updating named properties in the source config.", label))
			} else {
				labels[label] = label
			}
		}
	}

	return diags
}
