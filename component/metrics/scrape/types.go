package scrape

import (
	"fmt"
	"net/url"
	"strings"
	"time"

	"github.com/alecthomas/units"
	"github.com/hashicorp/hcl/v2"
	common_config "github.com/prometheus/common/config"
	"github.com/prometheus/common/model"
	"github.com/prometheus/prometheus/config"
	"github.com/rfratto/gohcl"
)

// var emptyAuthorization = common_config.Authorization{}
// var emptyBasicAuth = common_config.BasicAuth{}
// var emptyOAuth2 = &common_config.OAuth2{}

const bearer string = "Bearer"

// Config holds all of the attributes that can be used to configure a scrape
// component.
type Config struct {
	// TODO (@tpaschalis) I think we need to override this to be the same value
	// as the key in the targetGroups map that is being passed through the
	// channel, and not allow it to be freely set.
	// The job name to which the job label is set by default.
	JobName string `hcl:"job_name,attr"`

	// Indicator whether the scraped metrics should remain unmodified.
	HonorLabels bool `hcl:"honor_labels,optional"`
	// Indicator whether the scraped timestamps should be respected.
	HonorTimestamps bool `hcl:"honor_timestamps,optional"`
	// A set of query parameters with which the target is scraped.
	Params url.Values `hcl:"params,optional"`
	// How frequently to scrape the targets of this scrape config.
	ScrapeInterval model.Duration `hcl:"scrape_interval,optional"`
	// The timeout for scraping targets of this config.
	ScrapeTimeout model.Duration `hcl:"scrape_timeout,optional"`
	// The HTTP resource path on which to fetch metrics from targets.
	MetricsPath string `hcl:"metrics_path,optional"`
	// The URL scheme with which to fetch metrics from targets.
	Scheme string `hcl:"scheme,optional"`
	// An uncompressed response body larger than this many bytes will cause the
	// scrape to fail. 0 means no limit.
	BodySizeLimit units.Base2Bytes `hcl:"body_size_limit,optional"`
	// More than this many samples post metric-relabeling will cause the scrape to
	// fail.
	SampleLimit uint `hcl:"sample_limit,optional"`
	// More than this many targets after the target relabeling will cause the
	// scrapes to fail.
	TargetLimit uint `hcl:"target_limit,optional"`
	// More than this many labels post metric-relabeling will cause the scrape to
	// fail.
	LabelLimit uint `hcl:"label_limit,optional"`
	// More than this label name length post metric-relabeling will cause the
	// scrape to fail.
	LabelNameLengthLimit uint `hcl:"label_name_length_limit,optional"`
	// More than this label value length post metric-relabeling will cause the
	// scrape to fail.
	LabelValueLengthLimit uint `hcl:"label_value_length_limit,optional"`

	// HTTP Client Config
	BasicAuth     *BasicAuth     `hcl:"basic_auth,block"`
	Authorization *Authorization `hcl:"authorization,block"`
	OAuth2        *OAuth2Config  `hcl:"oauth2,block"`
	TLSConfig     *TLSConfig     `hcl:"tls_config,block"`

	BearerToken     string `hcl:"bearer_token,optional"`
	BearerTokenFile string `hcl:"bearer_token_file,optional"`
	ProxyURL        string `hcl:"proxy_url,optional"`

	FollowRedirects bool `hcl:"follow_redirects,optional"`
	EnableHTTP2     bool `hcl:"enable_http_2,optional"`
}

// BasicAuth configures Basic HTTP authentication credentials.
type BasicAuth struct {
	Username     string `hcl:"username,optional"`
	Password     string `hcl:"password,optional"`
	PasswordFile string `hcl:"password_file,optional"`
}

// Authorization sets up HTTP authorization credentials.
type Authorization struct {
	Type            string `hcl:"authorization_type,optional"`
	Credential      string `hcl:"authorization_credential,optional"`
	CredentialsFile string `hcl:"authorization_credentials_file,optional"`
}

// TLSConfig sets up options for TLS connections.
type TLSConfig struct {
	CAFile             string `hcl:"ca_file,optional"`
	CertFile           string `hcl:"cert_file,optional"`
	KeyFile            string `hcl:"key_file,optional"`
	ServerName         string `hcl:"server_name,optional"`
	InsecureSkipVerify bool   `hcl:"insecure_skip_verify,optional"`
}

// OAuth2Config sets up the OAuth2 client.
type OAuth2Config struct {
	ClientID         string            `hcl:"client_id,optional"`
	ClientSecret     string            `hcl:"client_secret,optional"`
	ClientSecretFile string            `hcl:"client_secret_file,optional"`
	Scopes           []string          `hcl:"scopes,optional"`
	TokenURL         string            `hcl:"token_url,optional"`
	EndpointParams   map[string]string `hcl:"endpoint_params,optional"`
	ProxyURL         string            `hcl:"proxy_url,optional"`
	TLSConfig        *TLSConfig        `hcl:"tls_config,optional"`
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

// DecodeHCL implements gohcl.Decoder.
// This method is only called on blocks, not objects.
func (c *Config) DecodeHCL(body hcl.Body, ctx *hcl.EvalContext) error {
	*c = DefaultConfig

	type scrapeConfig Config
	err := gohcl.DecodeBody(body, ctx, (*scrapeConfig)(c))
	if err != nil {
		return err
	}
	return nil
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
	// TODO(@tpaschalis) had to include this workaround with the OAuth2 config
	// object, otherwise it would get resolved to a non-nil object, trigger a
	// different behavior in the scrape requests and fail them. Let's check if
	// it's the same case with the other nested HTTPClientConfigs structs.
	// if reflect.DeepEqual(oauth2Config, emptyOAuth2) {
	// 	dec.HTTPClientConfig.OAuth2 = nil
	// } else {
	// 	dec.HTTPClientConfig.OAuth2 = oauth2Config
	// }

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
