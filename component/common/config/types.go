// Package config contains types from github.com/prometheus/common/config,
// but modifies them to be serializable with River.
package config

import (
	"fmt"
	"net/url"
	"strings"

	"github.com/grafana/agent/pkg/river/rivertypes"
	"github.com/prometheus/common/config"
)

const bearerAuth string = "Bearer"

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

// SetToDefault implements the river.Defaulter
func (h *HTTPClientConfig) SetToDefault() {
	*h = DefaultHTTPClientConfig
}

// Validate returns an error if h is invalid.
func (h *HTTPClientConfig) Validate() error {
	// Backwards compatibility with the bearer_token field.
	if len(h.BearerToken) > 0 && len(h.BearerTokenFile) > 0 {
		return fmt.Errorf("at most one of bearer_token & bearer_token_file must be configured")
	}
	if (h.BasicAuth != nil || h.OAuth2 != nil) && (len(h.BearerToken) > 0 || len(h.BearerTokenFile) > 0) {
		return fmt.Errorf("at most one of basic_auth, oauth2, bearer_token & bearer_token_file must be configured")
	}
	if h.BasicAuth != nil && (string(h.BasicAuth.Password) != "" && h.BasicAuth.PasswordFile != "") {
		return fmt.Errorf("at most one of basic_auth password & password_file must be configured")
	}
	if h.Authorization != nil {
		if len(h.BearerToken) > 0 || len(h.BearerTokenFile) > 0 {
			return fmt.Errorf("authorization is not compatible with bearer_token & bearer_token_file")
		}
		if string(h.Authorization.Credentials) != "" && h.Authorization.CredentialsFile != "" {
			return fmt.Errorf("at most one of authorization credentials & credentials_file must be configured")
		}
		h.Authorization.Type = strings.TrimSpace(h.Authorization.Type)
		if len(h.Authorization.Type) == 0 {
			h.Authorization.Type = bearerAuth
		}
		if strings.ToLower(h.Authorization.Type) == "basic" {
			return fmt.Errorf(`authorization type cannot be set to "basic", use "basic_auth" instead`)
		}
		if h.BasicAuth != nil || h.OAuth2 != nil {
			return fmt.Errorf("at most one of basic_auth, oauth2 & authorization must be configured")
		}
	} else {
		if len(h.BearerToken) > 0 {
			h.Authorization = &Authorization{Credentials: h.BearerToken}
			h.Authorization.Type = bearerAuth
			h.BearerToken = ""
		}
		if len(h.BearerTokenFile) > 0 {
			h.Authorization = &Authorization{CredentialsFile: h.BearerTokenFile}
			h.Authorization.Type = bearerAuth
			h.BearerTokenFile = ""
		}
	}
	if h.OAuth2 != nil {
		if h.BasicAuth != nil {
			return fmt.Errorf("at most one of basic_auth, oauth2 & authorization must be configured")
		}
		if len(h.OAuth2.ClientID) == 0 {
			return fmt.Errorf("oauth2 client_id must be configured")
		}
		if len(h.OAuth2.ClientSecret) == 0 && len(h.OAuth2.ClientSecretFile) == 0 {
			return fmt.Errorf("either oauth2 client_secret or client_secret_file must be configured")
		}
		if len(h.OAuth2.TokenURL) == 0 {
			return fmt.Errorf("oauth2 token_url must be configured")
		}
		if len(h.OAuth2.ClientSecret) > 0 && len(h.OAuth2.ClientSecretFile) > 0 {
			return fmt.Errorf("at most one of oauth2 client_secret & client_secret_file must be configured")
		}
	}
	return nil
}

