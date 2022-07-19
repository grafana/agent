package kubernetes

import (
	"net/url"

	"github.com/prometheus/common/config"
)

// TODO (cpeterson) move HTTPClientConfig and subtypes to dedicated package. make sure all custom serializers work
type HTTPClientConfig struct {
	// The HTTP basic authentication credentials for the targets.
	BasicAuth *BasicAuth `hcl:"basic_auth,optional"`
	// The HTTP authorization credentials for the targets.
	Authorization *Authorization `hcl:"authorization,optional"`
	// The OAuth2 client credentials used to fetch a token for the targets.
	OAuth2 *OAuth2 `hcl:"oauth2,optional"`
	// The bearer token for the targets. Deprecated in favour of
	// Authorization.Credentials.
	BearerToken Secret `hcl:"bearer_token,optional"`
	// The bearer token file for the targets. Deprecated in favour of
	// Authorization.CredentialsFile.
	BearerTokenFile string `hcl:"bearer_token_file,optional"`
	// HTTP proxy server to use to connect to the targets.
	ProxyURL URL `hcl:"proxy_url,optional"`
	// TLSConfig to use to connect to the targets.
	TLSConfig TLSConfig `hcl:"tls_config,optional"`
	// FollowRedirects specifies whether the client should follow HTTP 3xx redirects.
	// The optional flag is not set, because it would be hidden from the
	// marshalled configuration when set to false.
	FollowRedirects bool `hcl:"follow_redirects"`
	// EnableHTTP2 specifies whether the client should configure HTTP2.
	// The optional flag is not set, because it would be hidden from the
	// marshalled configuration when set to false.
	EnableHTTP2 bool `hcl:"enable_http2"`
}

func (h *HTTPClientConfig) Convert() *config.HTTPClientConfig {
	return &config.HTTPClientConfig{
		BasicAuth:       h.BasicAuth.Convert(),
		Authorization:   h.Authorization.Convert(),
		OAuth2:          h.OAuth2.Convert(),
		BearerToken:     config.Secret(h.BearerToken),
		BearerTokenFile: h.BearerTokenFile,
		ProxyURL:        h.ProxyURL.Convert(),
		TLSConfig:       *h.TLSConfig.Convert(),
		FollowRedirects: h.FollowRedirects,
		EnableHTTP2:     h.EnableHTTP2,
	}
}

type URL string

func (u URL) Convert() config.URL {
	urlp, _ := url.Parse(string(u))

	return config.URL{URL: urlp}
}

type BasicAuth struct {
	Username     string `hcl:"username"`
	Password     Secret `hcl:"password,optional"`
	PasswordFile string `hcl:"password_file,optional"`
}

func (b *BasicAuth) Convert() *config.BasicAuth {
	if b == nil {
		return nil
	}
	return &config.BasicAuth{
		Username:     b.Username,
		Password:     config.Secret(b.Password),
		PasswordFile: b.PasswordFile,
	}
}

type Secret string

type Authorization struct {
	Type            string `hcl:"type,optional"`
	Credentials     Secret `hcl:"credentials,optional"`
	CredentialsFile string `hcl:"credentials_file,optional"`
}

func (a *Authorization) Convert() *config.Authorization {
	if a == nil {
		return nil
	}
	return &config.Authorization{
		Type:            a.Type,
		Credentials:     config.Secret(a.Credentials),
		CredentialsFile: a.CredentialsFile,
	}
}

type OAuth2 struct {
	ClientID         string            `hcl:"client_id"`
	ClientSecret     Secret            `hcl:"client_secret"`
	ClientSecretFile string            `hcl:"client_secret_file"`
	Scopes           []string          `hcl:"scopes,optional"`
	TokenURL         string            `hcl:"token_url"`
	EndpointParams   map[string]string `hcl:"endpoint_params,optional"`

	// HTTP proxy server to use to connect to the targets.
	ProxyURL URL `hcl:"proxy_url,optional"`
	// TLSConfig is used to connect to the token URL.
	TLSConfig TLSConfig `hcl:"tls_config,optional"`
}

func (o *OAuth2) Convert() *config.OAuth2 {
	if o == nil {
		return nil
	}
	return &config.OAuth2{
		ClientID:         o.ClientID,
		ClientSecret:     config.Secret(o.ClientSecret),
		ClientSecretFile: o.ClientSecretFile,
		Scopes:           o.Scopes,
		TokenURL:         o.TokenURL,
		EndpointParams:   o.EndpointParams,
		ProxyURL:         o.ProxyURL.Convert(),
		TLSConfig:        *o.TLSConfig.Convert(),
	}
}

type TLSConfig struct {
	// The CA cert to use for the targets.
	CAFile string `hcl:"ca_file,optional"`
	// The client cert file for the targets.
	CertFile string `hcl:"cert_file,optional"`
	// The client key file for the targets.
	KeyFile string `hcl:"key_file,optional"`
	// Used to verify the hostname for the targets.
	ServerName string `hcl:"server_name,optional"`
	// Disable target certificate validation.
	InsecureSkipVerify bool `hcl:"insecure_skip_verify"`
	// Minimum TLS version.
	MinVersion TLSVersion `hcl:"min_version,optional"`
}

func (t *TLSConfig) Convert() *config.TLSConfig {
	return &config.TLSConfig{
		CAFile:             t.CAFile,
		CertFile:           t.CertFile,
		KeyFile:            t.KeyFile,
		ServerName:         t.ServerName,
		InsecureSkipVerify: t.InsecureSkipVerify,
		MinVersion:         config.TLSVersion(t.MinVersion),
	}
}

type TLSVersion uint16
