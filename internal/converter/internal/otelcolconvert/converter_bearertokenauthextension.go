package otelcolconvert

import (
	"fmt"
	"time"

	"github.com/grafana/agent/internal/component/local/file"
	"github.com/grafana/agent/internal/component/otelcol/auth/bearer"
	"github.com/grafana/agent/internal/converter/diag"
	"github.com/grafana/agent/internal/converter/internal/common"
	"github.com/grafana/river/rivertypes"
	"github.com/grafana/river/token/builder"
	"github.com/open-telemetry/opentelemetry-collector-contrib/extension/bearertokenauthextension"
	"go.opentelemetry.io/collector/component"
)

func init() {
	converters = append(converters, bearerTokenAuthExtensionConverter{})
}

type bearerTokenAuthExtensionConverter struct{}

func (bearerTokenAuthExtensionConverter) Factory() component.Factory {
	return bearertokenauthextension.NewFactory()
}

func (bearerTokenAuthExtensionConverter) InputComponentName() string { return "otelcol.auth.bearer" }

func (bearerTokenAuthExtensionConverter) ConvertAndAppend(state *state, id component.InstanceID, cfg component.Config) diag.Diagnostics {
	var diags diag.Diagnostics

	label := state.FlowComponentLabel()

	bcfg := cfg.(*bearertokenauthextension.Config)
	var block *builder.Block

	if bcfg.Filename == "" {
		args := toBearerTokenAuthExtension(bcfg)
		block = common.NewBlockWithOverride([]string{"otelcol", "auth", "bearer"}, label, args)
	} else {
		args, fileContents := toBearerTokenAuthExtensionWithFilename(state, bcfg)
		overrideHook := func(val interface{}) interface{} {
			switch value := val.(type) {
			case rivertypes.Secret:
				return common.CustomTokenizer{Expr: fileContents}
			default:
				return value
			}
		}
		block = common.NewBlockWithOverrideFn([]string{"otelcol", "auth", "bearer"}, label, args, overrideHook)
	}

	diags.Add(
		diag.SeverityLevelInfo,
		fmt.Sprintf("Converted %s into %s", stringifyInstanceID(id), stringifyBlock(block)),
	)

	state.Body().AppendBlock(block)
	return diags
}

func toBearerTokenAuthExtension(cfg *bearertokenauthextension.Config) *bearer.Arguments {
	return &bearer.Arguments{
		Scheme: cfg.Scheme,
		Token:  rivertypes.Secret(string(cfg.BearerToken)),
	}
}
func toBearerTokenAuthExtensionWithFilename(state *state, cfg *bearertokenauthextension.Config) (*bearer.Arguments, string) {
	label := state.FlowComponentLabel()
	args := &file.Arguments{
		Filename:      cfg.Filename,
		Type:          file.DefaultArguments.Type, // Using the default type (fsnotify) since that's what upstream also uses.
		PollFrequency: 60 * time.Second,           // Setting an arbitrary polling time.
		IsSecret:      true,
	}
	block := common.NewBlockWithOverride([]string{"local", "file"}, label, args)
	state.Body().AppendBlock(block)

	return &bearer.Arguments{
		Scheme: cfg.Scheme,
	}, fmt.Sprintf("%s.content", stringifyBlock(block))
}
