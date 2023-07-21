package promtailconvert

import (
	"bytes"
	"flag"
	"fmt"

	"github.com/grafana/agent/component/common/loki"
	"github.com/grafana/agent/converter/diag"
	"github.com/grafana/agent/converter/internal/common"
	"github.com/grafana/agent/converter/internal/promtailconvert/internal/build"
	"github.com/grafana/agent/pkg/river/token/builder"
	"github.com/grafana/dskit/flagext"
	promtailcfg "github.com/grafana/loki/clients/pkg/promtail/config"
	"github.com/grafana/loki/clients/pkg/promtail/limit"
	"github.com/grafana/loki/clients/pkg/promtail/positions"
	"github.com/grafana/loki/clients/pkg/promtail/scrapeconfig"
	lokicfgutil "github.com/grafana/loki/pkg/util/cfg"
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
	diags = AppendAll(f, &cfg.Config, "", diags)

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

// AppendAll analyzes the entire promtail config in memory and transforms it
// into Flow components. It then appends each argument to the file builder.
func AppendAll(f *builder.File, cfg *promtailcfg.Config, labelPrefix string, diags diag.Diagnostics) diag.Diagnostics {
	validateTopLevelConfig(cfg, &diags)

	var writeReceivers = make([]loki.LogsReceiver, len(cfg.ClientConfigs))
	var writeBlocks = make([]*builder.Block, len(cfg.ClientConfigs))
	// Each client config needs to be a separate remote_write,
	// because they may have different ExternalLabels fields.
	for i, cc := range cfg.ClientConfigs {
		writeBlocks[i], writeReceivers[i] = build.NewLokiWrite(&cc, &diags, i, labelPrefix)
	}

	gc := &build.GlobalContext{
		WriteReceivers:   writeReceivers,
		TargetSyncPeriod: cfg.TargetConfig.SyncPeriod,
		LabelPrefix:      labelPrefix,
	}

	for _, sc := range cfg.ScrapeConfig {
		appendScrapeConfig(f, &sc, &diags, gc)
	}

	for _, write := range writeBlocks {
		f.Body().AppendBlock(write)
	}

	return diags
}

func DefaultPositionsConfig() positions.Config {
	// We obtain the default by registering the flags
	cfg := positions.Config{}
	cfg.RegisterFlags(flag.NewFlagSet("", flag.PanicOnError))
	return cfg
}

func DefaultLimitsConfig() limit.Config {
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
	//TODO(thampiotr): need to support/warn about the following fields:
	//Encoding               string                 `mapstructure:"encoding,omitempty" yaml:"encoding,omitempty"`
	//DecompressionCfg       *DecompressionConfig   `yaml:"decompression,omitempty"`

	//TODO(thampiotr): support/warn about the following log producing promtail configs:
	//SyslogConfig         *SyslogTargetConfig         `mapstructure:"syslog,omitempty" yaml:"syslog,omitempty"`
	//GcplogConfig         *GcplogTargetConfig         `mapstructure:"gcplog,omitempty" yaml:"gcplog,omitempty"`
	//WindowsConfig        *WindowsEventsTargetConfig  `mapstructure:"windows_events,omitempty" yaml:"windows_events,omitempty"`
	//KafkaConfig          *KafkaTargetConfig          `mapstructure:"kafka,omitempty" yaml:"kafka,omitempty"`
	//AzureEventHubsConfig *AzureEventHubsTargetConfig `mapstructure:"azure_event_hubs,omitempty" yaml:"azure_event_hubs,omitempty"`
	//GelfConfig           *GelfTargetConfig           `mapstructure:"gelf,omitempty" yaml:"gelf,omitempty"`
	//HerokuDrainConfig    *HerokuDrainTargetConfig    `mapstructure:"heroku_drain,omitempty" yaml:"heroku_drain,omitempty"`

	b := build.NewScrapeConfigBuilder(f, diags, cfg, gctx)
	b.Validate()

	// Append all the SD components
	b.AppendKubernetesSDs()
	b.AppendDockerSDs()
	b.AppendStaticSDs() //TODO
	b.AppendFileSDs()
	b.AppendConsulSDs()
	b.AppendConsulAgentSDs()
	b.AppendDigitalOceanSDs()
	b.AppendGCESDs()
	b.AppendEC2SDs()
	b.AppendAzureSDs()

	// Append loki.source.file to process all SD components' targets.
	// If any relabelling is required, it will be done via a discovery.relabel component.
	// The files will be watched and the globs in file paths will be expanded using discovery.file component.
	// The log entries are sent to loki.process if processing is needed, or directly to loki.write components.
	b.AppendLokiSourceFile()

	// Append all the components that produce logs directly.
	// If any relabelling is required, it will be done via a loki.relabel component.
	// The logs are sent to loki.process if processing is needed, or directly to loki.write components.
	b.AppendCloudFlareConfig()
	b.AppendJournalConfig()
	b.AppendPushAPI()
}
