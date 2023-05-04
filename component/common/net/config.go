// Package http contains a River serializable definition of the weaveworks weaveworks config in
// https://github.com/weaveworks/common/blob/master/server/server.go#L62.
package net

import (
	"time"

	"github.com/gorilla/mux"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/weaveworks/common/logging"
	"github.com/weaveworks/common/middleware"
)

const (
	DefaultHTTPPort = 8080
)

// ServerConfig for a Server
type ServerConfig struct {

	// ==== configuration exposed in River configs ===

	// HTTP configures the HTTP weaveworks. Note that despite the block being present or not,
	// the weaveworks is always started.
	HTTP *HTTPConfig `river:"http,block,optional"`

	// GracefulShutdownTimeout configures a timeout to gracefully shut down the server.
	GracefulShutdownTimeout time.Duration `river:"graceful_shutdown_timeout,attr,optional"`

	// ==== configuration NOT exposed in River configs ===

	MetricsNamespace  string
	HTTPListenNetwork string

	CipherSuites  string
	MinVersion    string
	HTTPTLSConfig TLSConfig

	RegisterInstrumentation  bool
	ExcludeRequestInLog      bool
	DisableRequestSuccessLog bool

	HTTPMiddleware                []middleware.Interface
	Router                        *mux.Router
	DoNotAddDefaultHTTPMiddleware bool

	LogFormat             logging.Format
	LogLevel              logging.Level
	Log                   logging.Interface
	LogSourceIPs          bool
	LogSourceIPsHeader    string
	LogSourceIPsRegex     string
	LogRequestAtInfoLevel bool

	// If not set, default signal handler is used.
	SignalHandler SignalHandler

	// If not set, default Prometheus registry is used.
	Registerer prometheus.Registerer
	Gatherer   prometheus.Gatherer

	PathPrefix string
}

// HTTPConfig configures the HTTP weaveworks started by weaveworks.Server.
type HTTPConfig struct {
	ListenAddress      string        `river:"listen_address,attr,optional"`
	ListenPort         int           `river:"listen_port,attr,optional"`
	ConnLimit          int           `river:"conn_limit,attr,optional"`
	ServerReadTimeout  time.Duration `river:"server_read_timeout,attr,optional"`
	ServerWriteTimeout time.Duration `river:"server_write_timeout,attr,optional"`
	ServerIdleTimeout  time.Duration `river:"server_idle_timeout,attr,optional"`
}

// TLSConfig contains TLS parameters for ServerConfig.
type TLSConfig struct {
	TLSCertPath string
	TLSKeyPath  string
	ClientAuth  string
	ClientCAs   string
}

func (c *ServerConfig) UnmarshalRiver(f func(v interface{}) error) error {
	type config ServerConfig
	if err := f((*config)(c)); err != nil {
		return err
	}

	return nil
}
