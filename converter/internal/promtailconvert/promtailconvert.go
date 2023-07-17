package promtailconvert

import (
	"bytes"
	"flag"
	"fmt"

	"github.com/alecthomas/units"
	"github.com/grafana/agent/component/common/loki"
	lokiwrite "github.com/grafana/agent/component/loki/write"
	"github.com/grafana/agent/converter/diag"
	"github.com/grafana/agent/converter/internal/common"
	"github.com/grafana/agent/converter/internal/prometheusconvert"
	"github.com/grafana/agent/converter/internal/promtailconvert/internal/build"
	"github.com/grafana/agent/pkg/river/token/builder"
	"github.com/grafana/dskit/flagext"
	"github.com/grafana/loki/clients/pkg/promtail/client"
	promtailcfg "github.com/grafana/loki/clients/pkg/promtail/config"
	"github.com/grafana/loki/clients/pkg/promtail/limit"
	"github.com/grafana/loki/clients/pkg/promtail/positions"
	"github.com/grafana/loki/clients/pkg/promtail/scrapeconfig"
	lokicfgutil "github.com/grafana/loki/pkg/util/cfg"
	lokiflag "github.com/grafana/loki/pkg/util/flagext"
	"gopkg.in/yaml.v2"
)

type Config struct {
	promtailcfg.Config `yaml:",inline"`
}

// Clone takes advantage of pass-by-value semantics to return a distinct *Config.
// This is primarily used to parse a different flag set without mutating the original *Config.
func (c *Config) Clone() flagext.Registerer {
	return func(c Config) *Config {
		return &c
	}(*c)
}

// Convert implements a Promtail config converter.
func Convert(in []byte) ([]byte, diag.Diagnostics) {
	var (
		diags diag.Diagnostics
		cfg   Config
	)

	// Set default values first.
	flagSet := flag.NewFlagSet("", flag.PanicOnError)
	err := lokicfgutil.Unmarshal(&cfg,
		lokicfgutil.Defaults(flagSet),
	)
	if err != nil {
		diags.Add(diag.SeverityLevelCritical, fmt.Sprintf("failed to set default Promtail config values: %s", err))
		return nil, diags
	}

	// Unmarshall explicitly specified values
	if err := yaml.UnmarshalStrict(in, &cfg); err != nil {
		diags.Add(diag.SeverityLevelCritical, fmt.Sprintf("failed to parse Promtail config: %s", err))
		return nil, diags
	}

	// Replicate promtails' handling of this deprecated field.
	if cfg.ClientConfig.URL.URL != nil {
		// if a single client config is used we add it to the multiple client config for backward compatibility
		cfg.ClientConfigs = append(cfg.ClientConfigs, cfg.ClientConfig)
	}

	f := builder.NewFile()
	diags = AppendAll(f, &cfg.Config, diags)

	var buf bytes.Buffer
	if _, err := f.WriteTo(&buf); err != nil {
		diags.Add(diag.SeverityLevelCritical, fmt.Sprintf("failed to render Flow config: %s", err.Error()))
		return nil, diags
	}

	if len(buf.Bytes()) == 0 {
		return nil, diags
	}

	prettyByte, newDiags := common.PrettyPrint(buf.Bytes())
	diags = append(diags, newDiags...)
	return prettyByte, diags
}

// AppendAll analyzes the entire promtail config in memory and transforms it
// into Flow components. It then appends each argument to the file builder.
func AppendAll(f *builder.File, cfg *promtailcfg.Config, diags diag.Diagnostics) diag.Diagnostics {
	validateTopLevelConfig(cfg, &diags)

	var writeReceivers = make([]loki.LogsReceiver, len(cfg.ClientConfigs))
	var writeBlocks = make([]*builder.Block, len(cfg.ClientConfigs))
	// Each client config needs to be a separate remote_write,
	// because they may have different ExternalLabels fields.
	for i, cc := range cfg.ClientConfigs {
		writeBlocks[i], writeReceivers[i] = newLokiWrite(&cc, &diags, i)
	}

	gc := &build.GlobalContext{
		WriteReceivers:   writeReceivers,
		TargetSyncPeriod: cfg.TargetConfig.SyncPeriod,
	}

	for _, sc := range cfg.ScrapeConfig {
		appendScrapeConfig(f, &sc, &diags, gc)
	}

	for _, write := range writeBlocks {
		f.Body().AppendBlock(write)
	}

	return diags
}

func defaultPositionsConfig() positions.Config {
	// We obtain the default by registering the flags
	cfg := positions.Config{}
	cfg.RegisterFlags(flag.NewFlagSet("", flag.PanicOnError))
	return cfg
}

func defaultLimitsConfig() limit.Config {
	cfg := limit.Config{}
	cfg.RegisterFlagsWithPrefix("", flag.NewFlagSet("", flag.PanicOnError))
	return cfg
}

func appendScrapeConfig(
	f *builder.File,
	cfg *scrapeconfig.Config,
	diags *diag.Diagnostics,
	gctx *build.GlobalContext,
) {

	b := build.NewScrapeConfigBuilder(f, diags, cfg, gctx)

	// Append all the SD components
	b.AppendKubernetesSDs()
	//TODO(thampiotr): add support for other SDs

	// Append loki.source.file to process all SD components' targets.
	// If any relabelling is required, it will be done via a discovery.relabel component.
	// The files will be watched and the globs in file paths will be expanded using discovery.file component.
	// The log entries are sent to loki.process if processing is needed, or directly to loki.write components.
	b.AppendLokiSourceFile()

	// Append all the components that produce logs directly.
	// If any relabelling is required, it will be done via a loki.relabel component.
	// The logs are sent to loki.process if processing is needed, or directly to loki.write components.
	//TODO(thampiotr): add support for other integrations
	b.AppendCloudFlareConfig()
	b.AppendJournalConfig()
}

func newLokiWrite(client *client.Config, diags *diag.Diagnostics, index int) (*builder.Block, loki.LogsReceiver) {
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
