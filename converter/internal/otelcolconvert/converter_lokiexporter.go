package otelcolconvert

import (
	"fmt"
	"reflect"
	"strings"

	"github.com/grafana/agent/component/common/config"
	"github.com/grafana/agent/component/loki/write"
	"github.com/grafana/agent/component/otelcol/exporter/loki"
	"github.com/grafana/agent/converter/diag"
	"github.com/grafana/agent/converter/internal/common"
	"github.com/grafana/river/rivertypes"
	pconfig "github.com/prometheus/common/config"
	"go.opentelemetry.io/collector/component"
	"go.opentelemetry.io/collector/config/configtls"

	"github.com/open-telemetry/opentelemetry-collector-contrib/exporter/lokiexporter"
)

func init() {
	converters = append(converters, lokiExporterConverter{})
}

type lokiExporterConverter struct{}

func (lokiExporterConverter) Factory() component.Factory {
	return lokiexporter.NewFactory()
}

func (lokiExporterConverter) InputComponentName() string { return "otelcol.exporter.loki" }

func (lokiExporterConverter) ConvertAndAppend(state *state, id component.InstanceID, cfg component.Config) diag.Diagnostics {
	var diags diag.Diagnostics

	label := state.FlowComponentLabel()

	lokiWriteComponentID := []componentID{{
		Name:  strings.Split("loki.write", "."),
		Label: label,
	}}

	args1 := toOtelcolExporterLoki(lokiWriteComponentID)
	block1 := common.NewBlockWithOverride([]string{"otelcol", "exporter", "loki"}, label, args1)

	diags.Add(
		diag.SeverityLevelInfo,
		fmt.Sprintf("Converted %s into %s", stringifyInstanceID(id), stringifyBlock(block1)),
	)

	args2, err := toLokiWrite(label, cfg.(*lokiexporter.Config))
	if err != nil {
		diags.Add(
			diag.SeverityLevelError,
			fmt.Sprintf("could not build loki.write block: %s", err),
		)
	}
	block2 := common.NewBlockWithOverride([]string{"loki", "write"}, label, args2)

	diags.Add(
		diag.SeverityLevelInfo,
		fmt.Sprintf("Converted %s into %s", stringifyInstanceID(id), stringifyBlock(block2)),
	)
	diags.Add(
		diag.SeverityLevelInfo,
		"Created a loki.write block with a best-effort conversion of the lokiexporter's confighttp, retry and queue configuration settings. You may want to double check the converted configuration as most fields do not have a 1:1 match",
	)

	state.Body().AppendBlock(block1)
	state.Body().AppendBlock(block2)
	return diags
}

func toOtelcolExporterLoki(ids []componentID) *loki.Arguments {
	return &loki.Arguments{
		ForwardTo: toTokenizedLogsReceivers(ids),
	}
}

func toLokiWrite(name string, cfg *lokiexporter.Config) (*write.Arguments, error) {
	// Defaults for MaxStreams and WAL should be handled on the Flow side.
	res := &write.Arguments{}

	if cfg.Endpoint != "" {
		// TODO(@tpaschalis) Wire in auth from auth extension.
		e := write.GetDefaultEndpointOptions()
		e.Name = name
		e.URL = cfg.Endpoint

		e.RemoteTimeout = cfg.Timeout
		e.TenantID = string(cfg.Headers["X-Scope-OrgID"])
		if !reflect.DeepEqual(cfg.TLSSetting, configtls.TLSClientSetting{}) {
			minv, ok := pconfig.TLSVersions[cfg.TLSSetting.MinVersion]
			if !ok {
				return nil, fmt.Errorf("invalid min tls version provided: %s", cfg.TLSSetting.MinVersion)
			}
			e.HTTPClientConfig.TLSConfig.CA = string(cfg.TLSSetting.CAPem)
			e.HTTPClientConfig.TLSConfig.CAFile = cfg.TLSSetting.CAFile
			e.HTTPClientConfig.TLSConfig.Cert = string(cfg.TLSSetting.CertPem)
			e.HTTPClientConfig.TLSConfig.CertFile = cfg.TLSSetting.CertFile
			e.HTTPClientConfig.TLSConfig.Key = rivertypes.Secret(cfg.TLSSetting.KeyPem)
			e.HTTPClientConfig.TLSConfig.KeyFile = cfg.TLSSetting.KeyFile
			e.HTTPClientConfig.TLSConfig.ServerName = cfg.TLSSetting.ServerName
			e.HTTPClientConfig.TLSConfig.InsecureSkipVerify = cfg.TLSSetting.InsecureSkipVerify
			e.HTTPClientConfig.TLSConfig.MinVersion = config.TLSVersion(minv)
		}

		e.MaxBackoff = cfg.RetrySettings.MaxInterval
		e.MinBackoff = cfg.RetrySettings.InitialInterval

		headers := toHeadersMap(cfg.Headers)
		if len(headers) > 0 {
			e.Headers = headers
		}
		tenant, ok := headers["X-Scope-OrgID"]
		if ok {
			e.TenantID = tenant
		}

		// After trying to translate all the OTel HTTP Client options onto the
		// loki.write component, append it as an endpoint.
		res.Endpoints = append(res.Endpoints, e)
	}

	return res, nil
}
