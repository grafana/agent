// Package http contains a River serializable definition of the weaveworks server config in
// https://github.com/weaveworks/common/blob/master/server/server.go#L62.
package http

import (
	"flag"
	"github.com/weaveworks/common/server"
	"time"
)

// ServerConfig is a wrapper around server.ServerConfig.
type ServerConfig struct {
	// HTTP configures the HTTP server. Note that despite the blog being present or not,
	// the server is always started.
	HTTP *HTTPConfig `river:"http,block,optional"`

	// GRPC configures the gRPC server. Note that despite the blog being present or not,
	// the server is always started.
	GRPC *GRPCConfig `river:"grpc,block,optional"`
}

type HTTPConfig struct {
	ListenAddress string `river:"listen_address,attr,optional"`
	ListenPort    int    `river:"listen_port,attr,optional"`
	ConnLimit     int    `river:"conn_limit,attr,optional"`
}

func (h *HTTPConfig) Config(c *server.Config) {
	c.HTTPListenAddress = h.ListenAddress
	c.HTTPListenPort = h.ListenPort
	c.HTTPConnLimit = h.ConnLimit
}

type GRPCConfig struct {
	ListenAddress         string        `river:"listen_address,attr,optional"`
	ListenPort            int           `river:"listen_port,attr,optional"`
	ConnLimit             int           `river:"conn_limit,attr,optional"`
	MaxConnectionAge      time.Duration `river:"max_connection_age,attr,optional"`
	MaxConnectionAgeGrace time.Duration `river:"max_connection_age_grace,attr,optional"`
	MaxConnectionIdle     time.Duration `river:"max_connection_idle,attr,optional"`
}

func (g *GRPCConfig) Config(c *server.Config) {
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

func (c *ServerConfig) Convert() server.Config {
	cfg := newDefaultConfig()
	if c.HTTP != nil {
		c.HTTP.Config(&cfg)
	}
	if c.GRPC != nil {
		c.GRPC.Config(&cfg)
	}
	return cfg
}

func newDefaultConfig() server.Config {
	c := server.Config{}
	c.RegisterFlags(flag.NewFlagSet("empty", flag.ContinueOnError))
	return c
}
