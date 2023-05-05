package prometheusconvert

import (
	"bytes"
	"fmt"
	"time"

	"github.com/go-kit/log"
	"github.com/grafana/agent/component/common/config"
	"github.com/grafana/agent/component/discovery"
	"github.com/grafana/agent/component/prometheus/scrape"
	"github.com/grafana/agent/converter/internal/schema/riverschema"
	"github.com/grafana/agent/pkg/river/token/builder"
	promconfig "github.com/prometheus/prometheus/config"
	promdiscovery "github.com/prometheus/prometheus/discovery"
	_ "github.com/prometheus/prometheus/discovery/install" // Register Prometheus SDs
	"github.com/prometheus/prometheus/storage"
)

// Convert implements a Prometheus config converter.
func Convert(in []byte) ([]byte, error) {
	promConfig, err := promconfig.Load(string(in), false, log.NewNopLogger())
	if err != nil {
		return nil, err
	}
	f := builder.NewFile()

	for _, scrapeConfig := range promConfig.ScrapeConfigs {
		scrapeBlock := builder.NewBlock([]string{"prometheus", "scrape"}, scrapeConfig.JobName)
		scrapeBlock.Body().SetAttributeValue("targets", getTargets(scrapeConfig))
		scrapeBlock.Body().SetAttributeValue("forward_to", []*riverschema.Capsule{
			riverschema.ExprCapsule("prometheus.remote_write.default.receiver"),
		})

		if scrapeConfig.HonorLabels != scrape.DefaultArguments.HonorLabels {
			scrapeBlock.Body().SetAttributeValue("honor_labels", riverschema.NewBool(scrapeConfig.HonorLabels))
		}

		if scrapeConfig.HonorTimestamps != scrape.DefaultArguments.HonorTimestamps {
			scrapeBlock.Body().SetAttributeValue("honor_timestamps", riverschema.NewBool(scrapeConfig.HonorTimestamps))
		}

		if time.Duration(scrapeConfig.ScrapeInterval) != scrape.DefaultArguments.ScrapeInterval {
			scrapeBlock.Body().SetAttributeValue("scrape_interval", time.Duration(scrapeConfig.ScrapeInterval))
		}

		if time.Duration(scrapeConfig.ScrapeTimeout) != scrape.DefaultArguments.ScrapeTimeout {
			scrapeBlock.Body().SetAttributeValue("scrape_timeout", time.Duration(scrapeConfig.ScrapeTimeout))
		}

		f.Body().AppendBlock(scrapeBlock)
	}

	{
		rw := builder.NewBlock([]string{"prometheus", "remote_write"}, "default")

		for _, rwConf := range promConfig.RemoteWriteConfigs {
			endpoint := builder.NewBlock([]string{"endpoint"}, "")
			endpoint.Body().SetAttributeValue("url", rwConf.URL.String())
			rw.Body().AppendBlock(endpoint)
		}

		f.Body().AppendBlock(rw)
	}

	var buf bytes.Buffer
	if _, err := f.WriteTo(&buf); err != nil {
		return nil, fmt.Errorf("failed to render Flow config: %w", err)
	}
	return buf.Bytes(), nil
}

func getTargets(scrapeConfig *promconfig.ScrapeConfig) []map[string]string {
	targets := []map[string]string{}

	for _, serviceDiscoveryConfig := range scrapeConfig.ServiceDiscoveryConfigs {
		switch sdc := serviceDiscoveryConfig.(type) {
		case promdiscovery.StaticConfig:
			for _, target := range sdc {
				for _, labelSet := range target.Targets {
					for labelName, labelValue := range labelSet {
						targets = append(targets, map[string]string{string(labelName): string(labelValue)})
					}
				}
			}
		}
	}

	return targets
}

