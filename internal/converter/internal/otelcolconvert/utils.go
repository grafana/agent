package otelcolconvert

import (
	"bytes"
	"fmt"
	"strings"

	"github.com/grafana/agent/internal/converter/diag"
	"github.com/grafana/agent/internal/converter/internal/common"
	"github.com/grafana/river/token/builder"
	"go.opentelemetry.io/collector/component"
)

func stringifyInstanceID(id component.InstanceID) string {
	return fmt.Sprintf("%s/%s", stringifyKind(id.Kind), id.ID)
}

func stringifyKind(k component.Kind) string {
	switch k {
	case component.KindReceiver:
		return "receiver"
	case component.KindProcessor:
		return "processor"
	case component.KindExporter:
		return "exporter"
	case component.KindExtension:
		return "extension"
	case component.KindConnector:
		return "connector"
	default:
		return fmt.Sprintf("Kind(%d)", k)
	}
}

func stringifyBlock(block *builder.Block) string {
	return fmt.Sprintf("%s.%s", strings.Join(block.Name, "."), block.Label)
}

// ConvertWithoutValidation is similar to `otelcolconvert.go`'s Convert but without validating generated configs
// This is to help testing `sigv4authextension` converter as its Validate() method calls up external cloud
// service and we can't inject mock SigV4 credential provider since the attribute is set as internal in the
// upstream.
// Remove this once credentials provider is open for mocking.
func ConvertWithoutValidation(in []byte, extraArgs []string) ([]byte, diag.Diagnostics) {
	var diags diag.Diagnostics

	if len(extraArgs) > 0 {
		diags.Add(diag.SeverityLevelCritical, fmt.Sprintf("extra arguments are not supported for the otelcol converter: %s", extraArgs))
		return nil, diags
	}

	cfg, err := readOpentelemetryConfig(in)
	if err != nil {
		diags.Add(diag.SeverityLevelCritical, err.Error())
		return nil, diags
	}

	f := builder.NewFile()

	diags.AddAll(AppendConfig(f, cfg, ""))
	diags.AddAll(common.ValidateNodes(f))

	var buf bytes.Buffer
	if _, err := f.WriteTo(&buf); err != nil {
		diags.Add(diag.SeverityLevelCritical, fmt.Sprintf("failed to render Flow config: %s", err.Error()))
		return nil, diags
	}

	if len(buf.Bytes()) == 0 {
		return nil, diags
	}

	prettyByte, newDiags := common.PrettyPrint(buf.Bytes())
	diags.AddAll(newDiags)
	return prettyByte, diags
}
