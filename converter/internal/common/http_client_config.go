package common

import (
	"github.com/grafana/agent/component/common/config"
	"github.com/grafana/agent/converter/diag"
	"github.com/grafana/river/rivertypes"
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

	diags.AddAll(ValidateSupported(NotEquals, httpClientConfig.NoProxy, "", "HTTP Client no_proxy", ""))
	diags.AddAll(ValidateSupported(Equals, httpClientConfig.ProxyFromEnvironment, true, "HTTP Client proxy_from_environment", ""))
	diags.AddAll(ValidateSupported(Equals, len(httpClientConfig.ProxyConnectHeader) > 0, true, "HTTP Client proxy_connect_header", ""))
	diags.AddAll(ValidateSupported(NotEquals, httpClientConfig.TLSConfig.MaxVersion, prom_config.TLSVersion(0), "HTTP Client max_version", ""))

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
