package scrape

import (
	"fmt"
	"net/url"
	"strings"
	"time"

	"github.com/alecthomas/units"
	component_config "github.com/grafana/agent/component/common/config"
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
	ScrapeInterval time.Duration `river:"scrape_interval,attr,optional"`
	// The timeout for scraping targets of this config.
	ScrapeTimeout time.Duration `river:"scrape_timeout,attr,optional"`
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

	HTTPClientConfig component_config.HTTPClientConfig `river:"http_client_config,block,optional"`
}

// DefaultConfig is the set of default options applied before decoding a given
// scrape_config block.
var DefaultConfig = Config{
	MetricsPath:      "/metrics",
	Scheme:           "http",
	HonorLabels:      false,
	HonorTimestamps:  true,
	HTTPClientConfig: component_config.DefaultHTTPClientConfig,
	ScrapeInterval:   1 * time.Minute,  // From config.DefaultGlobalConfig
	ScrapeTimeout:    10 * time.Second, // From config.DefaultGlobalConfig
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
	dec.ScrapeInterval = model.Duration(c.ScrapeInterval)
	dec.ScrapeTimeout = model.Duration(c.ScrapeTimeout)
	dec.MetricsPath = c.MetricsPath
	dec.Scheme = c.Scheme
	dec.BodySizeLimit = c.BodySizeLimit
	dec.SampleLimit = c.SampleLimit
	dec.TargetLimit = c.TargetLimit
	dec.LabelLimit = c.LabelLimit
	dec.LabelNameLengthLimit = c.LabelNameLengthLimit
	dec.LabelValueLengthLimit = c.LabelValueLengthLimit

	// HTTP scrape client settings
	dec.HTTPClientConfig = *c.HTTPClientConfig.Convert()

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
