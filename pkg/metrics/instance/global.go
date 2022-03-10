package instance

import (
	"net/url"

	"github.com/prometheus/prometheus/config"

	"github.com/grafana/agent/pkg/config/converter"
	promconfig "github.com/prometheus/common/config"
	"github.com/prometheus/common/model"
	"github.com/prometheus/common/sigv4"
	"github.com/prometheus/prometheus/pkg/labels"
	"github.com/prometheus/prometheus/pkg/relabel"
	"gopkg.in/yaml.v3"
)

// DefaultGlobalConfig holds default global settings to be used across all instances.
var DefaultGlobalConfig = GlobalConfig{
	Prometheus: config.DefaultGlobalConfig,
}

// GlobalConfig holds global settings that apply to all instances by default.
type GlobalConfig struct {
	Prometheus  config.GlobalConfig         `yaml:",inline"`
	RemoteWrite []*config.RemoteWriteConfig `yaml:"remote_write,omitempty"`
}

// UnmarshalYAML implements yaml.Unmarshaler.
func (c *GlobalConfig) UnmarshalYAML(unmarshal func(interface{}) error) error {
	*c = DefaultGlobalConfig

	type plain GlobalConfig
	return unmarshal((*plain)(c))
}

var _ converter.Configurator = (*GlobalConfigV2)(nil)

// GlobalConfigV2 holds global settings that apply to all instances by default.
type GlobalConfigV2 struct {
	Prometheus  PrometheusGlobalConfig `yaml:",inline"`
	RemoteWrite []*RemoteWriteConfig   `yaml:"remote_write,omitempty"`

	node *yaml.Node
}

func (c *GlobalConfigV2) UnmarshalYAML(value *yaml.Node) error {
	c.node = value
	type plain GlobalConfigV2
	return value.Decode((*plain)(c))
}

func (c *GlobalConfigV2) RedactSecret(redaction string) {

}

func (c *GlobalConfigV2) ApplyDefaults() error {
	//TODO implement me
	panic("implement me")
}

type PrometheusGlobalConfig struct {
	// How frequently to scrape targets by default.
	ScrapeInterval model.Duration `yaml:"scrape_interval,omitempty" default:"1m"`
	// The default timeout when scraping targets.
	ScrapeTimeout model.Duration `yaml:"scrape_timeout,omitempty" default:"10s"`
	// How frequently to evaluate rules by default.
	EvaluationInterval model.Duration `yaml:"evaluation_interval,omitempty" default:"1m"`
	// File to which PromQL queries are logged.
	QueryLogFile string `yaml:"query_log_file,omitempty"`
	// The labels to add to any timeseries that this Prometheus instance scrapes.
	ExternalLabels labels.Labels `yaml:"external_labels,omitempty"`
}

// RemoteWriteConfig is the configuration for writing to remote storage.
type RemoteWriteConfig struct {
	URL                 *promconfig.URL   `yaml:"url"`
	RemoteTimeout       model.Duration    `yaml:"remote_timeout,omitempty"`
	Headers             map[string]string `yaml:"headers,omitempty"`
	WriteRelabelConfigs []*relabel.Config `yaml:"write_relabel_configs,omitempty"`
	Name                string            `yaml:"name,omitempty"`
	SendExemplars       bool              `yaml:"send_exemplars,omitempty"`

	// We cannot do proper Go type embedding below as the parser will then parse
	// values arbitrarily into the overflow maps of further-down types.
	HTTPClientConfig promconfig.HTTPClientConfig `yaml:",inline"`
	QueueConfig      QueueConfig                 `yaml:"queue_config,omitempty"`
	MetadataConfig   MetadataConfig              `yaml:"metadata_config,omitempty"`
	SigV4Config      *sigv4.SigV4Config          `yaml:"sigv4,omitempty"`
}

var _ converter.Configurator = (*QueueConfig)(nil)

// QueueConfig is the configuration for the queue used to write to remote
// storage.
type QueueConfig struct {
	// Number of samples to buffer per shard before we block. Defaults to
	// MaxSamplesPerSend.
	Capacity int `yaml:"capacity,omitempty"`

	// Max number of shards, i.e. amount of concurrency.
	MaxShards int `yaml:"max_shards,omitempty"`

	// Min number of shards, i.e. amount of concurrency.
	MinShards int `yaml:"min_shards,omitempty"`

	// Maximum number of samples per send.
	MaxSamplesPerSend int `yaml:"max_samples_per_send,omitempty"`

	// Maximum time sample will wait in buffer.
	BatchSendDeadline model.Duration `yaml:"batch_send_deadline,omitempty"`

	// On recoverable errors, backoff exponentially.
	MinBackoff       model.Duration `yaml:"min_backoff,omitempty"`
	MaxBackoff       model.Duration `yaml:"max_backoff,omitempty"`
	RetryOnRateLimit bool           `yaml:"retry_on_http_429,omitempty"`
}

func (q *QueueConfig) RedactSecret(_ string) {
	return
}

func (q *QueueConfig) ApplyDefaults() error {
	return nil
}

var _ converter.Configurator = (*HTTPClientConfig)(nil)

