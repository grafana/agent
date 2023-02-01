package kubernetes_crds

import (
	"fmt"
	"net/url"

	"github.com/go-kit/log/level"
	v1 "github.com/prometheus-operator/prometheus-operator/pkg/apis/monitoring/v1"
	commonConfig "github.com/prometheus/common/config"
	"github.com/prometheus/common/model"
	"github.com/prometheus/prometheus/config"
	"github.com/prometheus/prometheus/model/relabel"
)

func (cg *configGenerator) generateProbeConfig(m *v1.Probe) *config.ScrapeConfig {
	c := config.DefaultScrapeConfig
	cfg := &c
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
			level.Error(cg.logger).Log("msg", "failed to parse ProxyURL from probe", "err", err)
		} else {
			cfg.HTTPClientConfig.ProxyURL = commonConfig.URL{URL: u}
		}
	}
	if m.Spec.Module != "" {
		cfg.Params.Set("module", m.Spec.Module)
	}

	relabels := cg.initRelabelings(cfg)

	if m.Spec.JobName != "" {
		relabels.Add(&relabel.Config{
			TargetLabel: "job",
			Replacement: m.Spec.JobName,
		})
	}
	//labeler := namespacelabeler.New("", nil, false)
	// TODO: staticConfig

	// TODO: limits from spec
	return cfg
}

// 	if m.Spec.Targets.StaticConfig != nil {
// 		// Generate static_config section.
// 		staticConfig := yaml.MapSlice{
// 			{Key: "targets", Value: m.Spec.Targets.StaticConfig.Targets},
// 		}

// 		if m.Spec.Targets.StaticConfig.Labels != nil {
// 			if _, ok := m.Spec.Targets.StaticConfig.Labels["namespace"]; !ok {
// 				m.Spec.Targets.StaticConfig.Labels["namespace"] = m.Namespace
// 			}
// 		} else {
// 			m.Spec.Targets.StaticConfig.Labels = map[string]string{"namespace": m.Namespace}
// 		}

// 		staticConfig = append(staticConfig, yaml.MapSlice{
// 			{Key: "labels", Value: m.Spec.Targets.StaticConfig.Labels},
// 		}...)

// 		cfg = append(cfg, yaml.MapItem{
// 			Key:   "static_configs",
// 			Value: []yaml.MapSlice{staticConfig},
// 		})

// 		// Relabelings for prober.
// 		relabelings = append(relabelings, []yaml.MapSlice{
// 			{
// 				{Key: "source_labels", Value: []string{"__address__"}},
// 				{Key: "target_label", Value: "__param_target"},
// 			},
// 			{
// 				{Key: "source_labels", Value: []string{"__param_target"}},
// 				{Key: "target_label", Value: "instance"},
// 			},
// 			{
// 				{Key: "target_label", Value: "__address__"},
// 				{Key: "replacement", Value: m.Spec.ProberSpec.URL},
// 			},
// 		}...)

// 		// Add configured relabelings.
// 		xc := labeler.GetRelabelingConfigs(m.TypeMeta, m.ObjectMeta, m.Spec.Targets.StaticConfig.RelabelConfigs)
// 		relabelings = append(relabelings, generateRelabelConfig(xc)...)
// 		cfg = append(cfg, yaml.MapItem{Key: "relabel_configs", Value: relabelings})
// 	} else {
// 		// Generate kubernetes_sd_config section for the ingress resources.

// 		// Filter targets by ingresses selected by the monitor.
// 		// Exact label matches.
// 		labelKeys := make([]string, 0, len(m.Spec.Targets.Ingress.Selector.MatchLabels))
// 		for k := range m.Spec.Targets.Ingress.Selector.MatchLabels {
// 			labelKeys = append(labelKeys, k)
// 		}
// 		sort.Strings(labelKeys)

// 		for _, k := range labelKeys {
// 			relabelings = append(relabelings, yaml.MapSlice{
// 				{Key: "action", Value: "keep"},
// 				{Key: "source_labels", Value: []string{"__meta_kubernetes_ingress_label_" + sanitizeLabelName(k), "__meta_kubernetes_ingress_labelpresent_" + sanitizeLabelName(k)}},
// 				{Key: "regex", Value: fmt.Sprintf("(%s);true", m.Spec.Targets.Ingress.Selector.MatchLabels[k])},
// 			})
// 		}

