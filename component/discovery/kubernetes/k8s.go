package kubernetes

import (
	"context"
	"encoding/json"
	"fmt"
	"net/url"

	"github.com/grafana/agent/component"
	"github.com/prometheus/common/config"
	promk8s "github.com/prometheus/prometheus/discovery/kubernetes"
	"github.com/prometheus/prometheus/discovery/targetgroup"
)

// TODO: cpeterson: use defaults from here for hcl default
var _ = promk8s.DefaultSDConfig

func init() {
	component.Register(component.Registration{
		Name:    "discovery.k8s",
		Args:    SDConfig{},
		Exports: Exports{},

		Build: func(opts component.Options, args component.Arguments) (component.Component, error) {
			return New(opts, args.(SDConfig))
		},
	})
}

type SDConfig struct {
	APIServer          URL                `hcl:"api_server,optional"`
	Role               string             `hcl:"role"`
	KubeConfig         string             `hcl:"kubeconfig_file,optional"`
	HTTPClientConfig   HTTPClientConfig   `hcl:"http_client_config,optional"`
	NamespaceDiscovery NamespaceDiscovery `hcl:"namespaces,optional"`
	Selectors          []SelectorConfig   `hcl:"selectors,optional"`
}

func (sd *SDConfig) Convert() *promk8s.SDConfig {
	selectors := make([]promk8s.SelectorConfig, len(sd.Selectors))
	for i, s := range sd.Selectors {
		selectors[i] = *s.Convert()
	}
	return &promk8s.SDConfig{
		APIServer:          sd.APIServer.Convert(),
		Role:               promk8s.Role(sd.Role),
		KubeConfig:         sd.KubeConfig,
		HTTPClientConfig:   *sd.HTTPClientConfig.Convert(),
		NamespaceDiscovery: *sd.NamespaceDiscovery.Convert(),
		Selectors:          selectors,
	}
}

type NamespaceDiscovery struct {
	IncludeOwnNamespace bool     `hcl:"own_namespace,optional"`
	Names               []string `hcl:"names,optional"`
}

func (nd *NamespaceDiscovery) Convert() *promk8s.NamespaceDiscovery {
	return &promk8s.NamespaceDiscovery{
		IncludeOwnNamespace: nd.IncludeOwnNamespace,
		Names:               nd.Names,
	}
}

type SelectorConfig struct {
	Role  string `hcl:"role,optional"`
	Label string `hcl:"label,optional"`
	Field string `hcl:"field,optional"`
}

func (sc *SelectorConfig) Convert() *promk8s.SelectorConfig {
	return &promk8s.SelectorConfig{
		Role:  promk8s.Role(sc.Role),
		Label: sc.Label,
		Field: sc.Field,
	}
}

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

// Exports holds values which are exported by the discovery.k8s component.
type Exports struct {
	B string `hcl:"b,optional"`
}

// Component implements the discovery.k8s component.
type Component struct {
	opts   component.Options
	args   SDConfig
	cancel context.CancelFunc

	ch chan []*targetgroup.Group
}

// New creates a new discovery.k8s component.
func New(o component.Options, args SDConfig) (*Component, error) {
	c := &Component{
		opts: o,
		ch:   make(chan []*targetgroup.Group),
	}

	// Perform an update which will immediately set our exports
	if err := c.Update(args); err != nil {
		return nil, err
	}
	return c, nil
}

// Run implements component.Component.
func (c *Component) Run(ctx context.Context) error {
	for {
		select {
		case <-ctx.Done():
			if c.cancel != nil {
				c.cancel()
			}
			return nil
		case group := <-c.ch:
			d, _ := json.MarshalIndent(group, "", "  ")
			fmt.Println(string(d))
			fmt.Println(group[0].Source)
		}
	}

}

// Update implements component.Compnoent.
func (c *Component) Update(args component.Arguments) error {
	newArgs := args.(SDConfig)
	fmt.Println(newArgs)
	disc, err := promk8s.New(c.opts.Logger, newArgs.Convert())
	if err != nil {
		return err
	}
	if c.cancel != nil {
		c.cancel()
	}
	ctx, cancel := context.WithCancel(context.Background())
	c.cancel = cancel
	go disc.Run(ctx, c.ch)
	return nil
}
