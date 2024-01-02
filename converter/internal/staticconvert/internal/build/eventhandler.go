package build

import (
	"fmt"

	"github.com/grafana/agent/component/common/loki"
	"github.com/grafana/agent/component/loki/source/kubernetes_events"
	"github.com/grafana/agent/converter/diag"
	"github.com/grafana/agent/converter/internal/common"
	eventhandler_v2 "github.com/grafana/agent/pkg/integrations/v2/eventhandler"
	"github.com/grafana/river/scanner"
)

func (b *IntegrationsConfigBuilder) appendEventHandlerV2(config *eventhandler_v2.Config) {
	args := toEventHandlerV2(config)

	compLabel, err := scanner.SanitizeIdentifier(b.formatJobName(config.Name(), nil))
	if err != nil {
		b.diags.Add(diag.SeverityLevelCritical, fmt.Sprintf("failed to sanitize job name: %s", err))
	}

	b.diags.AddAll(common.ValidateSupported(common.NotDeepEquals, config.SendTimeout, eventhandler_v2.DefaultConfig.SendTimeout, "eventhandler send_timeout", "this field is not configurable in flow mode"))
	b.diags.AddAll(common.ValidateSupported(common.NotDeepEquals, config.CachePath, eventhandler_v2.DefaultConfig.CachePath, "eventhandler cache_path", "this field is not configurable in flow mode"))
	b.diags.AddAll(common.ValidateSupported(common.NotDeepEquals, config.InformerResync, eventhandler_v2.DefaultConfig.InformerResync, "eventhandler informer_resync", "this field is not configurable in flow mode"))
	b.diags.AddAll(common.ValidateSupported(common.NotDeepEquals, config.FlushInterval, eventhandler_v2.DefaultConfig.FlushInterval, "eventhandler flush_interval", "this field is not configurable in flow mode"))
	b.diags.AddAll(common.ValidateSupported(common.NotDeepEquals, len(config.ExtraLabels), 0, "eventhandler extra_labels", "extra_labels for logs can be configured in flow mode using a loki.process component"))

	b.f.Body().AppendBlock(common.NewBlockWithOverride(
		[]string{"loki", "source", "kubernetes_events"},
		compLabel,
		args,
	))
}

func toEventHandlerV2(config *eventhandler_v2.Config) *kubernetes_events.Arguments {
	logsReceiver := common.ConvertLogsReceiver{}
	if config.LogsInstance != "" {
		compLabel, err := scanner.SanitizeIdentifier("logs_" + config.LogsInstance)
		if err != nil {
			panic(fmt.Errorf("failed to sanitize job name: %s", err))
		}

		logsReceiver.Expr = fmt.Sprintf("loki.write.%s.receiver", compLabel)
	}

	defaultOverrides := kubernetes_events.DefaultArguments
	defaultOverrides.Client.KubeConfig = config.KubeconfigPath
	if config.Namespace != "" {
		defaultOverrides.Namespaces = []string{config.Namespace}
	}

	return &kubernetes_events.Arguments{
		ForwardTo:  []loki.LogsReceiver{logsReceiver},
		JobName:    kubernetes_events.DefaultArguments.JobName,
		Namespaces: defaultOverrides.Namespaces,
		LogFormat:  config.LogFormat,
		Client:     defaultOverrides.Client,
	}
}
