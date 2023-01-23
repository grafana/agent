package kubernetes_crds

// SEE https://github.com/prometheus-operator/prometheus-operator/blob/main/pkg/prometheus/promcfg.go

import (
	"context"
	"fmt"
	"net/url"
	"regexp"
	"sort"
	"strings"

	"github.com/alecthomas/units"
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

func (cg *configGenerator) generatePodMonitorConfig(m *v1.PodMonitor, ep v1.PodMetricsEndpoint, i int, shards int) *config.ScrapeConfig {
	c := config.DefaultScrapeConfig
	cfg := &c
	cfg.ScrapeInterval = config.DefaultGlobalConfig.ScrapeInterval
	cfg.ScrapeTimeout = config.DefaultGlobalConfig.ScrapeTimeout
	cfg.JobName = fmt.Sprintf("podMonitor/%s/%s/%d", m.Namespace, m.Name, i)
	cg.addHonorLabels(cfg, ep.HonorLabels)
	cg.addHonorTimestamps(cfg, ep.HonorTimestamps)

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
		cg.addSafeTLStoYaml(cfg, m.Namespace, ep.TLSConfig.SafeTLSConfig)
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

	cg.addOAuth2(cfg, ep.OAuth2, m.Namespace)
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
			Replacement: ep.Port,
			TargetLabel: ep.TargetPort.String(), //nolint:staticcheck // Ignore SA1019 this field is marked as deprecated.
		})
	}

	labeler := namespacelabeler.New(cg.config.EnforcedNamespaceLabel, cg.config.ExcludedFromEnforcement.Convert(), false)
	relabels.addFromV1(labeler.GetRelabelingConfigs(m.TypeMeta, m.ObjectMeta, ep.RelabelConfigs)...)
	if shards > 0 {
		relabels.generateAddressShardingRelabelingRules(shards)
	}

	cfg.RelabelConfigs = relabels.configs

	cg.addLimits(cfg, limitFuncs[sampleLimitKey], m.Spec.SampleLimit, cg.config.EnforcedSampleLimit)
	cg.addLimits(cfg, limitFuncs[targetLimitKey], m.Spec.TargetLimit, cg.config.EnforcedTargetLimit)
	cg.addLimits(cfg, limitFuncs[labelLimitKey], m.Spec.LabelLimit, cg.config.EnforcedLabelLimit)
	cg.addLimits(cfg, limitFuncs[labelNameLengthLimitKey], m.Spec.LabelNameLengthLimit, cg.config.EnforcedLabelNameLengthLimit)
	cg.addLimits(cfg, limitFuncs[labelValueLengthLimitKey], m.Spec.LabelValueLengthLimit, cg.config.EnforcedLabelValueLengthLimit)
	// TODO: body size is a parsed byte thing.
	//if cg.config.EnforcedBodySizeLimit != "" {
	// 	cfg = cg.WithMinimumVersion("2.28.0").AppendMapItem(cfg, "body_size_limit", cg.spec.EnforcedBodySizeLimit)
	//}

	metricRelabels := relabeler{}
	metricRelabels.addFromV1(labeler.GetRelabelingConfigs(m.TypeMeta, m.ObjectMeta, ep.MetricRelabelConfigs)...)
	cfg.MetricRelabelConfigs = metricRelabels.configs

	return cfg
}

const (
	sampleLimitKey           = "sampleLimit"
	targetLimitKey           = "targetLimit"
	labelLimitKey            = "labelLimit"
	labelNameLengthLimitKey  = "labelNameLengthLimit"
	labelValueLengthLimitKey = "labelValueLengthLimit"
	bodySizeLimitKey         = "bodySizeLimit"
)

type limitSetterFunc func(*config.ScrapeConfig, uint)

var limitFuncs = map[string]limitSetterFunc{
	sampleLimitKey:           func(cfg *config.ScrapeConfig, limit uint) { cfg.SampleLimit = limit },
	targetLimitKey:           func(cfg *config.ScrapeConfig, limit uint) { cfg.TargetLimit = limit },
	labelLimitKey:            func(cfg *config.ScrapeConfig, limit uint) { cfg.LabelLimit = limit },
	labelNameLengthLimitKey:  func(cfg *config.ScrapeConfig, limit uint) { cfg.LabelNameLengthLimit = limit },
	labelValueLengthLimitKey: func(cfg *config.ScrapeConfig, limit uint) { cfg.LabelValueLengthLimit = limit },
	bodySizeLimitKey:         func(cfg *config.ScrapeConfig, limit uint) { cfg.BodySizeLimit = units.Base2Bytes(limit) },
}

func (cg *configGenerator) addLimits(cfg *config.ScrapeConfig, f limitSetterFunc, userLimit uint64, enforcedLimit *uint64) {
	if userLimit == 0 && enforcedLimit == nil {
		return
	}
	limit := userLimit
	if enforcedLimit != nil {
		if *enforcedLimit > 0 && userLimit > 0 && userLimit > *enforcedLimit {
			limit = *enforcedLimit
		}
	}
	f(cfg, uint(limit))
}

