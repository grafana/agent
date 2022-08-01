// Package config contains types from github.com/prometheus/common/config, but changes them to be compatible with river
package config

import (
	"net/url"

	"github.com/prometheus/common/config"
)

// HTTPClientConfig mirrors config.HTTPClientConfig
type HTTPClientConfig struct {
	BasicAuth       *BasicAuth     `river:"basic_auth,block,optional"`
	Authorization   *Authorization `river:"authorization,block,optional"`
	OAuth2          *OAuth2Config  `river:"oauth2,block,optional"`
	BearerToken     Secret         `river:"bearer_token,attr,optional"`
	BearerTokenFile string         `river:"bearer_token_file,attr,optional"`
	ProxyURL        URL            `river:"proxy_url,attr,optional"`
	TLSConfig       TLSConfig      `river:"tls_config,block,optional"`
	FollowRedirects bool           `river:"follow_redirects,attr,optional"`
	EnableHTTP2     bool           `river:"enable_http2,attr,optional"`
}

// Convert converts our type to the native prometheus type
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

// DefaultHTTPCLientConfig for initializing objects
var DefaultHTTPClientConfig = HTTPClientConfig{
	FollowRedirects: true,
	EnableHTTP2:     true,
}

// BasicAuth configures Basic HTTP authentication credentials.
type BasicAuth struct {
	Username     string `river:"username,attr,optional"`
	Password     string `river:"password,attr,optional"`
	PasswordFile string `river:"password_file,attr,optional"`
}

// Convert converts our type to the native prometheus type
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

// URL mirrors config.URL
type URL string

// Convert converts our type to the native prometheus type
func (u URL) Convert() config.URL {
	if u == "" {
		return config.URL{}
	}
	urlp, _ := url.Parse(string(u))

	return config.URL{URL: urlp}
}

// Secret mirrors config.Secret
type Secret string

// Authorization sets up HTTP authorization credentials.
type Authorization struct {
	Type            string `river:"type,attr,optional"`
	Credentials     string `river:"credentials,attr,optional"`
	CredentialsFile string `river:"credentials_file,attr,optional"`
}

// Convert converts our type to the native prometheus type
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

// TLSVersion mirrors config.TLSVersion
type TLSVersion uint16

// TLSConfig sets up options for TLS connections.
type TLSConfig struct {
	CAFile             string     `river:"ca_file,attr,optional"`
	CertFile           string     `river:"cert_file,attr,optional"`
	KeyFile            string     `river:"key_file,attr,optional"`
	ServerName         string     `river:"server_name,attr,optional"`
	InsecureSkipVerify bool       `river:"insecure_skip_verify,attr,optional"`
	MinVersion         TLSVersion `river:"min_version,attr,optional"`
}

// Convert converts our type to the native prometheus type
func (t *TLSConfig) Convert() *config.TLSConfig {
	if t == nil {
		return nil
	}
	return &config.TLSConfig{
		CAFile:             t.CAFile,
		CertFile:           t.CertFile,
		KeyFile:            t.KeyFile,
		ServerName:         t.ServerName,
		InsecureSkipVerify: t.InsecureSkipVerify,
		MinVersion:         config.TLSVersion(t.MinVersion),
	}
}

// OAuth2Config sets up the OAuth2 client.
type OAuth2Config struct {
	ClientID         string            `river:"client_id,attr,optional"`
	ClientSecret     string            `river:"client_secret,attr,optional"`
	ClientSecretFile string            `river:"client_secret_file,attr,optional"`
	Scopes           []string          `river:"scopes,attr,optional"`
	TokenURL         string            `river:"token_url,attr,optional"`
	EndpointParams   map[string]string `river:"endpoint_params,attr,optional"`
	ProxyURL         URL               `river:"proxy_url,attr,optional"`
	TLSConfig        *TLSConfig        `river:"tls_config,attr,optional"`
}

// Convert converts our type to the native prometheus type
func (o *OAuth2Config) Convert() *config.OAuth2 {
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