// HTTPClientConfig configures an HTTP client.
type HTTPClientConfig struct {
	// The HTTP basic authentication credentials for the targets.
	BasicAuth *BasicAuth `yaml:"basic_auth,omitempty" json:"basic_auth,omitempty"`
	// The HTTP authorization credentials for the targets.
	Authorization *Authorization `yaml:"authorization,omitempty" json:"authorization,omitempty"`
	// The OAuth2 client credentials used to fetch a token for the targets.
	OAuth2 *OAuth2 `yaml:"oauth2,omitempty" json:"oauth2,omitempty"`
	// The bearer token for the targets. Deprecated in favour of
	// Authorization.Credentials.
	BearerToken string `yaml:"bearer_token,omitempty" json:"bearer_token,omitempty"`
	// The bearer token file for the targets. Deprecated in favour of
	// Authorization.CredentialsFile.
	BearerTokenFile string `yaml:"bearer_token_file,omitempty" json:"bearer_token_file,omitempty"`
	// HTTP proxy server to use to connect to the targets.
	ProxyURL URL `yaml:"proxy_url,omitempty" json:"proxy_url,omitempty"`
	// TLSConfig to use to connect to the targets.
	TLSConfig TLSConfig `yaml:"tls_config,omitempty" json:"tls_config,omitempty"`
	// FollowRedirects specifies whether the client should follow HTTP 3xx redirects.
	// The omitempty flag is not set, because it would be hidden from the
	// marshalled configuration when set to false.
	FollowRedirects bool `yaml:"follow_redirects" json:"follow_redirects"`
}

func (h *HTTPClientConfig) RedactSecret(redaction string) {
	h.OAuth2.RedactSecret(redaction)
	h.Authorization.RedactSecret(redaction)
	h.ProxyURL.RedactSecret(redaction)
	h.TLSConfig.RedactSecret(redaction)
}

func (h *HTTPClientConfig) ApplyDefaults() error {
	err := h.OAuth2.ApplyDefaults()
	if err != nil {
		return err
	}
	err = h.Authorization.ApplyDefaults()
	if err != nil {
		return err
	}
	err = h.ProxyURL.ApplyDefaults()
	if err != nil {
		return err
	}
	err = h.TLSConfig.ApplyDefaults()
	if err != nil {
		return err
	}
	return nil
}

var _ converter.Configurator = (*URL)(nil)

type URL struct {
	*url.URL
}

func (u *URL) RedactSecret(redaction string) {
	ru := *u.URL
	if _, ok := ru.User.Password(); ok {
		// We can not use secretToken because it would be escaped.
		ru.User = url.UserPassword(ru.User.Username(), redaction)
	}
}

func (u *URL) ApplyDefaults() error {
	return nil
}

var _ converter.Configurator = (*BasicAuth)(nil)

// BasicAuth contains basic HTTP authentication credentials.
type BasicAuth struct {
	Username     string `yaml:"username" json:"username"`
	Password     string `yaml:"password,omitempty" json:"password,omitempty"`
	PasswordFile string `yaml:"password_file,omitempty" json:"password_file,omitempty"`
}

func (b *BasicAuth) RedactSecret(redaction string) {
	b.Password = redaction
}

func (b *BasicAuth) ApplyDefaults() error {
	return nil
}

var _ converter.Configurator = (*Authorization)(nil)

// Authorization contains HTTP authorization credentials.
type Authorization struct {
	Type            string `yaml:"type,omitempty" json:"type,omitempty"`
	Credentials     string `yaml:"credentials,omitempty" json:"credentials,omitempty"`
	CredentialsFile string `yaml:"credentials_file,omitempty" json:"credentials_file,omitempty"`
}

func (a *Authorization) RedactSecret(redaction string) {
	a.Credentials = redaction
}

func (a *Authorization) ApplyDefaults() error {
	return nil
}

var _ converter.Configurator = (*OAuth2)(nil)

// OAuth2 is the oauth2 client configuration.
type OAuth2 struct {
	ClientID         string            `yaml:"client_id" json:"client_id"`
	ClientSecret     string            `yaml:"client_secret" json:"client_secret"`
	ClientSecretFile string            `yaml:"client_secret_file" json:"client_secret_file"`
	Scopes           []string          `yaml:"scopes,omitempty" json:"scopes,omitempty"`
	TokenURL         string            `yaml:"token_url" json:"token_url"`
	EndpointParams   map[string]string `yaml:"endpoint_params,omitempty" json:"endpoint_params,omitempty"`

	// TLSConfig is used to connect to the token URL.
	TLSConfig TLSConfig `yaml:"tls_config,omitempty"`
}

func (O *OAuth2) RedactSecret(redaction string) {
	O.ClientSecret = redaction
	O.TLSConfig.RedactSecret(redaction)
}

func (O *OAuth2) ApplyDefaults() error {
	return O.TLSConfig.ApplyDefaults()
}

var _ converter.Configurator = (*TLSConfig)(nil)

// TLSConfig configures the options for TLS connections.
type TLSConfig struct {
	// The CA cert to use for the targets.
	CAFile string `yaml:"ca_file,omitempty" json:"ca_file,omitempty"`
	// The client cert file for the targets.
	CertFile string `yaml:"cert_file,omitempty" json:"cert_file,omitempty"`
	// The client key file for the targets.
	KeyFile string `yaml:"key_file,omitempty" json:"key_file,omitempty"`
	// Used to verify the hostname for the targets.
	ServerName string `yaml:"server_name,omitempty" json:"server_name,omitempty"`
	// Disable target certificate validation.
	InsecureSkipVerify bool `yaml:"insecure_skip_verify" json:"insecure_skip_verify"`
}

func (T *TLSConfig) RedactSecret(_ string) {}

func (T *TLSConfig) ApplyDefaults() error {
	return nil
}

var _ converter.Configurator = (*MetadataConfig)(nil)

type MetadataConfig struct {
	// Send enables metric metadata to be sent to remote storage.
	Send bool `json:"send,omitempty"`
	// SendInterval controls how frequently metric metadata is sent to remote storage.
	SendInterval string `json:"sendInterval,omitempty"`
}

func (m *MetadataConfig) RedactSecret(_ string) {

}

func (m MetadataConfig) ApplyDefaults() error {
	return nil
}
