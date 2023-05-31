package prometheusconvert

import (
	"bytes"
	"fmt"

	"github.com/go-kit/log"
	"github.com/grafana/agent/converter/diag"
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
func Convert(in []byte) ([]byte, diag.Diagnostics) {
	var diags diag.Diagnostics

	promConfig, err := promconfig.Load(string(in), false, log.NewNopLogger())
	if err != nil {
		diags.Add(diag.SeverityLevelError, fmt.Sprintf("failed to parse Prometheus config: %s", err))
		return nil, diags
	}

	diags = ValidateUnsupported(promConfig)

	f := builder.NewFile()
	AppendAll(f, promConfig)

	var buf bytes.Buffer
	if _, err := f.WriteTo(&buf); err != nil {
		diags.Add(diag.SeverityLevelError, fmt.Sprintf("failed to render Flow config: %s", err.Error()))
		return nil, diags
	}
	return buf.Bytes(), diags
}

// ValidateUnsupported will traverse the Prometheus Config and return warnings
// for any config we knowingly do not support.
func ValidateUnsupported(promConfig *promconfig.Config) diag.Diagnostics {
	var diags diag.Diagnostics

	for _, scrapeConfig := range promConfig.ScrapeConfigs {
		for _, sdConfig := range scrapeConfig.ServiceDiscoveryConfigs {
			switch sdConfig.(type) {
			case discovery.StaticConfig:
				continue
			default:
				diags.Add(diag.SeverityLevelWarn, fmt.Sprintf("unsupported service discovery %s was provided", sdConfig.Name()))
			}
		}
	}

	return diags
}

// AppendAll analyzes the entire prometheus config in memory and transforms it
// into Flow Arguments. It then appends each argument to the file builder.
func AppendAll(f *builder.File, promConfig *promconfig.Config) {
	remoteWriteExports := appendRemoteWrite(f, promConfig)
	appendScrape(f, promConfig.ScrapeConfigs, []storage.Appendable{remoteWriteExports.Receiver})
}
