// Copyright The OpenTelemetry Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package oauth2clientauthextension // import "github.com/open-telemetry/opentelemetry-collector-contrib/extension/oauth2clientauthextension"

import (
	"context"
	"fmt"
	"net/http"

	"go.uber.org/multierr"
	"go.uber.org/zap"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/clientcredentials"
	"google.golang.org/grpc/credentials"
	grpcOAuth "google.golang.org/grpc/credentials/oauth"
)

// clientAuthenticator provides implementation for providing client authentication using OAuth2 client credentials
// workflow for both gRPC and HTTP clients.
type clientAuthenticator struct {
	clientCredentials *clientcredentials.Config
	logger            *zap.Logger
	client            *http.Client
}

type errorWrappingTokenSource struct {
	ts       oauth2.TokenSource
	tokenURL string
}

// errorWrappingTokenSource implements TokenSource
var _ oauth2.TokenSource = (*errorWrappingTokenSource)(nil)

// errFailedToGetSecurityToken indicates a problem communicating with OAuth2 server.
var errFailedToGetSecurityToken = fmt.Errorf("failed to get security token from token endpoint")

func newClientAuthenticator(cfg *Config, logger *zap.Logger) (*clientAuthenticator, error) {
	if cfg.ClientID == "" {
		return nil, errNoClientIDProvided
	}
	if cfg.ClientSecret == "" {
		return nil, errNoClientSecretProvided
	}
	if cfg.TokenURL == "" {
		return nil, errNoTokenURLProvided
	}

	transport := http.DefaultTransport.(*http.Transport).Clone()

	tlsCfg, err := cfg.TLSSetting.LoadTLSConfig()
	if err != nil {
		return nil, err
	}
	transport.TLSClientConfig = tlsCfg

	return &clientAuthenticator{
		clientCredentials: &clientcredentials.Config{
			ClientID:       cfg.ClientID,
			ClientSecret:   cfg.ClientSecret,
			TokenURL:       cfg.TokenURL,
			Scopes:         cfg.Scopes,
			EndpointParams: cfg.EndpointParams,
		},
		logger: logger,
		client: &http.Client{
			Transport: transport,
			Timeout:   cfg.Timeout,
		},
	}, nil
}

func (ewts errorWrappingTokenSource) Token() (*oauth2.Token, error) {
	tok, err := ewts.ts.Token()
	if err != nil {
		return tok, multierr.Combine(
			fmt.Errorf("%w (endpoint %q)", errFailedToGetSecurityToken, ewts.tokenURL),
			err)
	}
	return tok, nil
}

// roundTripper returns oauth2.Transport, an http.RoundTripper that performs "client-credential" OAuth flow and
// also auto refreshes OAuth tokens as needed.
func (o *clientAuthenticator) roundTripper(base http.RoundTripper) (http.RoundTripper, error) {
	ctx := context.WithValue(context.Background(), oauth2.HTTPClient, o.client)
	return &oauth2.Transport{
		Source: errorWrappingTokenSource{
			ts:       o.clientCredentials.TokenSource(ctx),
			tokenURL: o.clientCredentials.TokenURL,
		},
		Base: base,
	}, nil
}

// perRPCCredentials returns gRPC PerRPCCredentials that supports "client-credential" OAuth flow. The underneath
// oauth2.clientcredentials.Config instance will manage tokens performing auto refresh as necessary.
func (o *clientAuthenticator) perRPCCredentials() (credentials.PerRPCCredentials, error) {
	ctx := context.WithValue(context.Background(), oauth2.HTTPClient, o.client)
	return grpcOAuth.TokenSource{
		TokenSource: errorWrappingTokenSource{
			ts:       o.clientCredentials.TokenSource(ctx),
			tokenURL: o.clientCredentials.TokenURL,
		},
	}, nil
}
