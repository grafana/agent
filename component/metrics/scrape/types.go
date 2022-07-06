package scrape

import (
	"fmt"
	"net/url"
	"reflect"
	"strings"
	"time"

	"github.com/alecthomas/units"
	"github.com/go-kit/log/level"
	"github.com/hashicorp/hcl/v2"
	common_config "github.com/prometheus/common/config"
	"github.com/prometheus/common/model"
	"github.com/prometheus/prometheus/config"
	"github.com/rfratto/gohcl"
)

var emptyAuthorization = common_config.Authorization{}
var emptyBasicAuth = common_config.BasicAuth{}
var emptyOAuth2 = &common_config.OAuth2{}

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
	BasicAuthUsername     string `hcl:"basic_auth_username,optional"`
	BasicAuthPassword     string `hcl:"basic_auth_password,optional"`
	BasicAuthPasswordFile string `hcl:"basic_auth_password_file,optional"`

	AuthorizationType            string `hcl:"authorization_type,optional"`
	AuthorizationCredential      string `hcl:"authorization_credential,optional"`
	AuthorizationCredentialsFile string `hcl:"authorization_credentials_file,optional"`

	OAuth2ClientID                    string            `hcl:"oauth2_client_id,optional"`
	OAuth2ClientSecret                string            `hcl:"oauth2_client_secret,optional"`
	OAuth2ClientSecretFile            string            `hcl:"oauth2_client_secret_file,optional"`
	OAuth2Scopes                      []string          `hcl:"oauth2_scopes,optional"`
	OAuth2TokenURL                    string            `hcl:"oauth2_token_url,optional"`
	OAuth2EndpointParams              map[string]string `hcl:"oauth2_endpoint_params,optional"`
	OAuth2ProxyURL                    string            `hcl:"oauth2_proxy_url,optional"`
	OAuth2TLSConfigCAFile             string            `hcl:"oauth2_tls_config_ca_file,optional"`
	OAuth2TLSConfigCertFile           string            `hcl:"oauth2_tls_config_cert_file,optional"`
	OAuth2TLSConfigKeyFile            string            `hcl:"oauth2_tls_config_key_file,optional"`
	OAuth2TLSConfigServerName         string            `hcl:"oauth2_tls_config_server_name,optional"`
	OAuth2TLSConfigInsecureSkipVerify bool              `hcl:"oauth2_tls_config_insecure_skip_verify,optional"`

	BearerToken     string `hcl:"bearer_token,optional"`
	BearerTokenFile string `hcl:"bearer_token_file,optional"`
	ProxyURL        string `hcl:"proxy_url,optional"`

	TLSConfigCAFile             string `hcl:"tls_config_ca_file,optional"`
	TLSConfigCertFile           string `hcl:"tls_config_cert_file,optional"`
	TLSConfigKeyFile            string `hcl:"tls_config_key_file,optional"`
	TLSConfigServerName         string `hcl:"tls_config_server_name,optional"`
	TLSConfigInsecureSkipVerify bool   `hcl:"tls_config_insecure_skip_verify,optional"`

	FollowRedirects bool `hcl:"follow_redirects,optional"`
	EnableHTTP2     bool `hcl:"enable_http_2,optional"`
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
func (c *Component) getPromScrapeConfigs(scs []Config) []*config.ScrapeConfig {
	res := make([]*config.ScrapeConfig, 0)

	for _, sc := range scs {
		// General scrape settings
		dec := config.DefaultScrapeConfig
		dec.JobName = c.opts.ID + "/" + sc.JobName
		dec.HonorLabels = sc.HonorLabels
		dec.HonorTimestamps = sc.HonorTimestamps
		dec.Params = sc.Params
		dec.ScrapeInterval = model.Duration(sc.ScrapeInterval)
		dec.ScrapeTimeout = model.Duration(sc.ScrapeTimeout)
		dec.MetricsPath = sc.MetricsPath
		dec.Scheme = sc.Scheme
		dec.BodySizeLimit = sc.BodySizeLimit
		dec.SampleLimit = sc.SampleLimit
		dec.TargetLimit = sc.TargetLimit
		dec.LabelLimit = sc.LabelLimit
		dec.LabelNameLengthLimit = sc.LabelNameLengthLimit
		dec.LabelValueLengthLimit = sc.LabelValueLengthLimit

		// HTTP scrape client settings
		var proxyURL, oauth2ProxyURL *url.URL
		if sc.ProxyURL != "" {
			proxyURL, _ = url.Parse(sc.ProxyURL)
		}
		if sc.OAuth2ProxyURL != "" {
			oauth2ProxyURL, _ = url.Parse(sc.OAuth2ProxyURL)
		}
		httpClient := common_config.DefaultHTTPClientConfig
		dec.HTTPClientConfig = httpClient

		dec.HTTPClientConfig.BasicAuth = &common_config.BasicAuth{
			Username:     sc.BasicAuthUsername,
			Password:     common_config.Secret(sc.BasicAuthPassword),
			PasswordFile: sc.BasicAuthPasswordFile,
		}
		dec.HTTPClientConfig.Authorization = &common_config.Authorization{
			Type:            sc.AuthorizationType,
			Credentials:     common_config.Secret(sc.AuthorizationCredential),
			CredentialsFile: sc.AuthorizationCredentialsFile,
		}

		oauth2Config := &common_config.OAuth2{
			ClientID:         sc.OAuth2ClientID,
			ClientSecret:     common_config.Secret(sc.OAuth2ClientSecret),
			ClientSecretFile: sc.OAuth2ClientSecretFile,
			Scopes:           sc.OAuth2Scopes,
			TokenURL:         sc.OAuth2TokenURL,
			EndpointParams:   sc.OAuth2EndpointParams,
			ProxyURL:         common_config.URL{URL: oauth2ProxyURL},
			TLSConfig: common_config.TLSConfig{
				CAFile:             sc.OAuth2TLSConfigCAFile,
				CertFile:           sc.OAuth2TLSConfigCertFile,
				KeyFile:            sc.OAuth2TLSConfigKeyFile,
				ServerName:         sc.OAuth2TLSConfigServerName,
				InsecureSkipVerify: sc.OAuth2TLSConfigInsecureSkipVerify,
			},
		}
		// TODO(@tpaschalis) had to include this workaround with the OAuth2 config
		// object, otherwise it would get resolved to a non-nil object, trigger a
		// different behavior in the scrape requests and fail them. Let's check if
		// it's the same case with the other nested HTTPClientConfigs structs.
		if reflect.DeepEqual(oauth2Config, emptyOAuth2) {
			dec.HTTPClientConfig.OAuth2 = nil
		} else {
			dec.HTTPClientConfig.OAuth2 = oauth2Config
		}

		dec.HTTPClientConfig.BearerToken = common_config.Secret(sc.BearerToken)
		dec.HTTPClientConfig.BearerTokenFile = sc.BearerTokenFile
		dec.HTTPClientConfig.ProxyURL = common_config.URL{URL: proxyURL}
		dec.HTTPClientConfig.TLSConfig = common_config.TLSConfig{
			CAFile:             sc.TLSConfigCAFile,
			CertFile:           sc.TLSConfigCertFile,
			KeyFile:            sc.TLSConfigKeyFile,
			ServerName:         sc.TLSConfigServerName,
			InsecureSkipVerify: sc.TLSConfigInsecureSkipVerify,
		}
		dec.HTTPClientConfig.FollowRedirects = sc.FollowRedirects
		dec.HTTPClientConfig.EnableHTTP2 = sc.EnableHTTP2

		err := validateHTTPClientConfig(dec.HTTPClientConfig)
		if err != nil {
			level.Error(c.opts.Logger).Log("msg", "provided scrape_config resulted in an invalid HTTP client configuration, skipping it", "err", err)
		} else {
			res = append(res, &dec)
		}
	}

	return res
}

