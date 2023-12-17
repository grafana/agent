package build

import (
	"fmt"

	"github.com/alecthomas/units"
	"github.com/grafana/agent/component/common/loki"
	"github.com/grafana/agent/component/faro/receiver"
	"github.com/grafana/agent/component/otelcol"
	"github.com/grafana/agent/converter/diag"
	"github.com/grafana/agent/converter/internal/common"
	app_agent_receiver_v2 "github.com/grafana/agent/pkg/integrations/v2/app_agent_receiver"
	"github.com/grafana/river/rivertypes"
	"github.com/grafana/river/scanner"
)

func (b *IntegrationsConfigBuilder) appendAppAgentReceiverV2(config *app_agent_receiver_v2.Config) {
	args := toAppAgentReceiverV2(config)

	compLabel, err := scanner.SanitizeIdentifier(b.formatJobName(config.Name(), nil))
	if err != nil {
		b.diags.Add(diag.SeverityLevelCritical, fmt.Sprintf("failed to sanitize job name: %s", err))
	}

	b.f.Body().AppendBlock(common.NewBlockWithOverride(
		[]string{"faro", "receiver"},
		compLabel,
		args,
	))
}

func toAppAgentReceiverV2(config *app_agent_receiver_v2.Config) *receiver.Arguments {
	var logLabels map[string]string
	if config.LogsLabels != nil {
		logLabels = config.LogsLabels
	}

	logsReceiver := common.ConvertLogsReceiver{}
	if config.LogsInstance != "" {
		compLabel, err := scanner.SanitizeIdentifier("logs_" + config.LogsInstance)
		if err != nil {
			panic(fmt.Errorf("failed to sanitize job name: %s", err))
		}

		logsReceiver.Expr = fmt.Sprintf("loki.write.%s.receiver", compLabel)
	}

	return &receiver.Arguments{
		LogLabels: logLabels,
		Server: receiver.ServerArguments{
			Host:                  config.Server.Host,
			Port:                  config.Server.Port,
			CORSAllowedOrigins:    config.Server.CORSAllowedOrigins,
			APIKey:                rivertypes.Secret(config.Server.APIKey),
			MaxAllowedPayloadSize: units.Base2Bytes(config.Server.MaxAllowedPayloadSize),
			RateLimiting: receiver.RateLimitingArguments{
				Enabled:   config.Server.RateLimiting.Enabled,
				Rate:      config.Server.RateLimiting.RPS,
				BurstSize: float64(config.Server.RateLimiting.Burstiness),
			},
		},
		SourceMaps: receiver.SourceMapsArguments{
			Download:            config.SourceMaps.Download,
			DownloadFromOrigins: config.SourceMaps.DownloadFromOrigins,
			DownloadTimeout:     config.SourceMaps.DownloadTimeout,
			Locations:           toLocationArguments(config.SourceMaps.FileSystem),
		},
		Output: receiver.OutputArguments{
			Logs:   []loki.LogsReceiver{logsReceiver},
			Traces: []otelcol.Consumer{},
		},
	}
}

func toLocationArguments(locations []app_agent_receiver_v2.SourceMapFileLocation) []receiver.LocationArguments {
	args := make([]receiver.LocationArguments, len(locations))
	for i, location := range locations {
		args[i] = receiver.LocationArguments{
			Path:               location.Path,
			MinifiedPathPrefix: location.MinifiedPathPrefix,
		}
	}
	return args
}
