package prometheusconvert

import (
	"bytes"
	"fmt"

	"github.com/go-kit/log"
	"github.com/grafana/agent/component/discovery"
	"github.com/grafana/agent/converter/diag"
	"github.com/grafana/agent/pkg/river/token/builder"
	promconfig "github.com/prometheus/prometheus/config"
	promdiscover "github.com/prometheus/prometheus/discovery"
	promazure "github.com/prometheus/prometheus/discovery/azure"
	"github.com/prometheus/prometheus/storage"

	_ "github.com/prometheus/prometheus/discovery/install" // Register Prometheus SDs
)

// Convert implements a Prometheus config converter.
//
// TODO...
// The implementation of this API is a work in progress.
// Additional components must be implemented:
//
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

	f := builder.NewFile()
	diags = AppendAll(f, promConfig)

	var buf bytes.Buffer
	if _, err := f.WriteTo(&buf); err != nil {
		diags.Add(diag.SeverityLevelError, fmt.Sprintf("failed to render Flow config: %s", err.Error()))
		return nil, diags
	}
	return buf.Bytes(), diags
}

// AppendAll analyzes the entire prometheus config in memory and transforms it
// into Flow Arguments. It then appends each argument to the file builder.
func AppendAll(f *builder.File, promConfig *promconfig.Config) diag.Diagnostics {
	var diags diag.Diagnostics
	remoteWriteExports := appendRemoteWrite(f, promConfig)

	forwardTo := []storage.Appendable{remoteWriteExports.Receiver}
	for _, scrapeConfig := range promConfig.ScrapeConfigs {
		relabelExports := appendRelabel(f, scrapeConfig.RelabelConfigs, forwardTo, scrapeConfig.JobName)
		if relabelExports != nil {
			forwardTo = []storage.Appendable{relabelExports.Receiver}
		}

		var targets []discovery.Target
		for _, serviceDiscoveryConfig := range scrapeConfig.ServiceDiscoveryConfigs {
			switch sdc := serviceDiscoveryConfig.(type) {
			case promdiscover.StaticConfig:
				continue
			case *promazure.SDConfig:
				_, newDiags := appendDiscoveryAzure(f, scrapeConfig.JobName, sdc)
				// exports, newDiags := appendDiscoveryAzure(f, scrapeConfig.JobName, sdc)
				// targets = append(targets, exports.Targets)
				diags = append(diags, newDiags...)
			default:
				diags.Add(diag.SeverityLevelWarn, fmt.Sprintf("unsupported service discovery %s was provided", sdc.Name()))
			}
		}

		for _, serviceDiscoveryConfig := range scrapeConfig.ServiceDiscoveryConfigs {
			switch sdc := serviceDiscoveryConfig.(type) {
			case promdiscover.StaticConfig:
				targets = append(targets, getScrapeTargets(sdc)...)
			}
		}

		appendScrape(f, scrapeConfig, forwardTo, targets)
	}

	return diags
}
