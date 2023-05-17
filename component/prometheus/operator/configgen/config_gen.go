package configgen

// SEE https://github.com/prometheus-operator/prometheus-operator/blob/aa8222d7e9b66e9293ed11c9291ea70173021029/pkg/prometheus/promcfg.go

import (
	"regexp"

	k8sConfig "github.com/grafana/agent/component/common/kubernetes"
	promopv1 "github.com/prometheus-operator/prometheus-operator/pkg/apis/monitoring/v1"
	commonConfig "github.com/prometheus/common/config"
	"github.com/prometheus/common/model"
	promk8s "github.com/prometheus/prometheus/discovery/kubernetes"
	"github.com/prometheus/prometheus/model/relabel"
)

type ConfigGenerator struct {
	Client  *k8sConfig.ClientArguments
	Secrets SecretFetcher
}

var (
	invalidLabelCharRE = regexp.MustCompile(`[^a-zA-Z0-9_]`)
)

// generateK8SSDConfig generates a kubernetes service discovery config based on the given namespace selector.
// The k8s sd config is mostly dependent on our local config for accessing the kubernetes cluster.
// If undefined it will default to an in-cluster config
func (cg *ConfigGenerator) generateK8SSDConfig(namespaceSelector promopv1.NamespaceSelector, namespace string, role promk8s.Role, attachMetadata *promopv1.AttachMetadata) *promk8s.SDConfig {
	cfg := &promk8s.SDConfig{
		Role: role,
	}
	namespaces := cg.getNamespacesFromNamespaceSelector(namespaceSelector, namespace)
	if len(namespaces) != 0 {
		cfg.NamespaceDiscovery.Names = namespaces
	}
	client := cg.Client
	if client.KubeConfig != "" {
		cfg.KubeConfig = client.KubeConfig
	}
	if client.APIServer.URL != nil {
		hCfg := client.HTTPClientConfig
		cfg.APIServer = client.APIServer.Convert()

		if hCfg.BasicAuth != nil {
			cfg.HTTPClientConfig.BasicAuth = hCfg.BasicAuth.Convert()
		}

		if hCfg.BearerToken != "" {
			cfg.HTTPClientConfig.BearerToken = commonConfig.Secret(hCfg.BearerToken)
		}
		if hCfg.BearerTokenFile != "" {
			cfg.HTTPClientConfig.BearerTokenFile = hCfg.BearerTokenFile
		}
		cfg.HTTPClientConfig.TLSConfig = *hCfg.TLSConfig.Convert()
		if hCfg.Authorization != nil {
			if hCfg.Authorization.Type == "" {
				hCfg.Authorization.Type = "Bearer"
			}
			cfg.HTTPClientConfig.Authorization = hCfg.Authorization.Convert()
		}
	}
	if attachMetadata != nil {
		cfg.AttachMetadata.Node = attachMetadata.Node
	}
	return cfg
}

func (cg *ConfigGenerator) generateSafeTLS(tls promopv1.SafeTLSConfig, namespace string) (commonConfig.TLSConfig, error) {
	tc := commonConfig.TLSConfig{}
	tc.InsecureSkipVerify = tls.InsecureSkipVerify
	var err error
	var value string
	if tls.CA.Secret != nil || tls.CA.ConfigMap != nil {
		tc.CA, err = cg.Secrets.SecretOrConfigMapValue(namespace, tls.CA)
		if err != nil {
			return tc, err
		}
	}
	if tls.Cert.Secret != nil || tls.Cert.ConfigMap != nil {
		tc.Cert, err = cg.Secrets.SecretOrConfigMapValue(namespace, tls.Cert)
		if err != nil {
			return tc, err
		}
	}
	if tls.KeySecret != nil {
		value, err = cg.Secrets.GetSecretValue(namespace, *tls.KeySecret)
		if err != nil {
			return tc, err
		}
		tc.Key = commonConfig.Secret(value)
	}
	if tls.ServerName != "" {
		tc.ServerName = tls.ServerName
	}
	return tc, nil
}

func (cg *ConfigGenerator) generateBasicAuth(auth promopv1.BasicAuth, namespace string) (*commonConfig.BasicAuth, error) {
	un, err := cg.Secrets.GetSecretValue(namespace, auth.Username)
	if err != nil {
		return nil, err
	}
	pw, err := cg.Secrets.GetSecretValue(namespace, auth.Password)
	if err != nil {
		return nil, err
	}
	return &commonConfig.BasicAuth{
		Username: un,
		Password: commonConfig.Secret(pw),
	}, nil
}

func (cg *ConfigGenerator) generateOauth2(oa promopv1.OAuth2, namespace string) (*commonConfig.OAuth2, error) {

	clid, err := cg.Secrets.SecretOrConfigMapValue(namespace, oa.ClientID)
	if err != nil {
		return nil, err
	}
	clisecret, err := cg.Secrets.GetSecretValue(namespace, oa.ClientSecret)
	if err != nil {
		return nil, err
	}
	return &commonConfig.OAuth2{
		Scopes:         oa.Scopes,
		TokenURL:       oa.TokenURL,
		EndpointParams: oa.EndpointParams,
		ClientID:       clid,
		ClientSecret:   commonConfig.Secret(clisecret),
	}, nil
}

func (cg *ConfigGenerator) generateAuthorization(a promopv1.SafeAuthorization, namespace string) (*commonConfig.Authorization, error) {

	auth := &commonConfig.Authorization{
		Type: a.Type,
	}
	if a.Credentials != nil {
		creds, err := cg.Secrets.GetSecretValue(namespace, *a.Credentials)
		if err != nil {
			return nil, err
		}
		auth.Credentials = commonConfig.Secret(creds)
	}
	return auth, nil
}

type relabeler struct {
	configs []*relabel.Config
}

// add adds a relabel config to the relabeler. It sets defaults from prometheus defaults.
func (r *relabeler) add(cfgs ...*relabel.Config) {
	for _, cfg := range cfgs {
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

// addFromV1 converts from an externally generated monitoringv1 RelabelConfig. Used for converting relabel rules generated by external package
func (r *relabeler) addFromV1(cfgs ...*promopv1.RelabelConfig) (err error) {
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
			cfg.Regex, err = relabel.NewRegexp(c.Regex)
			if err != nil {
				return err
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
	return nil
}

func (cg *ConfigGenerator) initRelabelings() relabeler {
	r := relabeler{}
	// Relabel prometheus job name into a meta label
	r.add(&relabel.Config{
		SourceLabels: model.LabelNames{"job"},
		TargetLabel:  "__tmp_prometheus_job_name",
	})
	return r
}

func sanitizeLabelName(name string) model.LabelName {
	return model.LabelName(invalidLabelCharRE.ReplaceAllString(name, "_"))
}

func (cg *ConfigGenerator) getNamespacesFromNamespaceSelector(nsel promopv1.NamespaceSelector, namespace string) []string {
	if nsel.Any {
		return []string{}
	} else if len(nsel.MatchNames) == 0 {
		return []string{namespace}
	}
	return nsel.MatchNames
}
