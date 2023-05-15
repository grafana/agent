// Package http contains a River serializable definition of the weaveworks weaveworks config in
// https://github.com/weaveworks/common/blob/master/server/server.go#L62.
package net

import (
	"flag"
	"math"
	"time"

	weaveworks "github.com/weaveworks/common/server"
)

const (
	DefaultHTTPPort = 8080
	DefaultGRPCPort = 8081

	// defaults inherited from weaveworks
	DurationInfinity = time.Duration(math.MaxInt64)
	Size4MB          = 4 << 20
)

// ServerConfig is a River configuration that allows one to configure a weaveworks.Server. It
// exposes a subset of the available configurations.
type ServerConfig struct {
	// HTTP configures the HTTP weaveworks. Note that despite the block being present or not,
	// the weaveworks is always started.
	HTTP *HTTPConfig `river:"http,block,optional"`

	// GRPC configures the gRPC weaveworks. Note that despite the block being present or not,
	// the weaveworks is always started.
	GRPC *GRPCConfig `river:"grpc,block,optional"`

	// GracefulShutdownTimeout configures a timeout to gracefully shut down the server.
	GracefulShutdownTimeout time.Duration `river:"graceful_shutdown_timeout,attr,optional"`
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

// Into applies the configs from HTTPConfig into a weaveworks.Into.
func (h *HTTPConfig) Into(c *weaveworks.Config) {
	c.HTTPListenAddress = h.ListenAddress
	c.HTTPListenPort = h.ListenPort
	c.HTTPConnLimit = h.ConnLimit
	c.HTTPServerReadTimeout = h.ServerReadTimeout
	c.HTTPServerWriteTimeout = h.ServerWriteTimeout
	c.HTTPServerIdleTimeout = h.ServerIdleTimeout
}

// GRPCConfig configures the gRPC weaveworks started by weaveworks.Server.
type GRPCConfig struct {
	ListenAddress              string        `river:"listen_address,attr,optional"`
	ListenPort                 int           `river:"listen_port,attr,optional"`
	ConnLimit                  int           `river:"conn_limit,attr,optional"`
	MaxConnectionAge           time.Duration `river:"max_connection_age,attr,optional"`
	MaxConnectionAgeGrace      time.Duration `river:"max_connection_age_grace,attr,optional"`
	MaxConnectionIdle          time.Duration `river:"max_connection_idle,attr,optional"`
	ServerMaxRecvMsg           int           `river:"server_max_recv_msg_size,attr,optional"`
	ServerMaxSendMsg           int           `river:"server_max_send_msg_size,attr,optional"`
	ServerMaxConcurrentStreams uint          `river:"server_max_concurrent_streams,attr,optional"`
}

// Into applies the configs from GRPCConfig into a weaveworks.Into.
func (g *GRPCConfig) Into(c *weaveworks.Config) {
	c.GRPCListenAddress = g.ListenAddress
	c.GRPCListenPort = g.ListenPort
	c.GRPCConnLimit = g.ConnLimit
	c.GRPCServerMaxConnectionAge = g.MaxConnectionAge
	c.GRPCServerMaxConnectionAgeGrace = g.MaxConnectionAgeGrace
	c.GRPCServerMaxConnectionIdle = g.MaxConnectionIdle
	c.GPRCServerMaxRecvMsgSize = g.ServerMaxRecvMsg
	c.GRPCServerMaxSendMsgSize = g.ServerMaxSendMsg
	c.GPRCServerMaxConcurrentStreams = g.ServerMaxConcurrentStreams
}

func (c *ServerConfig) UnmarshalRiver(f func(v interface{}) error) error {
	*c = newDefaults()
	type config ServerConfig
	if err := f((*config)(c)); err != nil {
		return err
	}

	return nil
}

// Convert converts the River-based ServerConfig into a weaveworks.Config object.
func (c *ServerConfig) Convert() weaveworks.Config {
	cfg := newWeaveworksDefaultConfig()
	if c.HTTP != nil {
		c.HTTP.Into(&cfg)
	}
	if c.GRPC != nil {
		c.GRPC.Into(&cfg)
	}
	cfg.ServerGracefulShutdownTimeout = c.GracefulShutdownTimeout
	return cfg
}

// newWeaveworksDefaultConfig creates a new weaveworks.Config object with some overridden defaults.
func newWeaveworksDefaultConfig() weaveworks.Config {
	c := weaveworks.Config{}
	c.RegisterFlags(flag.NewFlagSet("empty", flag.ContinueOnError))
	// By default, do not register instrumentation since every metric is later registered
	// inside a custom register
	c.RegisterInstrumentation = false
	return c
}

// newDefaults creates a new ServerConfig with defaults applied. Not that some are inherited from
// weaveworks, but copied in our config model to make the overriding logic simpler.
func newDefaults() ServerConfig {
	return ServerConfig{
		HTTP: &HTTPConfig{
			ListenAddress:      "",
			ListenPort:         DefaultHTTPPort,
			ConnLimit:          0,
			ServerReadTimeout:  30 * time.Second,
			ServerWriteTimeout: 30 * time.Second,
			ServerIdleTimeout:  120 * time.Second,
		},
		GRPC: &GRPCConfig{
			ListenAddress:              "",
			ListenPort:                 DefaultGRPCPort,
			ConnLimit:                  0,
			MaxConnectionAge:           DurationInfinity,
			MaxConnectionAgeGrace:      DurationInfinity,
			MaxConnectionIdle:          DurationInfinity,
			ServerMaxConcurrentStreams: 100,
			ServerMaxSendMsg:           Size4MB,
			ServerMaxRecvMsg:           Size4MB,
		},
		GracefulShutdownTimeout: 30 * time.Second,
	}
}