// Convert implements a Prometheus config converter.
func Convert2(in []byte) ([]byte, error) {
	promConfig, err := promconfig.Load(string(in), false, log.NewNopLogger())
	if err != nil {
		return nil, err
	}
	f := builder.NewFile()

	for _, scrapeConfig := range promConfig.ScrapeConfigs {
		scrapeArgs := convertScrapeConfig(scrapeConfig)

		// In theory, something like this could be built instead.
		//
		// riverScrapeConfig := scrapeArgs.MarshalRiver()

		scrapeBlock := builder.NewBlock([]string{"prometheus", "scrape"}, scrapeArgs.JobName)
		scrapeBlock.Body().SetAttributeValue("targets", scrapeArgs.Targets)
		scrapeBlock.Body().SetAttributeValue("forward_to", []*riverschema.Capsule{
			riverschema.ExprCapsule("prometheus.remote_write.default.receiver"),
		})

		if scrapeArgs.HonorLabels != scrape.DefaultArguments.HonorLabels {
			scrapeBlock.Body().SetAttributeValue("honor_labels", riverschema.NewBool(scrapeConfig.HonorLabels))
		}

		if scrapeArgs.HonorTimestamps != scrape.DefaultArguments.HonorTimestamps {
			scrapeBlock.Body().SetAttributeValue("honor_timestamps", riverschema.NewBool(scrapeConfig.HonorTimestamps))
		}

		if scrapeArgs.ScrapeInterval != scrape.DefaultArguments.ScrapeInterval {
			scrapeBlock.Body().SetAttributeValue("scrape_interval", scrapeArgs.ScrapeInterval)
		}

		if scrapeArgs.ScrapeTimeout != scrape.DefaultArguments.ScrapeTimeout {
			scrapeBlock.Body().SetAttributeValue("scrape_timeout", scrapeArgs.ScrapeTimeout)
		}

		f.Body().AppendBlock(scrapeBlock)
	}

	{
		rw := builder.NewBlock([]string{"prometheus", "remote_write"}, "default")

		for _, rwConf := range promConfig.RemoteWriteConfigs {
			endpoint := builder.NewBlock([]string{"endpoint"}, "")
			endpoint.Body().SetAttributeValue("url", rwConf.URL.String())
			rw.Body().AppendBlock(endpoint)
		}

		f.Body().AppendBlock(rw)
	}

	var buf bytes.Buffer
	if _, err := f.WriteTo(&buf); err != nil {
		return nil, fmt.Errorf("failed to render Flow config: %w", err)
	}
	return buf.Bytes(), nil
}

func convertScrapeConfig(scrapeConfig *promconfig.ScrapeConfig) *scrape.Arguments {
	return &scrape.Arguments{
		Targets:               getTargets2(scrapeConfig), // TODO
		ForwardTo:             []storage.Appendable{},    // TODO
		JobName:               scrapeConfig.JobName,
		HonorLabels:           scrapeConfig.HonorLabels,
		HonorTimestamps:       scrapeConfig.HonorTimestamps,
		Params:                map[string][]string{}, //TODO
		ScrapeInterval:        time.Duration(scrapeConfig.ScrapeInterval),
		ScrapeTimeout:         time.Duration(scrapeConfig.ScrapeTimeout),
		MetricsPath:           scrapeConfig.MetricsPath,
		Scheme:                scrapeConfig.Scheme,
		BodySizeLimit:         scrapeConfig.BodySizeLimit,
		SampleLimit:           scrapeConfig.SampleLimit,
		TargetLimit:           scrapeConfig.TargetLimit,
		LabelLimit:            scrapeConfig.LabelLimit,
		LabelNameLengthLimit:  scrapeConfig.LabelNameLengthLimit,
		LabelValueLengthLimit: scrapeConfig.LabelValueLengthLimit,
		HTTPClientConfig:      config.HTTPClientConfig{}, // TODO
		ExtraMetrics:          false,                     // TODO
		Clustering:            scrape.Clustering{},       // TODO
	}
}

func getTargets2(scrapeConfig *promconfig.ScrapeConfig) []discovery.Target {
	targets := []discovery.Target{}

	for _, serviceDiscoveryConfig := range scrapeConfig.ServiceDiscoveryConfigs {
		switch sdc := serviceDiscoveryConfig.(type) {
		case promdiscovery.StaticConfig:
			for _, target := range sdc {
				for _, labelSet := range target.Targets {
					for labelName, labelValue := range labelSet {
						targets = append(targets, map[string]string{string(labelName): string(labelValue)})
					}
				}
			}
		}
	}

	return targets
}

// Convert implements a Prometheus config converter.
func Convert3(in []byte) ([]byte, error) {
	promConfig, err := promconfig.Load(string(in), false, log.NewNopLogger())
	if err != nil {
		return nil, err
	}
	f := builder.NewFile()

	for _, scrapeConfig := range promConfig.ScrapeConfigs {
		scrapeArgs := convertScrapeConfig(scrapeConfig)

		scrapeBlock := builder.NewBlock([]string{"prometheus", "scrape"}, scrapeArgs.JobName)
		scrapeBlock.Body().AppendFrom(scrapeArgs)

		f.Body().AppendBlock(scrapeBlock)
	}

	{
		rw := builder.NewBlock([]string{"prometheus", "remote_write"}, "default")

		for _, rwConf := range promConfig.RemoteWriteConfigs {
			endpoint := builder.NewBlock([]string{"endpoint"}, "")
			endpoint.Body().SetAttributeValue("url", rwConf.URL.String())
			rw.Body().AppendBlock(endpoint)
		}

		f.Body().AppendBlock(rw)
	}

	var buf bytes.Buffer
	if _, err := f.WriteTo(&buf); err != nil {
		return nil, fmt.Errorf("failed to render Flow config: %w", err)
	}
	return buf.Bytes(), nil
}
