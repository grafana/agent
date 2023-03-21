package config

import (
	"fmt"
	"io"
	"net/http"
	"net/url"

	"github.com/grafana/agent/pkg/config/instrumentation"
	"github.com/prometheus/common/config"
)

// supported remote config provider schemes
const (
	httpScheme  = "http"
	httpsScheme = "https"
)

// remoteOpts struct contains agent remote config options
type remoteOpts struct {
	url              *url.URL
	HTTPClientConfig *config.HTTPClientConfig
}

// remoteConfigProvider interface should be implemented by config providers
type remoteConfigProvider interface {
	retrieve() ([]byte, error)
}

// newRemoteProvider constructs a new remote configuration provider. The rawURL is parsed
// and a provider is constructed based on the URL's scheme.
func newRemoteProvider(rawURL string, opts *remoteOpts) (remoteConfigProvider, error) {
	u, err := url.Parse(rawURL)
	if err != nil {
		return nil, fmt.Errorf("error parsing rawURL %s: %w", rawURL, err)
	}
	if opts == nil {
		// Default provider opts
		opts = &remoteOpts{}
	}
	opts.url = u

	switch u.Scheme {
	case "":
		// if no scheme, assume local file path, return nil and let caller handle.
		return nil, nil
	case httpScheme, httpsScheme:
		httpP, err := newHTTPProvider(opts)
		if err != nil {
			return nil, fmt.Errorf("error constructing httpProvider: %w", err)
		}
		return httpP, nil
	default:
		return nil, fmt.Errorf("remote config scheme not supported: %s", u.Scheme)
	}
}

// Remote Config Providers
// httpProvider - http/https provider
type httpProvider struct {
	myURL      *url.URL
	httpClient *http.Client
}

// newHTTPProvider constructs an new httpProvider
func newHTTPProvider(opts *remoteOpts) (*httpProvider, error) {
	httpClientConfig := config.HTTPClientConfig{}
	if opts.HTTPClientConfig != nil {
		err := opts.HTTPClientConfig.Validate()
		if err != nil {
			return nil, err
		}
		httpClientConfig = *opts.HTTPClientConfig
	}
	httpClient, err := config.NewClientFromConfig(httpClientConfig, "remote-config")
	if err != nil {
		return nil, err
	}
	return &httpProvider{
		myURL:      opts.url,
		httpClient: httpClient,
	}, nil
}

// retrieve implements remoteProvider and fetches the config
func (p httpProvider) retrieve() ([]byte, error) {
	response, err := p.httpClient.Get(p.myURL.String())
	if err != nil {
		instrumentation.InstrumentRemoteConfigFetchError()
		return nil, fmt.Errorf("request failed: %w", err)
	}
	defer response.Body.Close()

	instrumentation.InstrumentRemoteConfigFetch(response.StatusCode)

	if response.StatusCode/100 != 2 {
		return nil, fmt.Errorf("error fetching config: status code: %d", response.StatusCode)
	}
	bb, err := io.ReadAll(response.Body)
	if err != nil {
		return nil, err
	}
	return bb, nil
}
