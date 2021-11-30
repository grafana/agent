package config

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"strings"

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
	ExpandEnv bool

	URL              *url.URL
	HTTPClientConfig *config.HTTPClientConfig
}

// RemoteProvider ...
type RemoteProvider interface {
	Retrieve() (*Config, error)
}

// RemoteConfig ...
type RemoteConfig struct {
	RemoteProvider
}

// NewRemoteConfig ...
func NewRemoteConfig(rawURL string, opts *RemoteOpts) (*RemoteConfig, error) {
	u, err := url.Parse(rawURL)
	if err != nil {
		return nil, err
	}
	if opts == nil {
		// Default provider opts
		opts = &RemoteOpts{
			ExpandEnv: false,
		}
	}
	opts.URL = u

	switch u.Scheme {
	case "":
		// if no scheme, assume local file path, return nil and let caller handle.
		return nil, nil
	case HTTP, HTTPS:
		return &RemoteConfig{
			RemoteProvider: newHTTPProvider(opts),
		}, nil
	}
	return nil, fmt.Errorf("remote config scheme not supported: %s", u.Scheme)
}

// Remote Config Providers
// httpP - http/https provider
type httpP struct {
	myURL            *url.URL
	expandEnv        bool
	httpClientConfig *config.HTTPClientConfig
}

func newHTTPProvider(opts *RemoteOpts) *httpP {
	return &httpP{
		myURL:            opts.URL,
		expandEnv:        opts.ExpandEnv,
		httpClientConfig: opts.HTTPClientConfig,
	}
}

// Retrieve implements RemoteProvider and fetches the config
// TODO: token auth, oauth2, etc.
func (p httpP) Retrieve() (*Config, error) {
	var (
		bb  []byte
		err error

		request   *http.Request
		response  *http.Response
		basicAuth *config.BasicAuth
		client    *http.Client

		result = &Config{}
	)

	client, err = config.NewClientFromConfig(*p.httpClientConfig, "remote-config", nil)
	if err != nil {
		return nil, err
	}
	if p.httpClientConfig != nil {
		err = p.httpClientConfig.Validate()
		if err != nil {
			return nil, err
		}
		if p.httpClientConfig.BasicAuth != nil {
			basicAuth = p.httpClientConfig.BasicAuth
		}
	}

	request, err = http.NewRequest(http.MethodGet, p.myURL.String(), nil)
	if basicAuth != nil {
		if basicAuth.PasswordFile != "" {
			bs, err := ioutil.ReadFile(basicAuth.PasswordFile)
			if err != nil {
				return nil, fmt.Errorf("unable to read basic auth password file %s: %s", basicAuth.PasswordFile, err)
			}
			basicAuth.Password = config.Secret(strings.TrimSpace(string(bs)))
		}
		request.SetBasicAuth(basicAuth.Username, string(basicAuth.Password))
	}
	if err != nil {
		return nil, err
	}
	response, err = client.Do(request)
	if err != nil {
		return nil, err
	}

	if response.StatusCode < http.StatusOK || response.StatusCode >= http.StatusMultipleChoices {
		return nil, fmt.Errorf("error fetching config: status code: %d", response.StatusCode)
	}
	bb, err = ioutil.ReadAll(response.Body)
	if err != nil {
		return nil, err
	}
	err = LoadBytes(bb, p.expandEnv, result)
	if err != nil {
		return nil, err
	}
	return result, nil
}
