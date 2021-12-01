package config

import (
	"fmt"
	"io/ioutil"
	"net/url"

	"github.com/prometheus/common/config"
)

// supported remote config provider schemes
const (
	HTTP  = "http"
	HTTPS = "https"
	// TODO: add s3, gcs, blob, and git providers backed by go-fsimpl
)

// RemoteOpts struct contains agent remote config options
type RemoteOpts struct {
	URL              *url.URL
	HTTPClientConfig *config.HTTPClientConfig
}

// RemoteProvider ...
type RemoteProvider interface {
	Retrieve() ([]byte, error)
}

// NewRemoteConfig ...
func NewRemoteConfig(rawURL string, opts *RemoteOpts) (RemoteProvider, error) {
	u, err := url.Parse(rawURL)
	if err != nil {
		return nil, err
	}
	if opts == nil {
		// Default provider opts
		opts = &RemoteOpts{}
	}
	opts.URL = u

	switch u.Scheme {
	case "":
		// if no scheme, assume local file path, return nil and let caller handle.
		return nil, nil
	case HTTP, HTTPS:
		httpP, err := newHTTPProvider(opts)
		if err != nil {
			return nil, err
		}
		return httpP, nil
	default:
		return nil, fmt.Errorf("remote config scheme not supported: %s", u.Scheme)
	}
}

// Remote Config Providers
// httpProvider - http/https provider
type httpProvider struct {
	myURL            url.URL
	httpClientConfig config.HTTPClientConfig
}

func newHTTPProvider(opts *RemoteOpts) (*httpProvider, error) {
	httpClientConfig := config.HTTPClientConfig{}
	if opts.HTTPClientConfig != nil {
		err := opts.HTTPClientConfig.Validate()
		if err != nil {
			return nil, err
		}
		httpClientConfig = *opts.HTTPClientConfig
	}
	return &httpProvider{
		myURL:            *opts.URL,
		httpClientConfig: httpClientConfig,
	}, nil
}

// Retrieve implements RemoteProvider and fetches the config
// TODO: token auth, oauth2, etc.
func (p httpProvider) Retrieve() ([]byte, error) {
	client, err := config.NewClientFromConfig(p.httpClientConfig, "remote-config")
	if err != nil {
		return nil, err
	}

	response, err := client.Get(p.myURL.String())
	if err != nil {
		return nil, err
	}
	defer response.Body.Close()

	if response.StatusCode/100 != 2 {
		return nil, fmt.Errorf("error fetching config: status code: %d", response.StatusCode)
	}
	bb, err := ioutil.ReadAll(response.Body)
	if err != nil {
		return nil, err
	}
	return bb, nil
}
