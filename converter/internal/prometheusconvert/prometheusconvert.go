package prometheusconvert

import (
	"bytes"
	"fmt"

	"github.com/go-kit/log"
	"github.com/grafana/agent/converter/internal/common"
	"github.com/grafana/agent/pkg/river/token/builder"
	promconfig "github.com/prometheus/prometheus/config"
	"github.com/prometheus/prometheus/discovery"
	"github.com/prometheus/prometheus/storage"

	_ "github.com/prometheus/prometheus/discovery/install" // Register Prometheus SDs
)

// Convert implements a Prometheus config converter.
//
// TODO...
// The implementation of this API is a work in progress.
// Additional components must be implemented:
//
//	prometheus.relabel
//	discovery.azure
//	discovery.consul
//	discovery.digitalocean
//	discovery.dns
//	discovery.docker
//	discovery.ec2
//	discovery.file
//	discovery.gce
//	discovery.kubernetes
//	discovery.lightsail
//	discovery.relabel
func Convert(in []byte) ([]byte, common.Diagnostics) {
	var diags common.Diagnostics

	promConfig, err := promconfig.Load(string(in), false, log.NewNopLogger())
	if err != nil {
		diags.Add(common.SeverityLevelError, fmt.Sprintf("failed to parse Prometheus config: %s", err))
		return nil, diags
	}

	diags = ValidateUnsupported(promConfig)

	f := builder.NewFile()

	remoteWriteArgs := toRemotewriteArguments(promConfig)
	common.AppendBlockWithOverride(f, []string{"prometheus", "remote_write"}, "default", remoteWriteArgs)

	forwardTo := make([]storage.Appendable, 0)
	forwardTo = append(forwardTo, common.ConvertAppendable{Expr: "prometheus.remote_write.default.receiver"})
	for _, scrapeConfig := range promConfig.ScrapeConfigs {
		scrapeArgs := toScrapeArguments(scrapeConfig, forwardTo)
		common.AppendBlockWithOverride(f, []string{"prometheus", "scrape"}, scrapeArgs.JobName, scrapeArgs)
	}

	var buf bytes.Buffer
	if _, err := f.WriteTo(&buf); err != nil {
		diags.Add(common.SeverityLevelError, fmt.Sprintf("failed to render Flow config: %s", err.Error()))
		return nil, diags
	}
	return buf.Bytes(), diags
}

// ValidateUnsupported will traverse the Prometheus Config and return warnings
// for any config we knowingly do not support.
func ValidateUnsupported(promConfig *promconfig.Config) common.Diagnostics {
	var diags common.Diagnostics

	for _, scrapeConfig := range promConfig.ScrapeConfigs {
		for _, sdConfig := range scrapeConfig.ServiceDiscoveryConfigs {
			switch sdConfig.(type) {
			case discovery.StaticConfig:
				continue
			default:
				diags.Add(common.SeverityLevelWarn, fmt.Sprintf("unsupported service discovery %s was provided", sdConfig.Name()))
			}
		}
	}

	return diags
}
