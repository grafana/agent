package otelcolconvert

import (
	"fmt"

	"github.com/grafana/agent/internal/component/otelcol"
	"github.com/grafana/agent/internal/component/otelcol/receiver/vcenter"
	"github.com/grafana/agent/internal/converter/diag"
	"github.com/grafana/agent/internal/converter/internal/common"
	"github.com/grafana/river/rivertypes"
	"github.com/open-telemetry/opentelemetry-collector-contrib/receiver/vcenterreceiver"
	"go.opentelemetry.io/collector/component"
)

func init() {
	converters = append(converters, vcenterReceiverConverter{})
}

type vcenterReceiverConverter struct{}

func (vcenterReceiverConverter) Factory() component.Factory { return vcenterreceiver.NewFactory() }

func (vcenterReceiverConverter) InputComponentName() string { return "" }

func (vcenterReceiverConverter) ConvertAndAppend(state *state, id component.InstanceID, cfg component.Config) diag.Diagnostics {
	var diags diag.Diagnostics

	label := state.FlowComponentLabel()

	args := toVcenterReceiver(state, id, cfg.(*vcenterreceiver.Config))
	block := common.NewBlockWithOverride([]string{"otelcol", "receiver", "vcenter"}, label, args)

	diags.Add(
		diag.SeverityLevelInfo,
		fmt.Sprintf("Converted %s into %s", stringifyInstanceID(id), stringifyBlock(block)),
	)

	state.Body().AppendBlock(block)
	return diags
}

func toVcenterReceiver(state *state, id component.InstanceID, cfg *vcenterreceiver.Config) *vcenter.Arguments {
	var (
		nextMetrics = state.Next(id, component.DataTypeMetrics)
		nextTraces  = state.Next(id, component.DataTypeTraces)
	)

	return &vcenter.Arguments{
		Endpoint: cfg.Endpoint,
		Username: cfg.Username,
		Password: rivertypes.Secret(cfg.Password),

		DebugMetrics: common.DefaultValue[vcenter.Arguments]().DebugMetrics,

		ScraperControllerArguments: otelcol.ScraperControllerArguments{
			CollectionInterval: cfg.CollectionInterval,
			InitialDelay:       cfg.InitialDelay,
			Timeout:            cfg.Timeout,
		},

		TLS: otelcol.TLSClientArguments{
			TLSSetting: otelcol.TLSSetting{
				CA:             string(cfg.CAPem),
				CAFile:         cfg.CAFile,
				Cert:           string(cfg.CertPem),
				CertFile:       cfg.CertFile,
				Key:            rivertypes.Secret(cfg.KeyPem),
				KeyFile:        cfg.KeyFile,
				MinVersion:     cfg.MinVersion,
				MaxVersion:     cfg.MaxVersion,
				ReloadInterval: cfg.ReloadInterval,
			},
			Insecure:           cfg.Insecure,
			InsecureSkipVerify: cfg.InsecureSkipVerify,
			ServerName:         cfg.ServerName,
		},

		Output: &otelcol.ConsumerArguments{
			Metrics: toTokenizedConsumers(nextMetrics),
			Traces:  toTokenizedConsumers(nextTraces),
		},
	}
}
