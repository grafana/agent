package build

import (
	"fmt"

	"github.com/alecthomas/units"
	"github.com/grafana/agent/component/common/loki"
	lokiwrite "github.com/grafana/agent/component/loki/write"
	"github.com/grafana/agent/converter/diag"
	"github.com/grafana/agent/converter/internal/common"
	"github.com/grafana/agent/converter/internal/prometheusconvert"
	"github.com/grafana/agent/pkg/river/token/builder"
	"github.com/grafana/loki/clients/pkg/promtail/client"
	lokiflag "github.com/grafana/loki/pkg/util/flagext"
)

func NewLokiWrite(client *client.Config, diags *diag.Diagnostics, index int) (*builder.Block, loki.LogsReceiver) {
	label := fmt.Sprintf("default_%d", index)
	lokiWriteArgs := toLokiWriteArguments(client, diags)
	block := common.NewBlockWithOverride([]string{"loki", "write"}, label, lokiWriteArgs)
	return block, common.ConvertLogsReceiver{
		Expr: fmt.Sprintf("loki.write.%s.receiver", label),
	}
}

func toLokiWriteArguments(config *client.Config, diags *diag.Diagnostics) *lokiwrite.Arguments {
	batchSize, err := units.ParseBase2Bytes(fmt.Sprintf("%dB", config.BatchSize))
	if err != nil {
		diags.Add(
			diag.SeverityLevelError,
			fmt.Sprintf("failed to parse BatchSize for client config %s: %s", config.Name, err.Error()),
		)
	}

	// This is not supported yet - see https://github.com/grafana/agent/issues/4335.
	if config.DropRateLimitedBatches {
		diags.Add(
			diag.SeverityLevelError,
			"DropRateLimitedBatches is currently not supported in Grafana Agent Flow.",
		)
	}

	// Also deprecated in promtail.
	if len(config.StreamLagLabels) != 0 {
		diags.Add(
			diag.SeverityLevelWarn,
			"stream_lag_labels is deprecated and the associated metric has been removed",
		)
	}

	return &lokiwrite.Arguments{
		Endpoints: []lokiwrite.EndpointOptions{
			{
				Name:              config.Name,
				URL:               config.URL.String(),
				BatchWait:         config.BatchWait,
				BatchSize:         batchSize,
				HTTPClientConfig:  prometheusconvert.ToHttpClientConfig(&config.Client),
				Headers:           config.Headers,
				MinBackoff:        config.BackoffConfig.MinBackoff,
				MaxBackoff:        config.BackoffConfig.MaxBackoff,
				MaxBackoffRetries: config.BackoffConfig.MaxRetries,
				RemoteTimeout:     config.Timeout,
				TenantID:          config.TenantID,
			},
		},
		ExternalLabels: convertFlagLabels(config.ExternalLabels),
	}
}

func convertFlagLabels(labels lokiflag.LabelSet) map[string]string {
	result := map[string]string{}
	for k, v := range labels.LabelSet {
		result[string(k)] = string(v)
	}
	return result
}
