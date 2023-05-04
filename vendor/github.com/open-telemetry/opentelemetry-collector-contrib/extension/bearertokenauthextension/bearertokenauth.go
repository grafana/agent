// Copyright The OpenTelemetry Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//       http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package bearertokenauthextension // import "github.com/open-telemetry/opentelemetry-collector-contrib/extension/bearertokenauthextension"

import (
	"context"
	"fmt"
	"net/http"

	"go.opentelemetry.io/collector/component"
	"go.opentelemetry.io/collector/config/configauth"
	"go.uber.org/zap"
	"google.golang.org/grpc/credentials"
)

var _ credentials.PerRPCCredentials = (*PerRPCAuth)(nil)

// PerRPCAuth is a gRPC credentials.PerRPCCredentials implementation that returns an 'authorization' header.
type PerRPCAuth struct {
	metadata map[string]string
}

// GetRequestMetadata returns the request metadata to be used with the RPC.
func (c *PerRPCAuth) GetRequestMetadata(context.Context, ...string) (map[string]string, error) {
	return c.metadata, nil
}

// RequireTransportSecurity always returns true for this implementation. Passing bearer tokens in plain-text connections is a bad idea.
func (c *PerRPCAuth) RequireTransportSecurity() bool {
	return true
}

// BearerTokenAuth is an implementation of configauth.GRPCClientAuthenticator. It embeds a static authorization "bearer" token in every rpc call.
type BearerTokenAuth struct {
	tokenString string
	logger      *zap.Logger
}

var _ configauth.ClientAuthenticator = (*BearerTokenAuth)(nil)

func newBearerTokenAuth(cfg *Config, logger *zap.Logger) *BearerTokenAuth {
	return &BearerTokenAuth{
		tokenString: cfg.BearerToken,
		logger:      logger,
	}
}

// Start of BearerTokenAuth does nothing and returns nil
func (b *BearerTokenAuth) Start(ctx context.Context, host component.Host) error {
	return nil
}

// Shutdown of BearerTokenAuth does nothing and returns nil
func (b *BearerTokenAuth) Shutdown(ctx context.Context) error {
	return nil
}

// PerRPCCredentials returns PerRPCAuth an implementation of credentials.PerRPCCredentials that
func (b *BearerTokenAuth) PerRPCCredentials() (credentials.PerRPCCredentials, error) {
	return &PerRPCAuth{
		metadata: map[string]string{"authorization": b.bearerToken()},
	}, nil
}

func (b *BearerTokenAuth) bearerToken() string {
	return fmt.Sprintf("Bearer %s", b.tokenString)
}

// RoundTripper is not implemented by BearerTokenAuth
func (b *BearerTokenAuth) RoundTripper(base http.RoundTripper) (http.RoundTripper, error) {
	return &BearerAuthRoundTripper{
		baseTransport: base,
		bearerToken:   b.bearerToken(),
	}, nil
}

// BearerAuthRoundTripper intercepts and adds Bearer token Authorization headers to each http request.
type BearerAuthRoundTripper struct {
	baseTransport http.RoundTripper
	bearerToken   string
}

// RoundTrip modifies the original request and adds Bearer token Authorization headers.
func (interceptor *BearerAuthRoundTripper) RoundTrip(req *http.Request) (*http.Response, error) {
	req2 := req.Clone(req.Context())
	if req2.Header == nil {
		req2.Header = make(http.Header)
	}
	req2.Header.Set("Authorization", interceptor.bearerToken)
	return interceptor.baseTransport.RoundTrip(req2)
}
