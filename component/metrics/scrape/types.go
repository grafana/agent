package scrape

import (
	"fmt"
	"net/url"
	"strings"
	"time"

	"github.com/alecthomas/units"
	common_config "github.com/prometheus/common/config"
	"github.com/prometheus/common/model"
	"github.com/prometheus/prometheus/config"
)

const bearer string = "Bearer"

// Config holds all of the attributes that can be used to configure a scrape
// component.
type Config struct {
	// The job name to which the job label is set by default.
	JobName string `river:"job_name,attr"`

	// Indicator whether the scraped metrics should remain unmodified.
	HonorLabels bool `river:"honor_labels,attr,optional"`
	// Indicator whether the scraped timestamps should be respected.
	HonorTimestamps bool `river:"honor_timestamps,attr,optional"`
	// A set of query parameters with which the target is scraped.
	Params url.Values `river:"params,attr,optional"`
	// How frequently to scrape the targets of this scrape config.
	ScrapeInterval model.Duration `river:"scrape_interval,attr,optional"`
	// The timeout for scraping targets of this config.
	ScrapeTimeout model.Duration `river:"scrape_timeout,attr,optional"`
	// The HTTP resource path on which to fetch metrics from targets.
	MetricsPath string `river:"metrics_path,attr,optional"`
	// The URL scheme with which to fetch metrics from targets.
	Scheme string `river:"scheme,attr,optional"`
	// An uncompressed response body larger than this many bytes will cause the
	// scrape to fail. 0 means no limit.
	BodySizeLimit units.Base2Bytes `river:"body_size_limit,attr,optional"`
	// More than this many samples post metric-relabeling will cause the scrape to
	// fail.
	SampleLimit uint `river:"sample_limit,attr,optional"`
	// More than this many targets after the target relabeling will cause the
	// scrapes to fail.
	TargetLimit uint `river:"target_limit,attr,optional"`
	// More than this many labels post metric-relabeling will cause the scrape to
	// fail.
	LabelLimit uint `river:"label_limit,attr,optional"`
	// More than this label name length post metric-relabeling will cause the
	// scrape to fail.
	LabelNameLengthLimit uint `river:"label_name_length_limit,attr,optional"`
	// More than this label value length post metric-relabeling will cause the
	// scrape to fail.
	LabelValueLengthLimit uint `river:"label_value_length_limit,attr,optional"`

	// HTTP Client Config
	BasicAuth     *BasicAuth     `river:"basic_auth,block,optional"`
	Authorization *Authorization `river:"authorization,block,optional"`
	OAuth2        *OAuth2Config  `river:"oauth2,block,optional"`
	TLSConfig     *TLSConfig     `river:"tls_config,block,optional"`

	BearerToken     string `river:"bearer_token,attr,optional"`
	BearerTokenFile string `river:"bearer_token_file,attr,optional"`
	ProxyURL        string `river:"proxy_url,attr,optional"`

	FollowRedirects bool `river:"follow_redirects,attr,optional"`
	EnableHTTP2     bool `river:"enable_http_2,attr,optional"`
}

// BasicAuth configures Basic HTTP authentication credentials.
type BasicAuth struct {
	Username     string `river:"username,attr,optional"`
	Password     string `river:"password,attr,optional"`
	PasswordFile string `river:"password_file,attr,optional"`
}

// Authorization sets up HTTP authorization credentials.
type Authorization struct {
	Type            string `river:"authorization_type,attr,optional"`
	Credential      string `river:"authorization_credential,attr,optional"`
	CredentialsFile string `river:"authorization_credentials_file,attr,optional"`
}

// TLSConfig sets up options for TLS connections.
type TLSConfig struct {
	CAFile             string `river:"ca_file,attr,optional"`
	CertFile           string `river:"cert_file,attr,optional"`
	KeyFile            string `river:"key_file,attr,optional"`
	ServerName         string `river:"server_name,attr,optional"`
	InsecureSkipVerify bool   `river:"insecure_skip_verify,attr,optional"`
}

// OAuth2Config sets up the OAuth2 client.
type OAuth2Config struct {
	ClientID         string            `river:"client_id,attr,optional"`
	ClientSecret     string            `river:"client_secret,attr,optional"`
	ClientSecretFile string            `river:"client_secret_file,attr,optional"`
	Scopes           []string          `river:"scopes,attr,optional"`
	TokenURL         string            `river:"token_url,attr,optional"`
	EndpointParams   map[string]string `river:"endpoint_params,attr,optional"`
	ProxyURL         string            `river:"proxy_url,attr,optional"`
	TLSConfig        *TLSConfig        `river:"tls_config,attr,optional"`
}

