package server

import (
	"context"
	"crypto/tls"
	"fmt"
	"net"
	"net/http"
	"testing"

	"github.com/go-kit/log"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/health"
	"google.golang.org/grpc/health/grpc_health_v1"
)

const anyLocalhost = "127.0.0.1:0"

func TestServer(t *testing.T) {
	cfg := newTestConfig()
	srv := runExampleServer(t, cfg)

	// Validate HTTP
	resp, err := http.Get(fmt.Sprintf("http://%s/testing", srv.HTTPAddress()))
	require.NoError(t, err)
	require.Equal(t, http.StatusOK, resp.StatusCode)
	_ = resp.Body.Close()

	// Validate gRPC
	creds := grpc.WithTransportCredentials(insecure.NewCredentials())
	cc, err := grpc.Dial(srv.GRPCAddress().String(), creds)
	require.NoError(t, err)
	_, err = grpc_health_v1.NewHealthClient(cc).Check(context.Background(), &grpc_health_v1.HealthCheckRequest{})
	require.NoError(t, err)
}

func TestServer_InMemory(t *testing.T) {
	cfg := newTestConfig()
	srv := runExampleServer(t, cfg)

	// Validate HTTP
	var httpClient http.Client
	httpClient.Transport = &http.Transport{DialContext: srv.DialContext}
	resp, err := httpClient.Get(fmt.Sprintf("http://%s/testing", cfg.Flags.HTTP.InMemoryAddr))
	require.NoError(t, err)
	require.Equal(t, http.StatusOK, resp.StatusCode)
	_ = resp.Body.Close()

	// Validate gRPC
	grpcDialer := grpc.WithContextDialer(func(ctx context.Context, s string) (net.Conn, error) {
		return srv.DialContext(ctx, "", s)
	})
	cc, err := grpc.Dial(cfg.Flags.GRPC.InMemoryAddr, grpc.WithTransportCredentials(insecure.NewCredentials()), grpcDialer)
	require.NoError(t, err)
	_, err = grpc_health_v1.NewHealthClient(cc).Check(context.Background(), &grpc_health_v1.HealthCheckRequest{})
	require.NoError(t, err)
}

func newTestConfig() Config {
	cfg := DefaultConfig
	cfg.Flags.HTTP.ListenAddress = anyLocalhost
	cfg.Flags.GRPC.ListenAddress = anyLocalhost
	return cfg
}

func runExampleServer(t *testing.T, cfg Config) *Server {
	t.Helper()

	srv, err := New(log.NewNopLogger(), nil, nil, cfg)
	require.NoError(t, err)

	// Set up some expected services for us to test against.
	srv.HTTP.HandleFunc("/testing", func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
	})
	grpc_health_v1.RegisterHealthServer(srv.GRPC, health.NewServer())

	// Run our server.
	ctx, cancel := context.WithCancel(context.Background())
	t.Cleanup(cancel)
	go func() {
		require.NoError(t, srv.Run(ctx))
	}()

	return srv
}

func TestServer_TLS(t *testing.T) {
	cfg := newTestConfig()
	cfg.Flags.HTTP.UseTLS = true
	cfg.Flags.GRPC.UseTLS = true

	tlsConfig := TLSConfig{
		TLSCertPath: "testdata/example-cert.pem",
		TLSKeyPath:  "testdata/example-key.pem",
	}
	cfg.HTTP.TLSConfig = tlsConfig
	cfg.GRPC.TLSConfig = tlsConfig

	srv := runExampleServer(t, cfg)

	// Validate HTTPS
	cli := http.Client{
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
		},
	}
	resp, err := cli.Get(fmt.Sprintf("https://%s/testing", srv.HTTPAddress()))
	require.NoError(t, err)
	require.Equal(t, http.StatusOK, resp.StatusCode)
	_ = resp.Body.Close()

	// Validate gRPC TLS
	creds := credentials.NewTLS(&tls.Config{InsecureSkipVerify: true})
	cc, err := grpc.Dial(srv.GRPCAddress().String(), grpc.WithTransportCredentials(creds))
	require.NoError(t, err)
	_, err = grpc_health_v1.NewHealthClient(cc).Check(context.Background(), &grpc_health_v1.HealthCheckRequest{})
	require.NoError(t, err)
}

