package prometheusconvert

import (
	"github.com/grafana/agent/component/common/config"
	"github.com/grafana/agent/component/discovery"
	"github.com/grafana/agent/converter/diag"
	"github.com/grafana/agent/pkg/river/rivertypes"
	prom_config "github.com/prometheus/common/config"
)

func ToHttpClientConfig(httpClientConfig *prom_config.HTTPClientConfig) *config.HTTPClientConfig {
	if httpClientConfig == nil {
		return nil
	}

	return &config.HTTPClientConfig{
		BasicAuth:       toBasicAuth(httpClientConfig.BasicAuth),
		Authorization:   toAuthorization(httpClientConfig.Authorization),
		OAuth2:          toOAuth2(httpClientConfig.OAuth2),
		BearerToken:     rivertypes.Secret(httpClientConfig.BearerToken),
		BearerTokenFile: httpClientConfig.BearerTokenFile,
		ProxyURL:        config.URL(httpClientConfig.ProxyURL),
		TLSConfig:       *ToTLSConfig(&httpClientConfig.TLSConfig),
		FollowRedirects: httpClientConfig.FollowRedirects,
		EnableHTTP2:     httpClientConfig.EnableHTTP2,
	}
}

// ValidateHttpClientConfig returns [diag.Diagnostics] for currently
// unsupported Flow features available in Prometheus.
func ValidateHttpClientConfig(httpClientConfig *prom_config.HTTPClientConfig) diag.Diagnostics {
	var diags diag.Diagnostics

	if httpClientConfig.NoProxy != "" {
		diags.Add(diag.SeverityLevelError, "unsupported HTTP Client config no_proxy was provided")
	}

	if httpClientConfig.ProxyFromEnvironment {
		diags.Add(diag.SeverityLevelError, "unsupported HTTP Client config proxy_from_environment was provided")
	}

	if len(httpClientConfig.ProxyConnectHeader) > 0 {
		diags.Add(diag.SeverityLevelError, "unsupported HTTP Client config proxy_connect_header was provided")
	}

	if httpClientConfig.TLSConfig.MaxVersion != 0 {
		diags.Add(diag.SeverityLevelError, "unsupported HTTP Client config max_version was provided")
	}

	return diags
}

func toBasicAuth(basicAuth *prom_config.BasicAuth) *config.BasicAuth {
	if basicAuth == nil {
		return nil
	}

	return &config.BasicAuth{
		Username:     basicAuth.Username,
		Password:     rivertypes.Secret(basicAuth.Password),
		PasswordFile: basicAuth.PasswordFile,
	}
}

func toAuthorization(authorization *prom_config.Authorization) *config.Authorization {
	if authorization == nil {
		return nil
	}

	return &config.Authorization{
		Type:            authorization.Type,
		Credentials:     rivertypes.Secret(authorization.Credentials),
		CredentialsFile: authorization.CredentialsFile,
	}
}

func toOAuth2(oAuth2 *prom_config.OAuth2) *config.OAuth2Config {
	if oAuth2 == nil {
		return nil
	}

	return &config.OAuth2Config{
		ClientID:         oAuth2.ClientID,
		ClientSecret:     rivertypes.Secret(oAuth2.ClientSecret),
		ClientSecretFile: oAuth2.ClientSecretFile,
		Scopes:           oAuth2.Scopes,
		TokenURL:         oAuth2.TokenURL,
		EndpointParams:   oAuth2.EndpointParams,
		ProxyURL:         config.URL(oAuth2.ProxyURL),
		TLSConfig:        ToTLSConfig(&oAuth2.TLSConfig),
	}
}

func ToTLSConfig(tlsConfig *prom_config.TLSConfig) *config.TLSConfig {
	if tlsConfig == nil {
		return nil
	}

	return &config.TLSConfig{
		CA:                 tlsConfig.CA,
		CAFile:             tlsConfig.CAFile,
		Cert:               tlsConfig.Cert,
		CertFile:           tlsConfig.CertFile,
		Key:                rivertypes.Secret(tlsConfig.Key),
		KeyFile:            tlsConfig.KeyFile,
		ServerName:         tlsConfig.ServerName,
		InsecureSkipVerify: tlsConfig.InsecureSkipVerify,
		MinVersion:         config.TLSVersion(tlsConfig.MinVersion),
	}
}

// NewDiscoverExports will return a new [discovery.Exports] with a specific
// key for converter component exports. The argument will be tokenized
// as a component export string rather than the standard [discovery.Target]
// RiverTokenize.
func NewDiscoverExports(expr string) discovery.Exports {
	return discovery.Exports{
		Targets: newDiscoveryTargets(expr),
	}
}

// newDiscoveryTargets will return a new [[]discovery.Target] with a specific
// key for converter component exports. The argument will be tokenized
// as a component export string rather than the standard [discovery.Target]
// RiverTokenize.
func newDiscoveryTargets(expr string) []discovery.Target {
	return []discovery.Target{map[string]string{"__expr__": expr}}
}
