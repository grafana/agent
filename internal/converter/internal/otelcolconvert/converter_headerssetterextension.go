package otelcolconvert

import (
	"fmt"

	"github.com/grafana/agent/internal/component/otelcol/auth/headers"
	"github.com/grafana/agent/internal/converter/diag"
	"github.com/grafana/agent/internal/converter/internal/common"
	"github.com/grafana/river/rivertypes"
	"github.com/open-telemetry/opentelemetry-collector-contrib/extension/headerssetterextension"
	"go.opentelemetry.io/collector/component"
)

func init() {
	converters = append(converters, headersSetterExtensionConverter{})
}

type headersSetterExtensionConverter struct{}

func (headersSetterExtensionConverter) Factory() component.Factory {
	return headerssetterextension.NewFactory()
}

func (headersSetterExtensionConverter) InputComponentName() string { return "otelcol.auth.headers" }

func (headersSetterExtensionConverter) ConvertAndAppend(state *State, id component.InstanceID, cfg component.Config) diag.Diagnostics {
	var diags diag.Diagnostics

	label := state.FlowComponentLabel()

	args := toHeadersSetterExtension(cfg.(*headerssetterextension.Config))
	block := common.NewBlockWithOverride([]string{"otelcol", "auth", "headers"}, label, args)

	diags.Add(
		diag.SeverityLevelInfo,
		fmt.Sprintf("Converted %s into %s", StringifyInstanceID(id), StringifyBlock(block)),
	)

	state.Body().AppendBlock(block)
	return diags
}

func toHeadersSetterExtension(cfg *headerssetterextension.Config) *headers.Arguments {
	res := make([]headers.Header, 0, len(cfg.HeadersConfig))
	for _, h := range cfg.HeadersConfig {
		var val *rivertypes.OptionalSecret
		if h.Value != nil {
			val = &rivertypes.OptionalSecret{
				IsSecret: false, // we default to non-secret so that the converted configuration includes the actual value instead of (secret).
				Value:    *h.Value,
			}
		}

		res = append(res, headers.Header{
			Key:         *h.Key, // h.Key cannot be nil or it's not valid configuration for the upstream component.
			Value:       val,
			FromContext: h.FromContext,
			Action:      headers.Action(h.Action),
		})
	}

	return &headers.Arguments{
		Headers: res,
	}
}
