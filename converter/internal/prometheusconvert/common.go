package prometheusconvert

import (
	"github.com/grafana/agent/component/common/config"
	"github.com/grafana/agent/converter/diag"
	"github.com/grafana/agent/pkg/river/rivertypes"
	promconfig "github.com/prometheus/common/config"
)

func toHttpClientConfig(httpClientConfig *promconfig.HTTPClientConfig) *config.HTTPClientConfig {
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
		TLSConfig:       *toTLSConfig(&httpClientConfig.TLSConfig),
		FollowRedirects: httpClientConfig.FollowRedirects,
		EnableHTTP2:     httpClientConfig.EnableHTTP2,
	}
}

func validateHttpClientConfig(httpClientConfig *promconfig.HTTPClientConfig) diag.Diagnostics {
	var diags diag.Diagnostics

	if httpClientConfig.NoProxy != "" {
		diags.Add(diag.SeverityLevelWarn, "unsupported HTTP Client config no_proxy was provided")
	}

	if httpClientConfig.ProxyFromEnvironment {
		diags.Add(diag.SeverityLevelWarn, "unsupported HTTP Client config proxy_from_environment was provided")
	}

	if len(httpClientConfig.ProxyConnectHeader) > 0 {
		diags.Add(diag.SeverityLevelWarn, "unsupported HTTP Client config proxy_connect_header was provided")
	}

	return diags
}

func toBasicAuth(basicAuth *promconfig.BasicAuth) *config.BasicAuth {
	if basicAuth == nil {
		return nil
	}

	return &config.BasicAuth{
		Username:     basicAuth.Username,
		Password:     rivertypes.Secret(basicAuth.Password),
		PasswordFile: basicAuth.PasswordFile,
	}
}

func toAuthorization(authorization *promconfig.Authorization) *config.Authorization {
	if authorization == nil {
		return nil
	}

	return &config.Authorization{
		Type:            authorization.Type,
		Credentials:     rivertypes.Secret(authorization.Credentials),
		CredentialsFile: authorization.CredentialsFile,
	}
}

func toOAuth2(oAuth2 *promconfig.OAuth2) *config.OAuth2Config {
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
		TLSConfig:        toTLSConfig(&oAuth2.TLSConfig),
	}
}

func toTLSConfig(tlsConfig *promconfig.TLSConfig) *config.TLSConfig {
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
