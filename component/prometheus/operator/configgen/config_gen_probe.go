package configgen

import (
	"fmt"
	"net/url"

	promopv1 "github.com/prometheus-operator/prometheus-operator/pkg/apis/monitoring/v1"
	namespacelabeler "github.com/prometheus-operator/prometheus-operator/pkg/namespace-labeler"
	commonConfig "github.com/prometheus/common/config"
	"github.com/prometheus/common/model"
	"github.com/prometheus/prometheus/config"
	"github.com/prometheus/prometheus/discovery"
	"github.com/prometheus/prometheus/discovery/targetgroup"
	"github.com/prometheus/prometheus/model/relabel"
)

// See https://github.com/prometheus-operator/prometheus-operator/blob/aa8222d7e9b66e9293ed11c9291ea70173021029/pkg/prometheus/promcfg.go#L835

func (cg *ConfigGenerator) GenerateProbeConfig(m *promopv1.Probe) (cfg *config.ScrapeConfig, err error) {
	c := config.DefaultScrapeConfig
	cfg = &c
	cfg.ScrapeInterval = config.DefaultGlobalConfig.ScrapeInterval
	cfg.ScrapeTimeout = config.DefaultGlobalConfig.ScrapeTimeout
	cfg.JobName = fmt.Sprintf("probe/%s/%s", m.Namespace, m.Name)
	cfg.HonorTimestamps = true
	cfg.MetricsPath = m.Spec.ProberSpec.Path
	if m.Spec.Interval != "" {
		cfg.ScrapeInterval, _ = model.ParseDuration(string(m.Spec.Interval))
	}
	if m.Spec.ScrapeTimeout != "" {
		cfg.ScrapeInterval, _ = model.ParseDuration(string(m.Spec.ScrapeTimeout))
	}
	if m.Spec.ProberSpec.Scheme != "" {
		cfg.Scheme = m.Spec.ProberSpec.Scheme
	}
	if m.Spec.ProberSpec.ProxyURL != "" {
		if u, err := url.Parse(m.Spec.ProberSpec.ProxyURL); err != nil {
			return nil, fmt.Errorf("parsing ProxyURL from probe: %w", err)
		} else {
			cfg.HTTPClientConfig.ProxyURL = commonConfig.URL{URL: u}
		}
	}
	if m.Spec.Module != "" {
		cfg.Params.Set("module", m.Spec.Module)
	}

	cfg.SampleLimit = uint(m.Spec.SampleLimit)
	cfg.TargetLimit = uint(m.Spec.TargetLimit)
	cfg.LabelLimit = uint(m.Spec.LabelLimit)
	cfg.LabelNameLengthLimit = uint(m.Spec.LabelNameLengthLimit)
	cfg.LabelValueLengthLimit = uint(m.Spec.LabelValueLengthLimit)

	relabels := cg.initRelabelings()
	if m.Spec.JobName != "" {
		relabels.add(&relabel.Config{
			Replacement: m.Spec.JobName,
			TargetLabel: "job",
		})
	}
	labeler := namespacelabeler.New("", nil, false)

	static := m.Spec.Targets.StaticConfig
	if static != nil {
		// Generate static_config section.
		grp := &targetgroup.Group{
			Labels: model.LabelSet{},
		}
		for k, v := range static.Labels {
			grp.Labels[model.LabelName(k)] = model.LabelValue(v)
		}
		for _, t := range static.Targets {
			grp.Targets = append(grp.Targets, model.LabelSet{
				model.AddressLabel: model.LabelValue(t),
			})
		}
		sc := discovery.StaticConfig{grp}
		cfg.ServiceDiscoveryConfigs = append(cfg.ServiceDiscoveryConfigs, sc)
		if err = relabels.addFromV1(labeler.GetRelabelingConfigs(m.TypeMeta, m.ObjectMeta, m.Spec.Targets.StaticConfig.RelabelConfigs)...); err != nil {
			return nil, fmt.Errorf("parsing relabel configs: %w", err)
		}
	} else {
		// Generate kubernetes_sd_config section for the ingress resources.
	}

	cfg.RelabelConfigs = relabels.configs
	return cfg, nil
}