// DefaultConfig is the set of default options applied before decoding a given
// scrape_config block.
var DefaultConfig = Config{
	MetricsPath:     "/metrics",
	Scheme:          "http",
	HonorLabels:     false,
	HonorTimestamps: true,
	FollowRedirects: true,                             // From common_config.DefaultHTTPClientConfig
	ScrapeInterval:  model.Duration(1 * time.Minute),  // From config.DefaultGlobalConfig
	ScrapeTimeout:   model.Duration(10 * time.Second), // From config.DefaultGlobalConfig
}

// UnmarshalRiver implements river.Unmarshaler.
func (c *Config) UnmarshalRiver(f func(interface{}) error) error {
	*c = DefaultConfig

	type scrapeConfig Config
	return f((*scrapeConfig)(c))
}

// Helper function to bridge the in-house configuration with the Prometheus
// scrape_config.
// As explained in the Config struct, the following fields are purposefully
// missing out, as they're being implemented by another components.
// - RelabelConfigs
// - MetricsRelabelConfigs
// - ServiceDiscoveryConfigs
func (c *Config) getPromScrapeConfigs(jobName string) (*config.ScrapeConfig, error) {
	dec := config.DefaultScrapeConfig
	dec.JobName = jobName
	dec.HonorLabels = c.HonorLabels
	dec.HonorTimestamps = c.HonorTimestamps
	dec.Params = c.Params
	dec.ScrapeInterval = c.ScrapeInterval
	dec.ScrapeTimeout = c.ScrapeTimeout
	dec.MetricsPath = c.MetricsPath
	dec.Scheme = c.Scheme
	dec.BodySizeLimit = c.BodySizeLimit
	dec.SampleLimit = c.SampleLimit
	dec.TargetLimit = c.TargetLimit
	dec.LabelLimit = c.LabelLimit
	dec.LabelNameLengthLimit = c.LabelNameLengthLimit
	dec.LabelValueLengthLimit = c.LabelValueLengthLimit

	// HTTP scrape client settings
	var proxyURL, oauth2ProxyURL *url.URL
	if c.ProxyURL != "" {
		proxyURL, _ = url.Parse(c.ProxyURL)
	}
	if c.OAuth2 != nil && c.OAuth2.ProxyURL != "" {
		oauth2ProxyURL, _ = url.Parse(c.OAuth2.ProxyURL)
	}
	httpClient := common_config.DefaultHTTPClientConfig
	dec.HTTPClientConfig = httpClient

	if c.BasicAuth != nil {
		dec.HTTPClientConfig.BasicAuth = &common_config.BasicAuth{
			Username:     c.BasicAuth.Username,
			Password:     common_config.Secret(c.BasicAuth.Password),
			PasswordFile: c.BasicAuth.PasswordFile,
		}
	}
	if c.Authorization != nil {
		dec.HTTPClientConfig.Authorization = &common_config.Authorization{
			Type:            c.Authorization.Type,
			Credentials:     common_config.Secret(c.Authorization.Credential),
			CredentialsFile: c.Authorization.CredentialsFile,
		}
	}

	if c.OAuth2 != nil {
		var oauth2TLSConfig common_config.TLSConfig
		if c.OAuth2.TLSConfig != nil {
			oauth2TLSConfig = common_config.TLSConfig{
				CAFile:             c.OAuth2.TLSConfig.CAFile,
				CertFile:           c.OAuth2.TLSConfig.CertFile,
				KeyFile:            c.OAuth2.TLSConfig.KeyFile,
				ServerName:         c.OAuth2.TLSConfig.ServerName,
				InsecureSkipVerify: c.OAuth2.TLSConfig.InsecureSkipVerify,
			}
		}
		dec.HTTPClientConfig.OAuth2 = &common_config.OAuth2{
			ClientID:         c.OAuth2.ClientID,
			ClientSecret:     common_config.Secret(c.OAuth2.ClientSecret),
			ClientSecretFile: c.OAuth2.ClientSecretFile,
			Scopes:           c.OAuth2.Scopes,
			TokenURL:         c.OAuth2.TokenURL,
			EndpointParams:   c.OAuth2.EndpointParams,
			ProxyURL:         common_config.URL{URL: oauth2ProxyURL},
			TLSConfig:        oauth2TLSConfig,
		}
	}

	dec.HTTPClientConfig.BearerToken = common_config.Secret(c.BearerToken)
	dec.HTTPClientConfig.BearerTokenFile = c.BearerTokenFile
	dec.HTTPClientConfig.ProxyURL = common_config.URL{URL: proxyURL}
	if c.TLSConfig != nil {
		dec.HTTPClientConfig.TLSConfig = common_config.TLSConfig{
			CAFile:             c.TLSConfig.CAFile,
			CertFile:           c.TLSConfig.CertFile,
			KeyFile:            c.TLSConfig.KeyFile,
			ServerName:         c.TLSConfig.ServerName,
			InsecureSkipVerify: c.TLSConfig.InsecureSkipVerify,
		}
	}
	dec.HTTPClientConfig.FollowRedirects = c.FollowRedirects
	dec.HTTPClientConfig.EnableHTTP2 = c.EnableHTTP2

	err := validateHTTPClientConfig(dec.HTTPClientConfig)
	if err != nil {
		return nil, fmt.Errorf("the provided scrape_config resulted in an invalid HTTP Client configuration: %w", err)
	}

	return &dec, nil
}

