package client

import (
	"flag"
	"io"

	"github.com/cortexproject/cortex/pkg/util/grpcclient"

	"github.com/grafana/agent/pkg/agentproto"
	otgrpc "github.com/opentracing-contrib/go-grpc"
	"github.com/opentracing/opentracing-go"
	"github.com/weaveworks/common/middleware"
	"google.golang.org/grpc"
)

// ScrapingServiceClient wraps agentproto.ScrapingServiceClient with a Close method.
type ScrapingServiceClient interface {
	agentproto.ScrapingServiceClient
	io.Closer
}

// Config controls how scraping service clients are created.
type Config struct {
	GRPCClientConfig grpcclient.Config `yaml:"grpc_client_config"`
}

// RegisterFlags registers flags to the provided flag set.
func (c *Config) RegisterFlags(f *flag.FlagSet) {
	c.GRPCClientConfig.RegisterFlags("prometheus.service-client", f)
}

// New returns a new scraping service client.
func New(cfg Config, addr string) (ScrapingServiceClient, error) {
	opts := []grpc.DialOption{
		grpc.WithInsecure(),
		grpc.WithDefaultCallOptions(cfg.GRPCClientConfig.CallOptions()...),
	}
	opts = append(opts, cfg.GRPCClientConfig.DialOption(instrumentation())...)
	conn, err := grpc.Dial(addr, opts...)
	if err != nil {
		return nil, err
	}

	return struct {
		agentproto.ScrapingServiceClient
		io.Closer
	}{
		ScrapingServiceClient: agentproto.NewScrapingServiceClient(conn),
		Closer:                conn,
	}, nil
}

func instrumentation() ([]grpc.UnaryClientInterceptor, []grpc.StreamClientInterceptor) {
	unary := []grpc.UnaryClientInterceptor{
		otgrpc.OpenTracingClientInterceptor(opentracing.GlobalTracer()),
		middleware.ClientUserHeaderInterceptor,
	}
	stream := []grpc.StreamClientInterceptor{
		otgrpc.OpenTracingStreamClientInterceptor(opentracing.GlobalTracer()),
		middleware.StreamClientUserHeaderInterceptor,
	}
	return unary, stream
}