func (cg *configGenerator) addSafeTLStoYaml(cfg *config.ScrapeConfig, namespace string, tls v1.SafeTLSConfig) {
	cfg.HTTPClientConfig.TLSConfig.InsecureSkipVerify = tls.InsecureSkipVerify
	var err error
	if tls.CA.Secret != nil {
		cfg.HTTPClientConfig.TLSConfig.CAFile, err = cg.secrets.StoreSecretData(context.Background(), namespace, tls.CA.Secret.Name, tls.CA.Secret.Key)
		if err != nil {
			// log error
		}
	} else if tls.CA.ConfigMap != nil {
		cfg.HTTPClientConfig.TLSConfig.CAFile, err = cg.secrets.StoreConfigMapData(context.Background(), namespace, tls.CA.ConfigMap.Name, tls.CA.ConfigMap.Key)
		if err != nil {
			// log error
		}
	}
	if tls.Cert.Secret != nil {
		cfg.HTTPClientConfig.TLSConfig.CertFile, err = cg.secrets.StoreSecretData(context.Background(), namespace, tls.Cert.Secret.Name, tls.Cert.Secret.Key)
		if err != nil {
			// log error
		}
	} else if tls.Cert.ConfigMap != nil {
		cfg.HTTPClientConfig.TLSConfig.CertFile, err = cg.secrets.StoreConfigMapData(context.Background(), namespace, tls.Cert.ConfigMap.Name, tls.Cert.ConfigMap.Key)
		if err != nil {
			// log error
		}
	}
	if tls.KeySecret != nil {
		cfg.HTTPClientConfig.TLSConfig.KeyFile, err = cg.secrets.StoreSecretData(context.Background(), namespace, tls.KeySecret.Name, tls.KeySecret.Key)
		if err != nil {
			// log error
		}
	}
	if tls.ServerName != "" {
		cfg.HTTPClientConfig.TLSConfig.ServerName = tls.ServerName
	}
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

// addHonorTimestamps adds the honor_timestamps field into scrape configurations.
// honor_timestamps is false only when the user specified it or when the global
// override applies.
func (cg *configGenerator) addHonorTimestamps(cfg *config.ScrapeConfig, userHonorTimestamps *bool) {
	if userHonorTimestamps != nil && *userHonorTimestamps {
		cfg.HonorTimestamps = true
	} else if cg.config.OverrideHonorTimestamps {
		cfg.HonorTimestamps = true
	}
}
func (cg *configGenerator) addHonorLabels(cfg *config.ScrapeConfig, honorLabels bool) {
	if cg.config.OverrideHonorLabels {
		cfg.HonorLabels = false
	}
	cfg.HonorLabels = honorLabels
}

var (
	invalidLabelCharRE = regexp.MustCompile(`[^a-zA-Z0-9_]`)
)

func sanitizeLabelName(name string) model.LabelName {
	return model.LabelName(invalidLabelCharRE.ReplaceAllString(name, "_"))
}

func (cg *configGenerator) getNamespacesFromNamespaceSelector(nsel v1.NamespaceSelector, namespace string) []string {
	if cg.config.IgnoreNamespaceSelectors {
		return []string{namespace}
	} else if nsel.Any {
		return []string{}
	} else if len(nsel.MatchNames) == 0 {
		return []string{namespace}
	}
	return nsel.MatchNames
}

func (cg *configGenerator) addOAuth2(cfg *config.ScrapeConfig, oauth2 *v1.OAuth2, ns string) {
	if oauth2 == nil {
		return
	}
	if oauth2.ClientID.Secret != nil {
		s := oauth2.ClientID.Secret
		cfg.HTTPClientConfig.OAuth2.ClientID = string(cg.getSecretData(ns, s.Name, s.Key))
	} else if oauth2.ClientID.ConfigMap != nil {
		cm := oauth2.ClientID.ConfigMap
		cfg.HTTPClientConfig.OAuth2.ClientID = cg.getConfigMapData(ns, cm.Name, cm.Key)
	}
	cfg.HTTPClientConfig.OAuth2.ClientSecret = cg.getSecretData(ns, oauth2.ClientSecret.Name, oauth2.ClientSecret.Key)
	cfg.HTTPClientConfig.OAuth2.TokenURL = oauth2.TokenURL
	if len(oauth2.Scopes) > 0 {
		cfg.HTTPClientConfig.OAuth2.Scopes = oauth2.Scopes
	}

	if len(oauth2.EndpointParams) > 0 {
		cfg.HTTPClientConfig.OAuth2.EndpointParams = oauth2.EndpointParams
	}
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

func (r *relabeler) generateAddressShardingRelabelingRules(shards int) {
	r.generateAddressShardingRelabelingRulesWithSourceLabel(shards, "__address__")
}

func (r *relabeler) generateAddressShardingRelabelingRulesForProbes(shards int) {
	r.generateAddressShardingRelabelingRulesWithSourceLabel(shards, "__param_target")
}

func (r *relabeler) generateAddressShardingRelabelingRulesWithSourceLabel(shards int, shardLabel string) {

	r.Add(
		&relabel.Config{
			SourceLabels: model.LabelNames{model.LabelName(shardLabel)},
			TargetLabel:  "__tmp_hash",
			Modulus:      uint64(shards),
			Action:       relabel.HashMod,
		},
		&relabel.Config{
			SourceLabels: model.LabelNames{model.LabelName("__tmp_hash")},
			Regex:        relabel.MustNewRegexp("$(SHARD)"),
			Action:       relabel.Keep,
		})
}