func validateHTTPClientConfig(c common_config.HTTPClientConfig) error {
	// Backwards compatibility with the bearer_token field.
	if len(c.BearerToken) > 0 && len(c.BearerTokenFile) > 0 {
		return fmt.Errorf("at most one of bearer_token & bearer_token_file must be configured")
	}
	if (c.BasicAuth != nil || c.OAuth2 != nil) && (len(c.BearerToken) > 0 || len(c.BearerTokenFile) > 0) {
		return fmt.Errorf("at most one of basic_auth, oauth2, bearer_token & bearer_token_file must be configured")
	}
	if c.BasicAuth != nil && (string(c.BasicAuth.Password) != "" && c.BasicAuth.PasswordFile != "") {
		return fmt.Errorf("at most one of basic_auth password & password_file must be configured")
	}
	if c.Authorization != nil {
		if len(c.BearerToken) > 0 || len(c.BearerTokenFile) > 0 {
			return fmt.Errorf("authorization is not compatible with bearer_token & bearer_token_file")
		}
		if string(c.Authorization.Credentials) != "" && c.Authorization.CredentialsFile != "" {
			return fmt.Errorf("at most one of authorization credentials & credentials_file must be configured")
		}
		c.Authorization.Type = strings.TrimSpace(c.Authorization.Type)
		if len(c.Authorization.Type) == 0 {
			c.Authorization.Type = bearer
		}
		if strings.ToLower(c.Authorization.Type) == "basic" {
			return fmt.Errorf(`authorization type cannot be set to "basic", use "basic_auth" instead`)
		}
		if c.BasicAuth != nil || c.OAuth2 != nil {
			return fmt.Errorf("at most one of basic_auth, oauth2 & authorization must be configured")
		}
	} else {
		if len(c.BearerToken) > 0 {
			c.Authorization = &common_config.Authorization{Credentials: c.BearerToken}
			c.Authorization.Type = bearer
			c.BearerToken = ""
		}
		if len(c.BearerTokenFile) > 0 {
			c.Authorization = &common_config.Authorization{CredentialsFile: c.BearerTokenFile}
			c.Authorization.Type = bearer
			c.BearerTokenFile = ""
		}
	}
	if c.OAuth2 != nil {
		if c.BasicAuth != nil {
			return fmt.Errorf("at most one of basic_auth, oauth2 & authorization must be configured")
		}
		if len(c.OAuth2.ClientID) == 0 {
			return fmt.Errorf("oauth2 client_id must be configured")
		}
		if len(c.OAuth2.ClientSecret) == 0 && len(c.OAuth2.ClientSecretFile) == 0 {
			return fmt.Errorf("either oauth2 client_secret or client_secret_file must be configured")
		}
		if len(c.OAuth2.TokenURL) == 0 {
			return fmt.Errorf("oauth2 token_url must be configured")
		}
		if len(c.OAuth2.ClientSecret) > 0 && len(c.OAuth2.ClientSecretFile) > 0 {
			return fmt.Errorf("at most one of oauth2 client_secret & client_secret_file must be configured")
		}
	}
	return nil
}
