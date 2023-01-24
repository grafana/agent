package kubernetes_crds

// SEE https://github.com/prometheus-operator/prometheus-operator/blob/main/pkg/prometheus/promcfg.go

import (
	"context"
	"fmt"
	"net/url"
	"regexp"
	"sort"
	"strings"

	v1 "github.com/prometheus-operator/prometheus-operator/pkg/apis/monitoring/v1"
	namespacelabeler "github.com/prometheus-operator/prometheus-operator/pkg/namespace-labeler"
	commonConfig "github.com/prometheus/common/config"
	"github.com/prometheus/common/model"
	"github.com/prometheus/prometheus/config"
	promk8s "github.com/prometheus/prometheus/discovery/kubernetes"
	"github.com/prometheus/prometheus/model/relabel"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type configGenerator struct {
	config  *Config
	secrets *secretManager
}

// the k8s sd config is mostly dependent on our local config for accessing the kubernetes cluster.
// if undefined it will default to an in-cluster config
func (cg *configGenerator) generateK8SSDConfig(
	namespaceSelector v1.NamespaceSelector,
	namespace string,
	role promk8s.Role,
	attachMetadata *v1.AttachMetadata,
) *promk8s.SDConfig {
	cfg := &promk8s.SDConfig{
		Role: role,
	}
	namespaces := cg.getNamespacesFromNamespaceSelector(namespaceSelector, namespace)
	if len(namespaces) != 0 {
		cfg.NamespaceDiscovery.Names = namespaces
	}
	if cg.config.KubeConfig != "" {
		cfg.KubeConfig = cg.config.KubeConfig
	}
	if cg.config.ApiServerConfig != nil {
		apiCfg := cg.config.ApiServerConfig
		cfg.APIServer = apiCfg.Host.Convert()

		if apiCfg.BasicAuth != nil {
			cfg.HTTPClientConfig.BasicAuth = apiCfg.BasicAuth.Convert()
		}

		if apiCfg.BearerToken != "" {
			cfg.HTTPClientConfig.BearerToken = commonConfig.Secret(apiCfg.BearerToken)
		}
		if apiCfg.BearerTokenFile != "" {
			cfg.HTTPClientConfig.BearerTokenFile = apiCfg.BearerTokenFile
		}
		if apiCfg.TLSConfig != nil {
			cfg.HTTPClientConfig.TLSConfig = *apiCfg.TLSConfig.Convert()
		}
		if apiCfg.Authorization != nil {
			if apiCfg.Authorization.Type == "" {
				apiCfg.Authorization.Type = "Bearer"
			}
			cfg.HTTPClientConfig.Authorization = apiCfg.Authorization.Convert()
		}
	}
	if attachMetadata != nil {
		cfg.AttachMetadata.Node = attachMetadata.Node
	}
	return cfg
}

func (cg *configGenerator) getSecretData(ns, name, key string) commonConfig.Secret {
	tok, err := cg.secrets.GetSecretData(context.Background(), ns, name, key)
	if err != nil {
		// TODO: log error or die
	} else {
		return commonConfig.Secret(tok)
	}
	return commonConfig.Secret("")
}

func (cg *configGenerator) getConfigMapData(ns, name, key string) string {
	tok, err := cg.secrets.GetConfigMapData(context.Background(), ns, name, key)
	if err != nil {
		// TODO: log error or die
	} else {
		return tok
	}
	return ""
}

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
	if ep.TLSConfig != nil {
		cfg.HTTPClientConfig.TLSConfig = cg.generateSafeTLS(m.Namespace, ep.TLSConfig.SafeTLSConfig)
	}

	if ep.BearerTokenSecret.Name != "" {
		bts := ep.BearerTokenSecret
		cfg.HTTPClientConfig.BearerToken = cg.getSecretData(m.Namespace, bts.Name, bts.Key)
	}

	if ep.BasicAuth != nil {
		cfg.HTTPClientConfig.BasicAuth = &commonConfig.BasicAuth{
			Username: string(cg.getSecretData(m.Namespace, ep.BasicAuth.Username.Name, ep.BasicAuth.Username.Key)),
			Password: cg.getSecretData(m.Namespace, ep.BasicAuth.Password.Name, ep.BasicAuth.Password.Key),
		}
	}

	cfg.HTTPClientConfig.OAuth2 = cg.generateOAuth2(ep.OAuth2, m.Namespace)
	cg.addSafeAuthorization(cfg, ep.Authorization, m.Namespace)

	relabels := cg.initRelabelings(cfg)
	if ep.FilterRunning == nil || *ep.FilterRunning {
		relabels.Add(&relabel.Config{
			SourceLabels: model.LabelNames{"__meta_kubernetes_pod_phase"},
			Action:       "drop",
			Regex:        relabel.MustNewRegexp("(Failed|Succeeded)"),
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
			Regex:        relabel.MustNewRegexp(fmt.Sprintf("(%s);true", m.Spec.Selector.MatchLabels[k])),
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
				Regex:        relabel.MustNewRegexp(fmt.Sprintf("(%s);true", strings.Join(exp.Values, "|"))),
			})
		case metav1.LabelSelectorOpNotIn:
			relabels.Add(&relabel.Config{
				SourceLabels: model.LabelNames{"__meta_kubernetes_pod_label_" + sanitizeLabelName(exp.Key), "__meta_kubernetes_pod_labelpresent_" + sanitizeLabelName(exp.Key)},
				Action:       "drop",
				Regex:        relabel.MustNewRegexp(fmt.Sprintf("(%s);true", strings.Join(exp.Values, "|"))),
			})
		case metav1.LabelSelectorOpExists:
			relabels.Add(&relabel.Config{
				SourceLabels: model.LabelNames{"__meta_kubernetes_pod_labelpresent_" + sanitizeLabelName(exp.Key)},
				Action:       "keep",
				Regex:        relabel.MustNewRegexp("true"),
			})
		case metav1.LabelSelectorOpDoesNotExist:
			relabels.Add(&relabel.Config{
				SourceLabels: model.LabelNames{"__meta_kubernetes_pod_labelpresent_" + sanitizeLabelName(exp.Key)},
				Action:       "drop",
				Regex:        relabel.MustNewRegexp("true"),
			})
		}
	}

	// Filter targets based on correct port for the endpoint.
	if ep.Port != "" {
		relabels.Add(&relabel.Config{
			SourceLabels: model.LabelNames{"__meta_kubernetes_pod_container_port_name"},
			Action:       "keep",
			Regex:        relabel.MustNewRegexp(ep.Port),
		})
	} else if ep.TargetPort != nil { //nolint:staticcheck // Ignore SA1019 this field is marked as deprecated.
		//nolint:staticcheck // Ignore SA1019 this field is marked as deprecated.
		if ep.TargetPort.StrVal != "" {
			relabels.Add(&relabel.Config{
				SourceLabels: model.LabelNames{"__meta_kubernetes_pod_container_port_name"},
				Action:       "keep",
				Regex:        relabel.MustNewRegexp(ep.TargetPort.String()),
			})
		}
	} else if ep.TargetPort.IntVal != 0 { //nolint:staticcheck // Ignore SA1019 this field is marked as deprecated.
		relabels.Add(&relabel.Config{
			SourceLabels: model.LabelNames{"__meta_kubernetes_pod_container_port_number"},
			Action:       "keep",
			Regex:        relabel.MustNewRegexp(ep.TargetPort.String()),
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
			Regex:        relabel.MustNewRegexp("(.+)"),
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
			Regex:        relabel.MustNewRegexp("(.+)"),
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

	return cfg
}

func (cg *configGenerator) generateSafeTLS(namespace string, tls v1.SafeTLSConfig) commonConfig.TLSConfig {
	tc := commonConfig.TLSConfig{}
	tc.InsecureSkipVerify = tls.InsecureSkipVerify
	ctx := context.Background()
	var err error
	if tls.CA.Secret != nil {
		tc.CAFile, err = cg.secrets.StoreSecretData(ctx, namespace, tls.CA.Secret.Name, tls.CA.Secret.Key)
		if err != nil {
			// log error
		}
	} else if tls.CA.ConfigMap != nil {
		tc.CAFile, err = cg.secrets.StoreConfigMapData(ctx, namespace, tls.CA.ConfigMap.Name, tls.CA.ConfigMap.Key)
		if err != nil {
			// log error
		}
	}
	if tls.Cert.Secret != nil {
		tc.CertFile, err = cg.secrets.StoreSecretData(ctx, namespace, tls.Cert.Secret.Name, tls.Cert.Secret.Key)
		if err != nil {
			// log error
		}
	} else if tls.Cert.ConfigMap != nil {
		tc.CertFile, err = cg.secrets.StoreConfigMapData(ctx, namespace, tls.Cert.ConfigMap.Name, tls.Cert.ConfigMap.Key)
		if err != nil {
			// log error
		}
	}
	if tls.KeySecret != nil {
		tc.KeyFile, err = cg.secrets.StoreSecretData(ctx, namespace, tls.KeySecret.Name, tls.KeySecret.Key)
		if err != nil {
			// log error
		}
	}
	if tls.ServerName != "" {
		tc.ServerName = tls.ServerName
	}
	return tc
}

type relabeler struct {
	configs []*relabel.Config
}

func (r *relabeler) Add(cfgs ...*relabel.Config) {
	for _, cfg := range cfgs {
		// set defaults from prom defaults.
		if cfg.Action == "" {
			cfg.Action = relabel.DefaultRelabelConfig.Action
		}
		if cfg.Separator == "" {
			cfg.Separator = relabel.DefaultRelabelConfig.Separator
		}
		if cfg.Regex.Regexp == nil {
			cfg.Regex = relabel.DefaultRelabelConfig.Regex
		}
		if cfg.Replacement == "" {
			cfg.Replacement = relabel.DefaultRelabelConfig.Replacement
		}
		r.configs = append(r.configs, cfg)
	}
}

// addFromMonitoring converts from an externally generated monitoringv1 RelabelConfig
func (r *relabeler) addFromV1(cfgs ...*v1.RelabelConfig) {
	for _, c := range cfgs {
		cfg := &relabel.Config{}
		for _, l := range c.SourceLabels {
			cfg.SourceLabels = append(cfg.SourceLabels, model.LabelName(l))
		}
		if c.Separator != "" {
			cfg.Separator = c.Separator
		}
		if c.TargetLabel != "" {
			cfg.TargetLabel = c.TargetLabel
		}
		if c.Regex != "" {
			if r, err := relabel.NewRegexp(c.Regex); err != nil {
				cfg.Regex = r
			} else {
				// TODO: LOG ERROR?
			}
		}
		if c.Modulus != 0 {
			cfg.Modulus = c.Modulus
		}
		if c.Replacement != "" {
			cfg.Replacement = c.Replacement
		}
		if c.Action != "" {
			cfg.Action = relabel.Action(c.Action)
		}
		r.configs = append(r.configs, cfg)
	}
}

func (cg *configGenerator) initRelabelings(cfg *config.ScrapeConfig) relabeler {
	r := relabeler{}
	// Relabel prometheus job name into a meta label
	r.Add(&relabel.Config{
		SourceLabels: model.LabelNames{"job"},
		TargetLabel:  "__tmp_prometheus_job_name",
	})
	return r
}

var (
	invalidLabelCharRE = regexp.MustCompile(`[^a-zA-Z0-9_]`)
)

func sanitizeLabelName(name string) model.LabelName {
	return model.LabelName(invalidLabelCharRE.ReplaceAllString(name, "_"))
}

func (cg *configGenerator) getNamespacesFromNamespaceSelector(nsel v1.NamespaceSelector, namespace string) []string {
	if nsel.Any {
		return []string{}
	} else if len(nsel.MatchNames) == 0 {
		return []string{namespace}
	}
	return nsel.MatchNames
}

func (cg *configGenerator) generateOAuth2(oauth2 *v1.OAuth2, ns string) *commonConfig.OAuth2 {
	if oauth2 == nil {
		return nil
	}
	oa2 := &commonConfig.OAuth2{}
	if oauth2.ClientID.Secret != nil {
		s := oauth2.ClientID.Secret
		oa2.ClientID = string(cg.getSecretData(ns, s.Name, s.Key))
	} else if oauth2.ClientID.ConfigMap != nil {
		cm := oauth2.ClientID.ConfigMap
		oa2.ClientID = cg.getConfigMapData(ns, cm.Name, cm.Key)
	}
	oa2.ClientSecret = cg.getSecretData(ns, oauth2.ClientSecret.Name, oauth2.ClientSecret.Key)
	oa2.TokenURL = oauth2.TokenURL
	if len(oauth2.Scopes) > 0 {
		oa2.Scopes = oauth2.Scopes
	}
	if len(oauth2.EndpointParams) > 0 {
		oa2.EndpointParams = oauth2.EndpointParams
	}
	return oa2
}

func (cg *configGenerator) addSafeAuthorization(cfg *config.ScrapeConfig, auth *v1.SafeAuthorization, ns string) {
	if auth == nil {
		return
	}
	if auth.Type == "" {
		auth.Type = "Bearer"
	}
	cfg.HTTPClientConfig.Authorization.Type = strings.TrimSpace(auth.Type)

	if auth.Credentials != nil {
		cfg.HTTPClientConfig.Authorization.Credentials = cg.getSecretData(ns, auth.Credentials.Name, auth.Credentials.Key)
	}
}