func validateHTTPClientConfig(c common_config.HTTPClientConfig) error {
	// Backwards compatibility with the bearer_token field.
	if len(c.BearerToken) > 0 && len(c.BearerTokenFile) > 0 {
		return fmt.Errorf("at most one of bearer_token & bearer_token_file must be configured")
	}
	if (*c.BasicAuth != emptyBasicAuth || c.OAuth2 != nil) && (len(c.BearerToken) > 0 || len(c.BearerTokenFile) > 0) {
		return fmt.Errorf("at most one of basic_auth, oauth2, bearer_token & bearer_token_file must be configured")
	}
	if *c.BasicAuth != emptyBasicAuth && (string(c.BasicAuth.Password) != "" && c.BasicAuth.PasswordFile != "") {
		return fmt.Errorf("at most one of basic_auth password & password_file must be configured")
	}
	if *c.Authorization != emptyAuthorization {
		if len(c.BearerToken) > 0 || len(c.BearerTokenFile) > 0 {
			return fmt.Errorf("authorization is not compatible with bearer_token & bearer_token_file")
		}
		if string(c.Authorization.Credentials) != "" && c.Authorization.CredentialsFile != "" {
			return fmt.Errorf("at most one of authorization credentials & credentials_file must be configured")
		}
		c.Authorization.Type = strings.TrimSpace(c.Authorization.Type)
		if len(c.Authorization.Type) == 0 {
			c.Authorization.Type = "Bearer"
		}
		if strings.ToLower(c.Authorization.Type) == "basic" {
			return fmt.Errorf(`authorization type cannot be set to "basic", use "basic_auth" instead`)
		}
		if *c.BasicAuth != emptyBasicAuth || c.OAuth2 != nil {
			return fmt.Errorf("at most one of basic_auth, oauth2 & authorization must be configured")
		}
	} else {
		if len(c.BearerToken) > 0 {
			c.Authorization = &common_config.Authorization{Credentials: c.BearerToken}
			c.Authorization.Type = "Bearer"
			c.BearerToken = ""
		}
		if len(c.BearerTokenFile) > 0 {
			c.Authorization = &common_config.Authorization{CredentialsFile: c.BearerTokenFile}
			c.Authorization.Type = "Bearer"
			c.BearerTokenFile = ""
		}
	}
	if c.OAuth2 != nil {
		if *c.BasicAuth != emptyBasicAuth {
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
