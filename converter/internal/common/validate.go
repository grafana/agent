package common

import (
	"fmt"
	"reflect"

	"github.com/grafana/agent/converter/diag"
)

func UnsupportedNotDeepEquals(a any, b any, name string) diag.Diagnostics {
	var diags diag.Diagnostics
	if !reflect.DeepEqual(a, b) {
		diags.Add(diag.SeverityLevelError, fmt.Sprintf("unsupported %s config was provided.", name))
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