// TestRunReturnsError validates that Run exits with an error when the
// HTTP/GRPC servers stop unexpectedly.
func TestRunReturnsError(t *testing.T) {
	cfg := newTestConfig()

	t.Run("http", func(t *testing.T) {
		srv, err := New(nil, nil, nil, cfg)
		require.NoError(t, err)

		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()

		errChan := make(chan error, 1)
		go func() {
			errChan <- srv.Run(ctx)
		}()

		require.NoError(t, srv.httpListener.Close())
		require.NotNil(t, <-errChan)
	})

	t.Run("grpc", func(t *testing.T) {
		srv, err := New(nil, nil, nil, cfg)
		require.NoError(t, err)

		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()

		errChan := make(chan error, 1)
		go func() {
			errChan <- srv.Run(ctx)
		}()

		require.NoError(t, srv.grpcListener.Close())
		require.NotNil(t, <-errChan)
	})
}

func TestServer_ApplyConfig(t *testing.T) {
	t.Run("no changes", func(t *testing.T) {
		cfg := newTestConfig()

		srv, err := New(nil, nil, nil, cfg)
		require.NoError(t, err)

		require.NoError(t, srv.ApplyConfig(cfg))
	})

	t.Run("valid changes", func(t *testing.T) {
		cfg := newTestConfig()

		srv, err := New(nil, nil, nil, cfg)
		require.NoError(t, err)

		cfg.LogLevel.Set("debug")
		require.NoError(t, srv.ApplyConfig(cfg))
	})

	t.Run("invalid changes", func(t *testing.T) {
		cfg := newTestConfig()

		srv, err := New(nil, nil, nil, cfg)
		require.NoError(t, err)

		cfg.Flags.HTTP.ListenPort = 2
		require.EqualError(t, srv.ApplyConfig(cfg), "cannot dynamically update values for deprecated YAML fields")
	})
}

func TestServer_ListenAddress_Precedence(t *testing.T) {
	// Reserve a port to listen on
	reservedHTTPLis, err := net.Listen("tcp", anyLocalhost)
	require.NoError(t, err)
	defer reservedHTTPLis.Close()

	reservedGRPCLis, err := net.Listen("tcp", anyLocalhost)
	require.NoError(t, err)
	defer reservedGRPCLis.Close()

	// Create a config which sets both ListenAddress and ListenHost/ListenPort.
	// ListenAddress should take precedence. If it doesn't, the port will collide
	// with our existing listeners.
	cfg := DefaultConfig
	cfg.Flags.HTTP.ListenAddress = anyLocalhost
	cfg.Flags.HTTP.ListenHost = "127.0.0.1"
	cfg.Flags.HTTP.ListenPort = reservedHTTPLis.Addr().(*net.TCPAddr).Port
	cfg.Flags.GRPC.ListenAddress = anyLocalhost
	cfg.Flags.GRPC.ListenHost = "127.0.0.1"
	cfg.Flags.GRPC.ListenPort = reservedGRPCLis.Addr().(*net.TCPAddr).Port

	srv, err := New(nil, nil, nil, cfg)
	require.NoError(t, err)
	require.NoError(t, srv.Close()) // Close listeners we just opened

	// Now we want to remove the ListenAddress override and ensure that creating
	// a new server fails.
	cfg.Flags.HTTP.ListenAddress = ""
	cfg.Flags.GRPC.ListenAddress = anyLocalhost
	srv, err = New(nil, nil, nil, cfg)
	require.NotNil(t, err) // The error message is different per platform, so we don't check for the error string here
	require.Nil(t, srv)

	cfg.Flags.HTTP.ListenAddress = anyLocalhost
	cfg.Flags.GRPC.ListenAddress = ""
	srv, err = New(nil, nil, nil, cfg)
	require.NotNil(t, err) // The error message is different per platform, so we don't check for the error string here
	require.Nil(t, srv)
}