// Convert converts HTTPClientConfig to the native Prometheus type. If h is
// nil, the default client config is returned.
func (h *HTTPClientConfig) Convert() *config.HTTPClientConfig {
	if h == nil {
		return &config.DefaultHTTPClientConfig
	}

	return &config.HTTPClientConfig{
		BasicAuth:       h.BasicAuth.Convert(),
		Authorization:   h.Authorization.Convert(),
		OAuth2:          h.OAuth2.Convert(),
		BearerToken:     config.Secret(h.BearerToken),
		BearerTokenFile: h.BearerTokenFile,
		TLSConfig:       *h.TLSConfig.Convert(),
		FollowRedirects: h.FollowRedirects,
		EnableHTTP2:     h.EnableHTTP2,
		ProxyConfig: config.ProxyConfig{
			ProxyURL: h.ProxyURL.Convert(),
		},
	}
}

// Clone creates a shallow clone of h.
func CloneDefaultHTTPClientConfig() *HTTPClientConfig {
	clone := DefaultHTTPClientConfig
	return &clone
}

// DefaultHTTPClientConfig for initializing objects
var DefaultHTTPClientConfig = HTTPClientConfig{
	FollowRedirects: true,
	EnableHTTP2:     true,
}

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
	CA                 string            `river:"ca_pem,attr,optional"`
	CAFile             string            `river:"ca_file,attr,optional"`
	Cert               string            `river:"cert_pem,attr,optional"`
	CertFile           string            `river:"cert_file,attr,optional"`
	Key                rivertypes.Secret `river:"key_pem,attr,optional"`
	KeyFile            string            `river:"key_file,attr,optional"`
	ServerName         string            `river:"server_name,attr,optional"`
	InsecureSkipVerify bool              `river:"insecure_skip_verify,attr,optional"`
	MinVersion         TLSVersion        `river:"min_version,attr,optional"`
}

// UnmarshalRiver implements river.Unmarshaler and reports whether the
// unmarshaled TLSConfig is valid.
func (t *TLSConfig) UnmarshalRiver(f func(interface{}) error) error {
	type tlsConfig TLSConfig
	if err := f((*tlsConfig)(t)); err != nil {
		return err
	}

	return t.Validate()
}

// Convert converts our type to the native prometheus type
func (t *TLSConfig) Convert() *config.TLSConfig {
	if t == nil {
		return nil
	}
	return &config.TLSConfig{
		CA:                 t.CA,
		CAFile:             t.CAFile,
		Cert:               t.Cert,
		CertFile:           t.CertFile,
		Key:                config.Secret(t.Key),
		KeyFile:            t.KeyFile,
		ServerName:         t.ServerName,
		InsecureSkipVerify: t.InsecureSkipVerify,
		MinVersion:         config.TLSVersion(t.MinVersion),
	}
}

// Validate reports whether t is valid.
func (t *TLSConfig) Validate() error {
	if len(t.CA) > 0 && len(t.CAFile) > 0 {
		return fmt.Errorf("at most one of ca_pem and ca_file must be configured")
	}
	if len(t.Cert) > 0 && len(t.CertFile) > 0 {
		return fmt.Errorf("at most one of cert_pem and cert_file must be configured")
	}
	if len(t.Key) > 0 && len(t.KeyFile) > 0 {
		return fmt.Errorf("at most one of key_pem and key_file must be configured")
	}

	var (
		usingClientCert = len(t.Cert) > 0 || len(t.CertFile) > 0
		usingClientKey  = len(t.Key) > 0 || len(t.KeyFile) > 0
	)

	if usingClientCert && !usingClientKey {
		return fmt.Errorf("exactly one of key_pem or key_file must be configured when a client certificate is configured")
	} else if usingClientKey && !usingClientCert {
		return fmt.Errorf("exactly one of cert_pem or cert_file must be configured when a client key is configured")
	}

	return nil
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
	TLSConfig        *TLSConfig        `river:"tls_config,block,optional"`
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
		TLSConfig:        *o.TLSConfig.Convert(),
		ProxyConfig: config.ProxyConfig{
			ProxyURL: o.ProxyURL.Convert(),
		},
	}
}
