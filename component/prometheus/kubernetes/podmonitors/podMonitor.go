package podmonitors

// SEE https://github.com/prometheus-operator/prometheus-operator/blob/main/pkg/prometheus/promcfg.go

import (
	"fmt"
	"net/url"
	"sort"
	"strings"

	"github.com/pkg/errors"

	v1 "github.com/prometheus-operator/prometheus-operator/pkg/apis/monitoring/v1"
	namespacelabeler "github.com/prometheus-operator/prometheus-operator/pkg/namespace-labeler"
	commonConfig "github.com/prometheus/common/config"
	"github.com/prometheus/common/model"
	"github.com/prometheus/prometheus/config"
	promk8s "github.com/prometheus/prometheus/discovery/kubernetes"
	"github.com/prometheus/prometheus/model/relabel"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

var (
	regexFilterRunning = relabel.MustNewRegexp("(Failed|Succeeded)")
	regexTrue          = relabel.MustNewRegexp("true")
	regexAnything      = relabel.MustNewRegexp("(.+)")
)

func (cg *configGenerator) generatePodMonitorConfig(m *v1.PodMonitor, ep v1.PodMetricsEndpoint, i int) (cfg *config.ScrapeConfig, err error) {
	c := config.DefaultScrapeConfig
	cfg = &c
	cfg.ScrapeInterval = config.DefaultGlobalConfig.ScrapeInterval
	cfg.ScrapeTimeout = config.DefaultGlobalConfig.ScrapeTimeout
	cfg.JobName = fmt.Sprintf("podMonitor/%s/%s/%d", m.Namespace, m.Name, i)
	cfg.HonorLabels = ep.HonorLabels
	if ep.HonorTimestamps != nil {
		cfg.HonorTimestamps = *ep.HonorTimestamps
	}

	cfg.ServiceDiscoveryConfigs = append(cfg.ServiceDiscoveryConfigs, cg.generateK8SSDConfig(m.Spec.NamespaceSelector, m.Namespace, promk8s.RolePod, m.Spec.AttachMetadata))

	if ep.Interval != "" {
		if cfg.ScrapeInterval, err = model.ParseDuration(string(ep.Interval)); err != nil {
			return nil, errors.Wrap(err, "parsing interval from podMonitor")
		}
	}
	if ep.ScrapeTimeout != "" {
		if cfg.ScrapeTimeout, err = model.ParseDuration(string(ep.ScrapeTimeout)); err != nil {
			return nil, errors.Wrap(err, "parsing timeout from podMonitor")
		}
	}
	if ep.Path != "" {
		cfg.MetricsPath = ep.Path
	}
	if ep.ProxyURL != nil {
		if u, err := url.Parse(*ep.ProxyURL); err != nil {
			return nil, errors.Wrap(err, "parsing ProxyURL from podMonitor")
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
		if cfg.HTTPClientConfig.TLSConfig, err = cg.GenerateSafeTLS(m.Namespace, ep.TLSConfig.SafeTLSConfig); err != nil {
			return nil, err
		}
	}
	if ep.BearerTokenSecret.Name != "" {
		return nil, fmt.Errorf("bearer tokens in podmonitors not supported yet")
	}
	if ep.BasicAuth != nil {
		return nil, fmt.Errorf("basic auth in podmonitors not supported yet")
	}
	if cfg.HTTPClientConfig.OAuth2, err = cg.GenerateOAuth2(ep.OAuth2, m.Namespace); err != nil {
		return nil, err
	}
	if cfg.HTTPClientConfig.Authorization, err = cg.GenerateSafeAuthorization(ep.Authorization, m.Namespace); err != nil {
		return nil, err
	}

	relabels := cg.initRelabelings()
	if ep.FilterRunning == nil || *ep.FilterRunning {
		relabels.Add(&relabel.Config{
			SourceLabels: model.LabelNames{"__meta_kubernetes_pod_phase"},
			Action:       "drop",
			Regex:        regexFilterRunning,
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
		regex, err := relabel.NewRegexp(fmt.Sprintf("(%s);true", m.Spec.Selector.MatchLabels[k]))
		if err != nil {
			return nil, errors.Wrap(err, "parsing MatchLabels regex")
		}
		relabels.Add(&relabel.Config{
			SourceLabels: model.LabelNames{"__meta_kubernetes_pod_label_" + sanitizeLabelName(k), "__meta_kubernetes_pod_labelpresent_" + sanitizeLabelName(k)},
			Action:       "keep",
			Regex:        regex,
		})
	}

	// Set based label matching. We have to map the valid relations
	// `In`, `NotIn`, `Exists`, and `DoesNotExist`, into relabeling rules.
	for _, exp := range m.Spec.Selector.MatchExpressions {
		switch exp.Operator {
		case metav1.LabelSelectorOpIn:
			regex, err := relabel.NewRegexp(fmt.Sprintf("(%s);true", strings.Join(exp.Values, "|")))
			if err != nil {
				return nil, errors.Wrap(err, "parsing MatchExpressions regex")
			}
			relabels.Add(&relabel.Config{
				SourceLabels: model.LabelNames{"__meta_kubernetes_pod_label_" + sanitizeLabelName(exp.Key), "__meta_kubernetes_pod_labelpresent_" + sanitizeLabelName(exp.Key)},
				Action:       "keep",
				Regex:        regex,
			})
		case metav1.LabelSelectorOpNotIn:
			regex, err := relabel.NewRegexp(fmt.Sprintf("(%s);true", strings.Join(exp.Values, "|")))
			if err != nil {
				return nil, errors.Wrap(err, "parsing MatchExpressions regex")
			}
			relabels.Add(&relabel.Config{
				SourceLabels: model.LabelNames{"__meta_kubernetes_pod_label_" + sanitizeLabelName(exp.Key), "__meta_kubernetes_pod_labelpresent_" + sanitizeLabelName(exp.Key)},
				Action:       "drop",
				Regex:        regex,
			})
		case metav1.LabelSelectorOpExists:
			relabels.Add(&relabel.Config{
				SourceLabels: model.LabelNames{"__meta_kubernetes_pod_labelpresent_" + sanitizeLabelName(exp.Key)},
				Action:       "keep",
				Regex:        regexTrue,
			})
		case metav1.LabelSelectorOpDoesNotExist:
			relabels.Add(&relabel.Config{
				SourceLabels: model.LabelNames{"__meta_kubernetes_pod_labelpresent_" + sanitizeLabelName(exp.Key)},
				Action:       "drop",
				Regex:        regexTrue,
			})
		}
	}

	// Filter targets based on correct port for the endpoint.
	if ep.Port != "" {
		regex, err := relabel.NewRegexp(ep.Port)
		if err != nil {
			return nil, errors.Wrap(err, "parsing Port as regex")
		}
		relabels.Add(&relabel.Config{
			SourceLabels: model.LabelNames{"__meta_kubernetes_pod_container_port_name"},
			Action:       "keep",
			Regex:        regex,
		})
	} else if ep.TargetPort != nil { //nolint:staticcheck // Ignore SA1019 this field is marked as deprecated.
		//nolint:staticcheck // Ignore SA1019 this field is marked as deprecated.
		regex, err := relabel.NewRegexp(ep.TargetPort.String())
		if err != nil {
			return nil, errors.Wrap(err, "parsing TargetPort as regex")
		}
		if ep.TargetPort.StrVal != "" {
			relabels.Add(&relabel.Config{
				SourceLabels: model.LabelNames{"__meta_kubernetes_pod_container_port_name"},
				Action:       "keep",
				Regex:        regex,
			})
		}
	} else if ep.TargetPort.IntVal != 0 { //nolint:staticcheck // Ignore SA1019 this field is marked as deprecated.
		regex, err := relabel.NewRegexp(ep.TargetPort.String())
		if err != nil {
			return nil, errors.Wrap(err, "parsing TargetPort as regex")
		}
		relabels.Add(&relabel.Config{
			SourceLabels: model.LabelNames{"__meta_kubernetes_pod_container_port_number"},
			Action:       "keep",
			Regex:        regex,
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
			Regex:        regexAnything,
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
			Regex:        regexAnything,
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
	if err = relabels.addFromV1(labeler.GetRelabelingConfigs(m.TypeMeta, m.ObjectMeta, ep.RelabelConfigs)...); err != nil {
		return nil, errors.Wrap(err, "Parsing relabel configs")
	}

	cfg.RelabelConfigs = relabels.configs

	metricRelabels := relabeler{}
	if err = metricRelabels.addFromV1(labeler.GetRelabelingConfigs(m.TypeMeta, m.ObjectMeta, ep.MetricRelabelConfigs)...); err != nil {
		return nil, errors.Wrap(err, "Parsing metric relabel configs")
	}
	cfg.MetricRelabelConfigs = metricRelabels.configs

	cfg.SampleLimit = uint(m.Spec.SampleLimit)
	cfg.TargetLimit = uint(m.Spec.TargetLimit)
	cfg.LabelLimit = uint(m.Spec.LabelLimit)
	cfg.LabelNameLengthLimit = uint(m.Spec.LabelNameLengthLimit)
	cfg.LabelValueLengthLimit = uint(m.Spec.LabelValueLengthLimit)

	return cfg, nil
}