// 		// Set based label matching. We have to map the valid relations
// 		// `In`, `NotIn`, `Exists`, and `DoesNotExist`, into relabeling rules.
// 		for _, exp := range m.Spec.Targets.Ingress.Selector.MatchExpressions {
// 			switch exp.Operator {
// 			case metav1.LabelSelectorOpIn:
// 				relabelings = append(relabelings, yaml.MapSlice{
// 					{Key: "action", Value: "keep"},
// 					{Key: "source_labels", Value: []string{"__meta_kubernetes_ingress_label_" + sanitizeLabelName(exp.Key), "__meta_kubernetes_ingress_labelpresent_" + sanitizeLabelName(exp.Key)}},
// 					{Key: "regex", Value: fmt.Sprintf("(%s);true", strings.Join(exp.Values, "|"))},
// 				})
// 			case metav1.LabelSelectorOpNotIn:
// 				relabelings = append(relabelings, yaml.MapSlice{
// 					{Key: "action", Value: "drop"},
// 					{Key: "source_labels", Value: []string{"__meta_kubernetes_ingress_label_" + sanitizeLabelName(exp.Key), "__meta_kubernetes_ingress_labelpresent_" + sanitizeLabelName(exp.Key)}},
// 					{Key: "regex", Value: fmt.Sprintf("(%s);true", strings.Join(exp.Values, "|"))},
// 				})
// 			case metav1.LabelSelectorOpExists:
// 				relabelings = append(relabelings, yaml.MapSlice{
// 					{Key: "action", Value: "keep"},
// 					{Key: "source_labels", Value: []string{"__meta_kubernetes_ingress_labelpresent_" + sanitizeLabelName(exp.Key)}},
// 					{Key: "regex", Value: "true"},
// 				})
// 			case metav1.LabelSelectorOpDoesNotExist:
// 				relabelings = append(relabelings, yaml.MapSlice{
// 					{Key: "action", Value: "drop"},
// 					{Key: "source_labels", Value: []string{"__meta_kubernetes_ingress_labelpresent_" + sanitizeLabelName(exp.Key)}},
// 					{Key: "regex", Value: "true"},
// 				})
// 			}
// 		}

// 		cfg = append(cfg, cg.generateK8SSDConfig(m.Spec.Targets.Ingress.NamespaceSelector, m.Namespace, apiserverConfig, store, kubernetesSDRoleIngress, nil))

// 		// Relabelings for ingress SD.
// 		relabelings = append(relabelings, []yaml.MapSlice{
// 			{
// 				{Key: "source_labels", Value: []string{"__meta_kubernetes_ingress_scheme", "__address__", "__meta_kubernetes_ingress_path"}},
// 				{Key: "separator", Value: ";"},
// 				{Key: "regex", Value: "(.+);(.+);(.+)"},
// 				{Key: "target_label", Value: "__param_target"},
// 				{Key: "replacement", Value: "${1}://${2}${3}"},
// 				{Key: "action", Value: "replace"},
// 			},
// 			{
// 				{Key: "source_labels", Value: []string{"__meta_kubernetes_namespace"}},
// 				{Key: "target_label", Value: "namespace"},
// 			},
// 			{
// 				{Key: "source_labels", Value: []string{"__meta_kubernetes_ingress_name"}},
// 				{Key: "target_label", Value: "ingress"},
// 			},
// 		}...)

// 		// Relabelings for prober.
// 		relabelings = append(relabelings, []yaml.MapSlice{
// 			{
// 				{Key: "source_labels", Value: []string{"__address__"}},
// 				{Key: "separator", Value: ";"},
// 				{Key: "regex", Value: "(.*)"},
// 				{Key: "target_label", Value: "__tmp_ingress_address"},
// 				{Key: "replacement", Value: "$1"},
// 				{Key: "action", Value: "replace"},
// 			},
// 			{
// 				{Key: "source_labels", Value: []string{"__param_target"}},
// 				{Key: "target_label", Value: "instance"},
// 			},
// 			{
// 				{Key: "target_label", Value: "__address__"},
// 				{Key: "replacement", Value: m.Spec.ProberSpec.URL},
// 			},
// 		}...)

// 		// Add configured relabelings.
// 		relabelings = append(relabelings, generateRelabelConfig(labeler.GetRelabelingConfigs(m.TypeMeta, m.ObjectMeta, m.Spec.Targets.Ingress.RelabelConfigs))...)
// 		relabelings = generateAddressShardingRelabelingRulesForProbes(relabelings, shards)

// 		cfg = append(cfg, yaml.MapItem{Key: "relabel_configs", Value: relabelings})

// 	}

// 	if m.Spec.TLSConfig != nil {
// 		cfg = addSafeTLStoYaml(cfg, m.Namespace, m.Spec.TLSConfig.SafeTLSConfig)
// 	}

// 	if m.Spec.BearerTokenSecret.Name != "" {
// 		pnKey := fmt.Sprintf("probe/%s/%s", m.GetNamespace(), m.GetName())
// 		if s, ok := store.TokenAssets[pnKey]; ok {
// 			cfg = append(cfg, yaml.MapItem{Key: "bearer_token", Value: s})
// 		}
// 	}

// 	cfg = cg.addBasicAuthToYaml(cfg, fmt.Sprintf("probe/%s/%s", m.Namespace, m.Name), store, m.Spec.BasicAuth)

// 	assetKey := fmt.Sprintf("probe/%s/%s", m.Namespace, m.Name)
// 	cfg = cg.addOAuth2ToYaml(cfg, m.Spec.OAuth2, store.OAuth2Assets, assetKey)

// 	cfg = cg.addSafeAuthorizationToYaml(cfg, fmt.Sprintf("probe/auth/%s/%s", m.Namespace, m.Name), store, m.Spec.Authorization)

// 	cfg = append(cfg, yaml.MapItem{Key: "metric_relabel_configs", Value: generateRelabelConfig(labeler.GetRelabelingConfigs(m.TypeMeta, m.ObjectMeta, m.Spec.MetricRelabelConfigs))})

// 	return cfg
