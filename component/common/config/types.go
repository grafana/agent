// Package config contains types from github.com/prometheus/common/config,
// but modifiys them to be serializable with River.
package config

import (
	"fmt"
	"net/url"

	"github.com/grafana/agent/pkg/river"

	"github.com/grafana/agent/pkg/flow/rivertypes"
	"github.com/prometheus/common/config"
)

// HTTPClientConfig mirrors config.HTTPClientConfig
type HTTPClientConfig struct {
	BasicAuth       *BasicAuth        `river:"basic_auth,block,optional"`
	Authorization   *Authorization    `river:"authorization,block,optional"`
	OAuth2          *OAuth2Config     `river:"oauth2,block,optional"`
	BearerToken     rivertypes.Secret `river:"bearer_token,attr,optional"`
	BearerTokenFile string            `river:"bearer_token_file,attr,optional"`
	ProxyURL        URL               `river:"proxy_url,attr,optional"`
	TLSConfig       TLSConfig         `river:"tls_config,block,optional"`
	FollowRedirects bool              `river:"follow_redirects,attr,optional"`
	EnableHTTP2     bool              `river:"enable_http2,attr,optional"`
}

// UnmarshalRiver implements the umarshaller
func (h *HTTPClientConfig) UnmarshalRiver(f func(v interface{}) error) error {
	*h = HTTPClientConfig{
		FollowRedirects: true,
		EnableHTTP2:     true,
	}
	type config HTTPClientConfig
	return f((*config)(h))
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

// DefaultHTTPClientConfig for initializing objects
var DefaultHTTPClientConfig = HTTPClientConfig{
	FollowRedirects: true,
	EnableHTTP2:     true,
}

var _ river.Unmarshaler = (*HTTPClientConfig)(nil)

// BasicAuth configures Basic HTTP authentication credentials.
type BasicAuth struct {
	Username     string            `river:"username,attr,optional"`
	Password     rivertypes.Secret `river:"password,attr,optional"`
	PasswordFile string            `river:"password_file,attr,optional"`
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
type URL struct {
	*url.URL
}

// MarshalText implements encoding.TextMarshaler
func (u URL) MarshalText() (text []byte, err error) {
	u2 := &config.URL{
		URL: u.URL,
	}
	if u.URL != nil {
		return []byte(u2.Redacted()), nil
	}
	return nil, nil
}

// UnmarshalText implements encoding.TextUnmarshaler
func (u *URL) UnmarshalText(text []byte) error {
	s := string(text)
	urlp, err := url.Parse(s)
	if err != nil {
		return err
	}
	u.URL = urlp
	return nil
}

// Convert converts our type to the native prometheus type
func (u URL) Convert() config.URL {
	return config.URL{URL: u.URL}
}

// Authorization sets up HTTP authorization credentials.
type Authorization struct {
	Type            string            `river:"type,attr,optional"`
	Credentials     rivertypes.Secret `river:"credentials,attr,optional"`
	CredentialsFile string            `river:"credentials_file,attr,optional"`
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

// MarshalText implements encoding.TextMarshaler
func (tv TLSVersion) MarshalText() (text []byte, err error) {
	for s, v := range config.TLSVersions {
		if config.TLSVersion(tv) == v {
			return []byte(s), nil
		}
	}
	return nil, fmt.Errorf("unknown TLS version: %d", tv)
}

// UnmarshalText implements encoding.TextUnmarshaler
func (tv *TLSVersion) UnmarshalText(text []byte) error {
	if v, ok := config.TLSVersions[string(text)]; ok {
		*tv = TLSVersion(v)
		return nil
	}
	return fmt.Errorf("unknown TLS version: %s", string(text))
}

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
	ClientSecret     rivertypes.Secret `river:"client_secret,attr,optional"`
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
