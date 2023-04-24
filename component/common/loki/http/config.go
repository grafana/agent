// Package http contains a River serializable definition of the weaveworks server config in
// https://github.com/weaveworks/common/blob/master/server/server.go#L62.
package http

import (
	"flag"
	"time"

	"github.com/weaveworks/common/server"
)

// ServerConfig is a River configuration that allows one to configure a server.Server. It
// exposes a subset of the available configurations.
type ServerConfig struct {
	// HTTP configures the HTTP server. Note that despite the block being present or not,
	// the server is always started.
	HTTP *HTTPConfig `river:"http,block,optional"`

	// GRPC configures the gRPC server. Note that despite the block being present or not,
	// the server is always started.
	GRPC *GRPCConfig `river:"grpc,block,optional"`
}

// HTTPConfig configures the HTTP server started by server.Server.
type HTTPConfig struct {
	ListenAddress string `river:"listen_address,attr,optional"`
	ListenPort    int    `river:"listen_port,attr,optional"`
	ConnLimit     int    `river:"conn_limit,attr,optional"`
}

// Into applies the configs from HTTPConfig into a server.Into.
func (h *HTTPConfig) Into(c *server.Config) {
	c.HTTPListenAddress = h.ListenAddress
	c.HTTPListenPort = h.ListenPort
	c.HTTPConnLimit = h.ConnLimit
}

// GRPCConfig configures the gRPC server started by server.Server.
type GRPCConfig struct {
	ListenAddress         string        `river:"listen_address,attr,optional"`
	ListenPort            int           `river:"listen_port,attr,optional"`
	ConnLimit             int           `river:"conn_limit,attr,optional"`
	MaxConnectionAge      time.Duration `river:"max_connection_age,attr,optional"`
	MaxConnectionAgeGrace time.Duration `river:"max_connection_age_grace,attr,optional"`
	MaxConnectionIdle     time.Duration `river:"max_connection_idle,attr,optional"`
}

// Into applies the configs from GRPCConfig into a server.Into.
func (g *GRPCConfig) Into(c *server.Config) {
	c.GRPCListenAddress = g.ListenAddress
	c.GRPCListenPort = g.ListenPort
	c.GRPCConnLimit = g.ConnLimit
	c.GRPCServerMaxConnectionAge = g.MaxConnectionAge
	c.GRPCServerMaxConnectionAgeGrace = g.MaxConnectionAgeGrace
	c.GRPCServerMaxConnectionIdle = g.MaxConnectionIdle
}

func (c *ServerConfig) UnmarshalRiver(f func(v interface{}) error) error {
	type config ServerConfig
	if err := f((*config)(c)); err != nil {
		return err
	}

	return nil
}

// Convert converts the River-based ServerConfig into a server.Config object.
func (c *ServerConfig) Convert() server.Config {
	cfg := newDefaultConfig()
	if c.HTTP != nil {
		c.HTTP.Into(&cfg)
	}
	if c.GRPC != nil {
		c.GRPC.Into(&cfg)
	}
	return cfg
}

// newDefaultConfig creates a new server.Config object with some overridden defaults.
func newDefaultConfig() server.Config {
	c := server.Config{}
	c.RegisterFlags(flag.NewFlagSet("empty", flag.ContinueOnError))
	// Opting by default 0, which used in net.Listen assigns a random port
	c.HTTPListenPort = 0
	c.GRPCListenPort = 0
	return c
}
