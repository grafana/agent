package kubernetes_crds

// SEE https://github.com/prometheus-operator/prometheus-operator/blob/main/pkg/prometheus/promcfg.go

import (
	"fmt"
	"net/url"
	"regexp"
	"sort"
	"strings"

	v1 "github.com/prometheus-operator/prometheus-operator/pkg/apis/monitoring/v1"
	commonConfig "github.com/prometheus/common/config"
	"github.com/prometheus/common/model"
	"github.com/prometheus/prometheus/config"
	promk8s "github.com/prometheus/prometheus/discovery/kubernetes"
	"github.com/prometheus/prometheus/model/relabel"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func generateK8SSDConfig(
	namespaceSelector v1.NamespaceSelector,
	namespace string,
	//apiserverConfig *v1.APIServerConfig,
	//store *assets.Store,
	role promk8s.Role,
	attachMetadata *v1.AttachMetadata,
) *promk8s.SDConfig {
	cfg := &promk8s.SDConfig{
		Role: role,
	}

	namespaces := getNamespacesFromNamespaceSelector(namespaceSelector, namespace)
	if len(namespaces) != 0 {
		cfg.NamespaceDiscovery.Names = namespaces
	}

	// if apiserverConfig != nil {
	// 	k8sSDConfig = append(k8sSDConfig, yaml.MapItem{
	// 		Key: "api_server", Value: apiserverConfig.Host,
	// 	})

	// 	if apiserverConfig.BasicAuth != nil && store.BasicAuthAssets != nil {
	// 		if s, ok := store.BasicAuthAssets["apiserver"]; ok {
	// 			k8sSDConfig = append(k8sSDConfig, yaml.MapItem{
	// 				Key: "basic_auth", Value: yaml.MapSlice{
	// 					{Key: "username", Value: s.Username},
	// 					{Key: "password", Value: s.Password},
	// 				},
	// 			})
	// 		}
	// 	}

	// 	if apiserverConfig.BearerToken != "" {
	// 		k8sSDConfig = append(k8sSDConfig, yaml.MapItem{Key: "bearer_token", Value: apiserverConfig.BearerToken})
	// 	}

	// 	if apiserverConfig.BearerTokenFile != "" {
	// 		k8sSDConfig = append(k8sSDConfig, yaml.MapItem{Key: "bearer_token_file", Value: apiserverConfig.BearerTokenFile})
	// 	}

	// 	k8sSDConfig = cg.addAuthorizationToYaml(k8sSDConfig, "apiserver/auth", store, apiserverConfig.Authorization)

	// 	// TODO: If we want to support secret refs for k8s service discovery tls
	// 	// config as well, make sure to path the right namespace here.
	// 	k8sSDConfig = addTLStoYaml(k8sSDConfig, "", apiserverConfig.TLSConfig)
	// }

	// TODO:
	if attachMetadata != nil {

		//k8sSDConfig = cg.WithMinimumVersion("2.35.0").AppendMapItem(k8sSDConfig, "attach_metadata", yaml.MapSlice{
		//	{Key: "node", Value: attachMetadata.Node},
		//})
	}
	return cfg
}

func generatePodMonitorConfig(m *v1.PodMonitor, ep v1.PodMetricsEndpoint, i int) *config.ScrapeConfig {
	cfg := &config.ScrapeConfig{}
	cfg.JobName = fmt.Sprintf("podMonitor/%s/%s/%d", m.Namespace, m.Name, i)
	addHonorLabels(cfg, ep.HonorLabels)
	addHonorTimestamps(cfg, ep.HonorTimestamps)

	cfg.ServiceDiscoveryConfigs = append(cfg.ServiceDiscoveryConfigs, generateK8SSDConfig(m.Spec.NamespaceSelector, m.Namespace, promk8s.RolePod, m.Spec.AttachMetadata))

	if ep.Interval != "" {
		// TODO: correct way to convert the durations?
		cfg.ScrapeInterval, _ = model.ParseDuration(string(ep.Interval))
	}
	if ep.ScrapeTimeout != "" {
		cfg.ScrapeInterval, _ = model.ParseDuration(string(ep.ScrapeTimeout))
	}
	if ep.Path != "" {
		cfg.MetricsPath = ep.Path
	}
	if ep.ProxyURL != nil {
		u, _ := url.Parse(*ep.ProxyURL)
		cfg.HTTPClientConfig.ProxyURL = commonConfig.URL{URL: u}
	}
	if ep.Params != nil {
		cfg.Params = ep.Params
	}
	if ep.Scheme != "" {
		cfg.Scheme = ep.Scheme
	}
	if ep.FollowRedirects != nil {
		cfg.HTTPClientConfig.FollowRedirects = *ep.FollowRedirects
	}
	if ep.EnableHttp2 != nil {
		cfg.HTTPClientConfig.EnableHTTP2 = *ep.EnableHttp2
	}
	//if ep.TLSConfig != nil {
	//TODO:
	//cfg = addSafeTLStoYaml(cfg, m.Namespace, ep.TLSConfig.SafeTLSConfig)
	//}

	//TODO: Secret store needs to be figured out
	// if ep.BearerTokenSecret.Name != "" {
	// 	if s, ok := store.TokenAssets[fmt.Sprintf("podMonitor/%s/%s/%d", m.Namespace, m.Name, i)]; ok {
	// 		cfg = append(cfg, yaml.MapItem{Key: "bearer_token", Value: s})
	// 	}
	// }

	// TODO:
	// if ep.BasicAuth != nil {
	// 	if s, ok := store.BasicAuthAssets[fmt.Sprintf("podMonitor/%s/%s/%d", m.Namespace, m.Name, i)]; ok {
	// 		cfg = append(cfg, yaml.MapItem{
	// 			Key: "basic_auth", Value: yaml.MapSlice{
	// 				{Key: "username", Value: s.Username},
	// 				{Key: "password", Value: s.Password},
	// 			},
	// 		})
	// 	}
	// }

	// TODO:
	//assetKey := fmt.Sprintf("podMonitor/%s/%s/%d", m.Namespace, m.Name, i)
	//cfg = cg.addOAuth2ToYaml(cfg, ep.OAuth2, store.OAuth2Assets, assetKey)
	//cfg = cg.addSafeAuthorizationToYaml(cfg, fmt.Sprintf("podMonitor/auth/%s/%s/%d", m.Namespace, m.Name, i), store, ep.Authorization)
	relabels := initRelabelings(cfg)

	if ep.FilterRunning == nil || *ep.FilterRunning {
		relabels = append(relabels, &relabel.Config{
			SourceLabels: model.LabelNames{"__meta_kubernetes_pod_phase"},
			Action:       "drop",
			// TODO: maybe mustNewRegexp needs error handling
			Regex: relabel.MustNewRegexp("(Failed|Succeeded)"),
		})
	}

	var labelKeys []string
	// Filter targets by pods selected by the monitor.
	// Exact label matches.
	for k := range m.Spec.Selector.MatchLabels {
		labelKeys = append(labelKeys, k)
	}
	sort.Strings(labelKeys)

	for _, k := range labelKeys {
		relabels = append(relabels, &relabel.Config{
			SourceLabels: model.LabelNames{"__meta_kubernetes_pod_label_" + sanitizeLabelName(k), "__meta_kubernetes_pod_labelpresent_" + sanitizeLabelName(k)},
			Action:       "keep",
			Regex:        relabel.MustNewRegexp(fmt.Sprintf("(%s);true", m.Spec.Selector.MatchLabels[k])),
		})
	}

	// Set based label matching. We have to map the valid relations
	// `In`, `NotIn`, `Exists`, and `DoesNotExist`, into relabeling rules.
	for _, exp := range m.Spec.Selector.MatchExpressions {
		switch exp.Operator {
		case metav1.LabelSelectorOpIn:
			relabels = append(relabels, &relabel.Config{
				SourceLabels: model.LabelNames{"__meta_kubernetes_pod_label_" + sanitizeLabelName(exp.Key), "__meta_kubernetes_pod_labelpresent_" + sanitizeLabelName(exp.Key)},
				Action:       "keep",
				Regex:        relabel.MustNewRegexp(fmt.Sprintf("(%s);true", strings.Join(exp.Values, "|"))),
			})
		case metav1.LabelSelectorOpNotIn:
			relabels = append(relabels, &relabel.Config{
				SourceLabels: model.LabelNames{"__meta_kubernetes_pod_label_" + sanitizeLabelName(exp.Key), "__meta_kubernetes_pod_labelpresent_" + sanitizeLabelName(exp.Key)},
				Action:       "drop",
				Regex:        relabel.MustNewRegexp(fmt.Sprintf("(%s);true", strings.Join(exp.Values, "|"))),
			})
		case metav1.LabelSelectorOpExists:
			relabels = append(relabels, &relabel.Config{
				SourceLabels: model.LabelNames{"__meta_kubernetes_pod_labelpresent_" + sanitizeLabelName(exp.Key)},
				Action:       "keep",
				Regex:        relabel.MustNewRegexp("true"),
			})
		case metav1.LabelSelectorOpDoesNotExist:
			relabels = append(relabels, &relabel.Config{
				SourceLabels: model.LabelNames{"__meta_kubernetes_pod_labelpresent_" + sanitizeLabelName(exp.Key)},
				Action:       "drop",
				Regex:        relabel.MustNewRegexp("true"),
			})
		}
	}

	// Filter targets based on correct port for the endpoint.
	if ep.Port != "" {
		relabels = append(relabels, &relabel.Config{
			SourceLabels: model.LabelNames{"__meta_kubernetes_pod_container_port_name"},
			Action:       "keep",
			Regex:        relabel.MustNewRegexp(ep.Port),
		})
	} else if ep.TargetPort != nil { //nolint:staticcheck // Ignore SA1019 this field is marked as deprecated.
		//TODO: logging
		// 	level.Warn(cg.logger).Log("msg", "'targetPort' is deprecated, use 'port' instead.")
		//nolint:staticcheck // Ignore SA1019 this field is marked as deprecated.
		if ep.TargetPort.StrVal != "" {
			relabels = append(relabels, &relabel.Config{
				SourceLabels: model.LabelNames{"__meta_kubernetes_pod_container_port_name"},
				Action:       "keep",
				Regex:        relabel.MustNewRegexp(ep.TargetPort.String()),
			})
		}
	} else if ep.TargetPort.IntVal != 0 { //nolint:staticcheck // Ignore SA1019 this field is marked as deprecated.
		relabels = append(relabels, &relabel.Config{
			SourceLabels: model.LabelNames{"__meta_kubernetes_pod_container_port_number"},
			Action:       "keep",
			Regex:        relabel.MustNewRegexp(ep.TargetPort.String()),
		})
	}

	// Relabel namespace and pod and service labels into proper labels.
	relabels = append(relabels, &relabel.Config{
		SourceLabels: model.LabelNames{"__meta_kubernetes_namespace"},
		TargetLabel:  "namespace",
	}, &relabel.Config{
		SourceLabels: model.LabelNames{"__meta_kubernetes_pod_container_name"},
		TargetLabel:  "container",
	}, &relabel.Config{
		SourceLabels: model.LabelNames{"__meta_kubernetes_pod_name"},
		TargetLabel:  "pod",
	})

	// Relabel targetLabels from Pod onto target.
	for _, l := range m.Spec.PodTargetLabels {
		relabels = append(relabels, &relabel.Config{
			SourceLabels: model.LabelNames{"__meta_kubernetes_pod_label_" + sanitizeLabelName(l)},
			Replacement:  "${1}",
			Regex:        relabel.MustNewRegexp("(.+)"),
			TargetLabel:  string(sanitizeLabelName(l)),
		})
	}

	// By default, generate a safe job name from the PodMonitor. We also keep
	// this around if a jobLabel is set in case the targets don't actually have a
	// value for it. A single pod may potentially have multiple metrics
	// endpoints, therefore the endpoints labels is filled with the ports name or
	// as a fallback the port number.

	relabels = append(relabels, &relabel.Config{
		Replacement: fmt.Sprintf("%s/%s", m.GetNamespace(), m.GetName()),
		TargetLabel: "job",
	})
	if m.Spec.JobLabel != "" {
		relabels = append(relabels, &relabel.Config{
			Replacement:  "${1}",
			TargetLabel:  "job",
			Regex:        relabel.MustNewRegexp("(.+)"),
			SourceLabels: model.LabelNames{"__meta_kubernetes_pod_label_" + sanitizeLabelName(m.Spec.JobLabel)},
		})
	}

	if ep.Port != "" {
		relabels = append(relabels, &relabel.Config{
			Replacement: ep.Port,
			TargetLabel: "endpoint",
		})
	} else if ep.TargetPort != nil && ep.TargetPort.String() != "" { //nolint:staticcheck // Ignore SA1019 this field is marked as deprecated.
		relabels = append(relabels, &relabel.Config{
			Replacement: ep.Port,
			TargetLabel: ep.TargetPort.String(), //nolint:staticcheck // Ignore SA1019 this field is marked as deprecated.
		})
	}

	// TODO: more relabeling including global config stuff. And some sharding questions
	// labeler := namespacelabeler.New(cg.spec.EnforcedNamespaceLabel, cg.spec.ExcludedFromEnforcement, false)
	//relabelings = append(relabelings, generateRelabelConfig(labeler.GetRelabelingConfigs(m.TypeMeta, m.ObjectMeta, ep.RelabelConfigs))...)
	// relabelings = generateAddressShardingRelabelingRules(relabelings, shards)

	cfg.RelabelConfigs = relabels

	// TODO: limits include stuff from global config
	// cfg = cg.AddLimitsToYAML(cfg, sampleLimitKey, m.Spec.SampleLimit, cg.spec.EnforcedSampleLimit)
	// cfg = cg.AddLimitsToYAML(cfg, targetLimitKey, m.Spec.TargetLimit, cg.spec.EnforcedTargetLimit)
	// cfg = cg.AddLimitsToYAML(cfg, labelLimitKey, m.Spec.LabelLimit, cg.spec.EnforcedLabelLimit)
	// cfg = cg.AddLimitsToYAML(cfg, labelNameLengthLimitKey, m.Spec.LabelNameLengthLimit, cg.spec.EnforcedLabelNameLengthLimit)
	// cfg = cg.AddLimitsToYAML(cfg, labelValueLengthLimitKey, m.Spec.LabelValueLengthLimit, cg.spec.EnforcedLabelValueLengthLimit)
	// if cg.spec.EnforcedBodySizeLimit != "" {
	// 	cfg = cg.WithMinimumVersion("2.28.0").AppendMapItem(cfg, "body_size_limit", cg.spec.EnforcedBodySizeLimit)
	// }

	// TODO: metric relabeling configs are a little tricky
	// cfg = append(cfg, yaml.MapItem{Key: "metric_relabel_configs", Value: generateRelabelConfig(labeler.GetRelabelingConfigs(m.TypeMeta, m.ObjectMeta, ep.MetricRelabelConfigs))})

	return cfg
}

func initRelabelings(cfg *config.ScrapeConfig) []*relabel.Config {
	// Relabel prometheus job name into a meta label
	return []*relabel.Config{
		{
			SourceLabels: model.LabelNames{"job"},
			TargetLabel:  "__tmp_prometheus_job_name",
		},
	}
}

// addHonorTimestamps adds the honor_timestamps field into scrape configurations.
// honor_timestamps is false only when the user specified it or when the global
// override applies.
// For backwards compatibility with Prometheus <2.9.0 we don't set
// honor_timestamps.
func addHonorTimestamps(cfg *config.ScrapeConfig, userHonorTimestamps *bool) {
	//TODO: for now I haven't added the full configGenerator concept. We may still need some of this global config
	// Fast path.
	if userHonorTimestamps == nil { //&& !cg.spec.OverrideHonorTimestamps {
		return
	}
	honor := false
	if userHonorTimestamps != nil {
		honor = *userHonorTimestamps
	}
	cfg.HonorTimestamps = honor
	//return cg.WithMinimumVersion("2.9.0").AppendMapItem(cfg, "honor_timestamps", honor && !cg.spec.OverrideHonorTimestamps)
}
func addHonorLabels(cfg *config.ScrapeConfig, honorLabels bool) {
	//TODO:
	//if cg.spec.OverrideHonorLabels {
	//	honorLabels = false
	//}
	cfg.HonorLabels = honorLabels
}

var (
	invalidLabelCharRE = regexp.MustCompile(`[^a-zA-Z0-9_]`)
)

func sanitizeLabelName(name string) model.LabelName {
	return model.LabelName(invalidLabelCharRE.ReplaceAllString(name, "_"))
}

func getNamespacesFromNamespaceSelector(nsel v1.NamespaceSelector, namespace string) []string {
	// TODO:
	//if cg.spec.IgnoreNamespaceSelectors {
	//	return []string{namespace}
	//} else
	if nsel.Any {
		return []string{}
	} else if len(nsel.MatchNames) == 0 {
		return []string{namespace}
	}
	return nsel.MatchNames
}
