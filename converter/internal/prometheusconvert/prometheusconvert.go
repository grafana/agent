package prometheusconvert

import (
	"bytes"
	"fmt"

	"github.com/go-kit/log"
	"github.com/grafana/agent/converter/internal/common"
	"github.com/grafana/agent/pkg/river/token/builder"
	promconfig "github.com/prometheus/prometheus/config"
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
func Convert(in []byte) ([]byte, error) {
	promConfig, err := promconfig.Load(string(in), false, log.NewNopLogger())
	if err != nil {
		return nil, err
	}

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
		return nil, fmt.Errorf("failed to render Flow config: %w", err)
	}
	return buf.Bytes(), nil
}
