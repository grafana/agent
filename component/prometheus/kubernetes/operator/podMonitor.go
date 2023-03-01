package operator

// SEE https://github.com/prometheus-operator/prometheus-operator/blob/main/pkg/prometheus/promcfg.go

import (
	"fmt"
	"net/url"
	"sort"
	"strings"

	"github.com/go-kit/log/level"
	"github.com/grafana/agent/pkg/util/k8sfs"
	v1 "github.com/prometheus-operator/prometheus-operator/pkg/apis/monitoring/v1"
	namespacelabeler "github.com/prometheus-operator/prometheus-operator/pkg/namespace-labeler"
	commonConfig "github.com/prometheus/common/config"
	"github.com/prometheus/common/model"
	"github.com/prometheus/prometheus/config"
	promk8s "github.com/prometheus/prometheus/discovery/kubernetes"
	"github.com/prometheus/prometheus/model/relabel"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func (cg *configGenerator) generatePodMonitorConfig(m *v1.PodMonitor, ep v1.PodMetricsEndpoint, i int) *config.ScrapeConfig {
	c := config.DefaultScrapeConfig
	cfg := &c
	cfg.ScrapeInterval = config.DefaultGlobalConfig.ScrapeInterval
	cfg.ScrapeTimeout = config.DefaultGlobalConfig.ScrapeTimeout
	cfg.JobName = fmt.Sprintf("podMonitor/%s/%s/%d", m.Namespace, m.Name, i)
	cfg.HonorLabels = ep.HonorLabels
	if ep.HonorTimestamps != nil {
		cfg.HonorTimestamps = *ep.HonorTimestamps
	}

	cfg.ServiceDiscoveryConfigs = append(cfg.ServiceDiscoveryConfigs, cg.generateK8SSDConfig(m.Spec.NamespaceSelector, m.Namespace, promk8s.RolePod, m.Spec.AttachMetadata))

	if ep.Interval != "" {
		var err error
		cfg.ScrapeInterval, err = model.ParseDuration(string(ep.Interval))
		if err != nil {
			level.Error(cg.logger).Log("msg", "failed to parse Interval from podMonitor", "err", err)
		}
	}
	if ep.ScrapeTimeout != "" {
		cfg.ScrapeInterval, _ = model.ParseDuration(string(ep.ScrapeTimeout))
	}
	if ep.Path != "" {
		cfg.MetricsPath = ep.Path
	}
	if ep.ProxyURL != nil {
		if u, err := url.Parse(*ep.ProxyURL); err != nil {
			level.Error(cg.logger).Log("msg", "failed to parse ProxyURL from podMonitor", "err", err)
		} else {
			cfg.HTTPClientConfig.ProxyURL = commonConfig.URL{URL: u}
		}
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
	if ep.TLSConfig != nil {
		cfg.HTTPClientConfig.TLSConfig = cg.generateSafeTLS(m.Namespace, ep.TLSConfig.SafeTLSConfig)
	}
	if ep.BearerTokenSecret.Name != "" {
		bts := ep.BearerTokenSecret
		cfg.HTTPClientConfig.BearerTokenFile = k8sfs.SecretFilename(m.Namespace, bts.Name, bts.Key)
	}
	if ep.BasicAuth != nil {
		uname, err := cg.secretfs.ReadSecret(m.Namespace, ep.BasicAuth.Username.Name, ep.BasicAuth.Username.Key)
		if err != nil {
			level.Error(cg.logger).Log("msg", "failed to fetch basic auth username", "err", err)
		}
		cfg.HTTPClientConfig.BasicAuth = &commonConfig.BasicAuth{
			Username:     uname,
			PasswordFile: k8sfs.SecretFilename(m.Namespace, ep.BasicAuth.Password.Name, ep.BasicAuth.Password.Key),
		}
	}
	cfg.HTTPClientConfig.OAuth2 = cg.generateOAuth2(ep.OAuth2, m.Namespace)
	cfg.HTTPClientConfig.Authorization = cg.generateSafeAuthorization(ep.Authorization, m.Namespace)

	relabels := cg.initRelabelings(cfg)
	if ep.FilterRunning == nil || *ep.FilterRunning {
		relabels.Add(&relabel.Config{
			SourceLabels: model.LabelNames{"__meta_kubernetes_pod_phase"},
			Action:       "drop",
			Regex:        parseRegexp("(Failed|Succeeded)", cg.logger),
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
		relabels.Add(&relabel.Config{
			SourceLabels: model.LabelNames{"__meta_kubernetes_pod_label_" + sanitizeLabelName(k), "__meta_kubernetes_pod_labelpresent_" + sanitizeLabelName(k)},
			Action:       "keep",
			Regex:        parseRegexp(fmt.Sprintf("(%s);true", m.Spec.Selector.MatchLabels[k]), cg.logger),
		})
	}

	// Set based label matching. We have to map the valid relations
	// `In`, `NotIn`, `Exists`, and `DoesNotExist`, into relabeling rules.
	for _, exp := range m.Spec.Selector.MatchExpressions {
		switch exp.Operator {
		case metav1.LabelSelectorOpIn:
			relabels.Add(&relabel.Config{
				SourceLabels: model.LabelNames{"__meta_kubernetes_pod_label_" + sanitizeLabelName(exp.Key), "__meta_kubernetes_pod_labelpresent_" + sanitizeLabelName(exp.Key)},
				Action:       "keep",
				Regex:        parseRegexp(fmt.Sprintf("(%s);true", strings.Join(exp.Values, "|")), cg.logger),
			})
		case metav1.LabelSelectorOpNotIn:
			relabels.Add(&relabel.Config{
				SourceLabels: model.LabelNames{"__meta_kubernetes_pod_label_" + sanitizeLabelName(exp.Key), "__meta_kubernetes_pod_labelpresent_" + sanitizeLabelName(exp.Key)},
				Action:       "drop",
				Regex:        parseRegexp(fmt.Sprintf("(%s);true", strings.Join(exp.Values, "|")), cg.logger),
			})
		case metav1.LabelSelectorOpExists:
			relabels.Add(&relabel.Config{
				SourceLabels: model.LabelNames{"__meta_kubernetes_pod_labelpresent_" + sanitizeLabelName(exp.Key)},
				Action:       "keep",
				Regex:        parseRegexp("true", cg.logger),
			})
		case metav1.LabelSelectorOpDoesNotExist:
			relabels.Add(&relabel.Config{
				SourceLabels: model.LabelNames{"__meta_kubernetes_pod_labelpresent_" + sanitizeLabelName(exp.Key)},
				Action:       "drop",
				Regex:        parseRegexp("true", cg.logger),
			})
		}
	}

	// Filter targets based on correct port for the endpoint.
	if ep.Port != "" {
		relabels.Add(&relabel.Config{
			SourceLabels: model.LabelNames{"__meta_kubernetes_pod_container_port_name"},
			Action:       "keep",
			Regex:        parseRegexp(ep.Port, cg.logger),
		})
	} else if ep.TargetPort != nil { //nolint:staticcheck // Ignore SA1019 this field is marked as deprecated.
		//nolint:staticcheck // Ignore SA1019 this field is marked as deprecated.
		if ep.TargetPort.StrVal != "" {
			relabels.Add(&relabel.Config{
				SourceLabels: model.LabelNames{"__meta_kubernetes_pod_container_port_name"},
				Action:       "keep",
				Regex:        parseRegexp(ep.TargetPort.String(), cg.logger),
			})
		}
	} else if ep.TargetPort.IntVal != 0 { //nolint:staticcheck // Ignore SA1019 this field is marked as deprecated.
		relabels.Add(&relabel.Config{
			SourceLabels: model.LabelNames{"__meta_kubernetes_pod_container_port_number"},
			Action:       "keep",
			Regex:        parseRegexp(ep.TargetPort.String(), cg.logger),
		})
	}

	// Relabel namespace and pod and service labels into proper labels.
	relabels.Add(&relabel.Config{
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
		relabels.Add(&relabel.Config{
			SourceLabels: model.LabelNames{"__meta_kubernetes_pod_label_" + sanitizeLabelName(l)},
			Replacement:  "${1}",
			Regex:        parseRegexp("(.+)", cg.logger),
			TargetLabel:  string(sanitizeLabelName(l)),
		})
	}

	// By default, generate a safe job name from the PodMonitor. We also keep
	// this around if a jobLabel is set in case the targets don't actually have a
	// value for it. A single pod may potentially have multiple metrics
	// endpoints, therefore the endpoints labels is filled with the ports name or
	// as a fallback the port number.

	relabels.Add(&relabel.Config{
		Replacement: fmt.Sprintf("%s/%s", m.GetNamespace(), m.GetName()),
		TargetLabel: "job",
	})
	if m.Spec.JobLabel != "" {
		relabels.Add(&relabel.Config{
			Replacement:  "${1}",
			TargetLabel:  "job",
			Regex:        parseRegexp("(.+)", cg.logger),
			SourceLabels: model.LabelNames{"__meta_kubernetes_pod_label_" + sanitizeLabelName(m.Spec.JobLabel)},
		})
	}

	if ep.Port != "" {
		relabels.Add(&relabel.Config{
			Replacement: ep.Port,
			TargetLabel: "endpoint",
		})
	} else if ep.TargetPort != nil && ep.TargetPort.String() != "" { //nolint:staticcheck // Ignore SA1019 this field is marked as deprecated.
		relabels.Add(&relabel.Config{
			TargetLabel: "endpoint",
			Replacement: ep.TargetPort.String(), //nolint:staticcheck // Ignore SA1019 this field is marked as deprecated.
		})
	}

	labeler := namespacelabeler.New("", nil, false)
	relabels.addFromV1(labeler.GetRelabelingConfigs(m.TypeMeta, m.ObjectMeta, ep.RelabelConfigs)...)

	cfg.RelabelConfigs = relabels.configs

	metricRelabels := relabeler{}
	metricRelabels.addFromV1(labeler.GetRelabelingConfigs(m.TypeMeta, m.ObjectMeta, ep.MetricRelabelConfigs)...)
	cfg.MetricRelabelConfigs = metricRelabels.configs

	// TODO: limits from spec

	return cfg
}
