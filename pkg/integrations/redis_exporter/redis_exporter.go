// package redis_exporter embeds https://github.com/oliver006/redis_exporter
package redis_exporter //nolint:golint

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"

	"github.com/go-kit/kit/log"
	"github.com/go-kit/kit/log/level"
	"github.com/gorilla/mux"

	re "github.com/oliver006/redis_exporter/lib/exporter"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"

	"github.com/grafana/agent/pkg/integrations/config"
)

// Integration is the redis_exporter integration. The integration queries
// a redis instance's INFO and exposes the results as metrics.
type Integration struct {
	c        Config
	exporter *re.Exporter
}

// New creates a new redis_exporter integration.
func New(log log.Logger, c Config) (*Integration, error) {
	level.Debug(log).Log("msg", "initialising redis_exporer with config %v", c)

	address := c.RedisAddr
	if len(address) == 0 {
		address = os.Getenv("REDIS_EXPORTER_ADDRESS")
	}
	if len(address) == 0 {
		return nil, fmt.Errorf("cannot create redis_exporter; neither redis_exporter.redis_addr or $REDIS_EXPORTER_ADDRESS is set")
	}
	level.Debug(log).Log("msg", "Redis exporter address is %s", address)

	exporterConfig := c.GetExporterOptions()

	if c.ScriptPath != "" {
		ls, err := ioutil.ReadFile(c.ScriptPath)
		if err != nil {
			return nil, fmt.Errorf("Error loading script file %s    err: %s", c.ScriptPath, err)
		}
		exporterConfig.LuaScript = ls
	}

	var tlsClientCertificates []tls.Certificate
	if (c.TLSClientKeyFile != "") != (c.TLSClientCertFile != "") {
		return nil, fmt.Errorf("TLS client key file and cert file should both be present")
	}
	if c.TLSClientKeyFile != "" && c.TLSClientCertFile != "" {
		cert, err := tls.LoadX509KeyPair(c.TLSClientCertFile, c.TLSClientKeyFile)
		if err != nil {

			return nil, fmt.Errorf("couldn't load TLS client key pair, err: %s", err)
		}
		tlsClientCertificates = append(tlsClientCertificates, cert)
	}
	exporterConfig.ClientCertificates = tlsClientCertificates

	var tlsCaCertificates *x509.CertPool
	if c.TLSCaCertFile != "" {
		caCert, err := ioutil.ReadFile(c.TLSCaCertFile)
		if err != nil {
			return nil, fmt.Errorf("couldn't load TLS Ca certificate, err: %s", err)
		}
		tlsCaCertificates = x509.NewCertPool()
		tlsCaCertificates.AppendCertsFromPEM(caCert)
	}
	exporterConfig.CaCertificates = tlsCaCertificates

	// optional password file to take precedence over password property
	if c.RedisPasswordFile != "" {
		password, err := ioutil.ReadFile(c.RedisPasswordFile)
		if err != nil {
			return nil, fmt.Errorf("Error loading password file %s: %w", c.RedisPasswordFile, err)
		}
		exporterConfig.Password = string(password)
	}

	exporter, err := re.NewRedisExporter(
		address,
		exporterConfig,
		re.BuildInfo{},
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create redis exporter: %s", err)
	}

	return &Integration{
		c:        c,
		exporter: exporter,
	}, nil
}

// CommonConfig satisfies Integration.CommonConfig.
func (i *Integration) CommonConfig() config.Common { return i.c.CommonConfig }

// Name satisfies Integration.Name.
func (i *Integration) Name() string { return "redis_exporter" }

// RegisterRoutes satisfies Integration.RegisterRoutes. The mux.Router provided
// here is expected to be a subrouter, where all registered paths will be
// registered within that subroute.
func (i *Integration) RegisterRoutes(r *mux.Router) error {
	handler, err := i.handler()
	if err != nil {
		return err
	}

	r.Handle("/metrics", handler)
	return nil
}

func (i *Integration) handler() (http.Handler, error) {
	r := prometheus.NewRegistry()
	if err := r.Register(i.exporter); err != nil {
		return nil, fmt.Errorf("couldn't register redis_exporter: %w", err)
	}

	handler := promhttp.HandlerFor(
		r,
		promhttp.HandlerOpts{
			ErrorHandling: promhttp.ContinueOnError,
		},
	)

	if i.c.IncludeExporterMetrics {
		// Note that we have to use reg here to use the same promhttp metrics for
		// all expositions.
		handler = promhttp.InstrumentMetricHandler(r, handler)
	}

	return handler, nil
}

// ScrapeConfigs satisfies Integration.ScrapeConfigs.
func (i *Integration) ScrapeConfigs() []config.ScrapeConfig {
	return []config.ScrapeConfig{{
		JobName:     i.Name(),
		MetricsPath: "/metrics",
	}}
}

// Run satisfies Integration.Run.
func (i *Integration) Run(ctx context.Context) error {
	// We don't need to do anything here, so we can just wait for the context to
	// finish.
	<-ctx.Done()
	return ctx.Err()
}
