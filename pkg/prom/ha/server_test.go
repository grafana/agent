package ha

import (
	"context"
	"errors"
	"fmt"
	"math"
	"net"
	"net/http"
	"sort"
	"testing"
	"time"

	"github.com/cortexproject/cortex/pkg/ring"
	"github.com/cortexproject/cortex/pkg/util/flagext"
	"github.com/cortexproject/cortex/pkg/util/test"
	"github.com/go-kit/kit/log"
	"github.com/golang/protobuf/ptypes/empty"
	"github.com/grafana/agent/pkg/agentproto"
	"github.com/grafana/agent/pkg/prom/ha/client"
	"github.com/grafana/agent/pkg/prom/instance"
	"github.com/grafana/agent/pkg/prom/instance/configstore"
	"github.com/prometheus/prometheus/config"
	"github.com/stretchr/testify/require"
	"go.uber.org/atomic"
	"google.golang.org/grpc"
)

func TestServer_Reshard_On_Start(t *testing.T) {
	r := &mockFuncReadRing{}
	im := newFakeInstanceManager()

	injectRingIngester(r)

	// Preconfigure some configs for the server to reshard and use.
	mockStore := storeWithKeys("a", "b", "c")

	srv := newTestServer(r, mockStore, im, time.Minute*60)
	defer func() { require.NoError(t, srv.Stop()) }()

	test.Poll(t, time.Second*5, []string{"a", "b", "c"}, func() interface{} {
		return getRunningConfigs(im)
	})
}

func TestServer_Config_Detection(t *testing.T) {
	r := &mockFuncReadRing{}
	im := newFakeInstanceManager()

	store := configstore.Mock{
		AllFunc: func(ctx context.Context, keep func(key string) bool) (<-chan instance.Config, error) {
			return nil, fmt.Errorf("no configs")
		},
		WatchFunc: func() <-chan configstore.WatchEvent {
			ev := make(chan configstore.WatchEvent)
			go func() {
				ev <- configstore.WatchEvent{Key: "unowned", Config: testConfig("unowned")}
				ev <- configstore.WatchEvent{Key: "a", Config: testConfig("a")}
			}()
			return ev
		},
	}

	injectRingIngester(r)

	r.GetFunc = func(key uint32, op ring.Operation, bufDescs []ring.InstanceDesc, bufHosts, bufZones []string) (ring.ReplicationSet, error) {
		if key == keyHash("unowned") {
			return ring.ReplicationSet{}, nil
		}

		return ring.ReplicationSet{
			Ingesters: []ring.InstanceDesc{{Addr: "test"}},
		}, nil
	}

	srv := newTestServer(r, &store, im, time.Minute*60)
	defer func() { require.NoError(t, srv.Stop()) }()

	test.Poll(t, time.Second*5, []string{"a"}, func() interface{} {
		return getRunningConfigs(im)
	})
}

func TestServer_DeletedConfig_Detection(t *testing.T) {
	r := &mockFuncReadRing{}
	im := newFakeInstanceManager()
	injectRingIngester(r)

	watchEv := make(chan configstore.WatchEvent)

	store := &configstore.Mock{
		AllFunc: func(ctx context.Context, keep func(key string) bool) (<-chan instance.Config, error) {
			return nil, fmt.Errorf("no configs")
		},
		WatchFunc: func() <-chan configstore.WatchEvent {
			return watchEv
		},
	}

	srv := newTestServer(r, store, im, time.Minute*60)
	defer func() { require.NoError(t, srv.Stop()) }()

	watchEv <- configstore.WatchEvent{Key: "a", Config: testConfig("a")}
	test.Poll(t, time.Second*5, []string{"a"}, func() interface{} {
		return getRunningConfigs(im)
	})

	watchEv <- configstore.WatchEvent{Key: "a", Config: nil}
	test.Poll(t, time.Second*10, []string{}, func() interface{} {
		return getRunningConfigs(im)
	})
}

func TestServer_Reshard_On_Interval(t *testing.T) {
	r := &mockFuncReadRing{}
	im := newFakeInstanceManager()
	injectRingIngester(r)

	store := storeWithKeys("a")
	srv := newTestServer(r, store, im, time.Millisecond*250)
	defer func() { require.NoError(t, srv.Stop()) }()

	test.Poll(t, time.Second*5, []string{"a"}, func() interface{} {
		return getRunningConfigs(im)
	})
}

func TestServer_Cluster_Reshard_On_Start_And_Leave(t *testing.T) {
	// Create a few fake agent servers that our agent can connect
	// to to tell them to reshard.
	var (
		agent1Resharded = atomic.NewBool(false)
		agent2Resharded = atomic.NewBool(false)
	)

	agent1 := mockFuncAgentProtoServer{}
	agent1.ReshardFunc = func(_ context.Context, _ *agentproto.ReshardRequest) (*empty.Empty, error) {
		agent1Resharded.Store(true)
		return &empty.Empty{}, nil
	}
	agent1Desc := startScrapingServiceServer(t, &agent1)

	agent2 := mockFuncAgentProtoServer{}
	agent2.ReshardFunc = func(_ context.Context, _ *agentproto.ReshardRequest) (*empty.Empty, error) {
		agent2Resharded.Store(true)
		return &empty.Empty{}, nil
	}
	agent2Desc := startScrapingServiceServer(t, &agent2)

	r := &mockFuncReadRing{}
	im := newFakeInstanceManager()

	// Inject the GetFunc to always return the local node but override GetAll to
	// return our custom fake nodes.
	injectRingIngester(r)
	r.GetAllHealthyFunc = func(_ ring.Operation) (ring.ReplicationSet, error) {
		return ring.ReplicationSet{
			Ingesters: []ring.InstanceDesc{
				{Addr: "test"},
				agent1Desc,
				agent2Desc,
			},
		}, nil
	}

	// Launch a local agent. It should connect over gRPC to our fake agents
	// and tell them to reshard on startup and after TransferOut is called.
	srv := newTestServer(r, storeWithKeys(), im, time.Minute*60)

	test.Poll(t, time.Second*5, true, func() interface{} {
		return agent1Resharded.Load() && agent2Resharded.Load()
	})

	// Reset flags for testing shutdown.
	agent1Resharded.Store(false)
	agent2Resharded.Store(false)

	// We're not using a lifecycler so we have to emulate the transfer out by
	// calling it manually here.
	require.NoError(t, srv.Stop())
	require.NoError(t, srv.TransferOut(context.Background()))

	test.Poll(t, time.Second*5, true, func() interface{} {
		return agent1Resharded.Load() && agent2Resharded.Load()
	})
}

func injectRingIngester(r *mockFuncReadRing) {
	r.GetFunc = func(_ uint32, _ ring.Operation, _ []ring.InstanceDesc, _, _ []string) (ring.ReplicationSet, error) {
		return ring.ReplicationSet{
			Ingesters: []ring.InstanceDesc{{Addr: "test"}},
		}, nil
	}

	r.GetAllHealthyFunc = func(_ ring.Operation) (ring.ReplicationSet, error) {
		return ring.ReplicationSet{
			Ingesters: []ring.InstanceDesc{{Addr: "test"}},
		}, nil
	}
}

func newTestServer(r ReadRing, store configstore.Store, im instance.Manager, reshard time.Duration) *Server {
	var (
		cfg          Config
		clientConfig client.Config
	)
	flagext.DefaultValues(&cfg, &clientConfig)
	cfg.ReshardInterval = reshard

	logger := log.NewNopLogger()
	closer := func() error { return nil }

	return newServer(cfg, &config.DefaultGlobalConfig, clientConfig, logger, im, "test", r, store, closer, nil)
}

func getRunningConfigs(im instance.Manager) []string {
	configs := im.ListConfigs()
	configKeys := make([]string, 0, len(configs))
	for n := range configs {
		configKeys = append(configKeys, n)
	}
	sort.Strings(configKeys)
	return configKeys
}

type mockFuncAgentProtoServer struct {
	ReshardFunc func(ctx context.Context, req *agentproto.ReshardRequest) (*empty.Empty, error)
}

func (m mockFuncAgentProtoServer) Reshard(ctx context.Context, req *agentproto.ReshardRequest) (*empty.Empty, error) {
	if m.ReshardFunc != nil {
		return m.ReshardFunc(ctx, req)
	}
	return nil, errors.New("not implemented")
}

// startScrapingServiceServer launches a gRPC server and registers a ScrapingServiceServer
// against it. The ring.InstanceDesc to add to a ring implementation is returned.
//
// The gRPC server will be stopped when the test exits.
func startScrapingServiceServer(t *testing.T, srv agentproto.ScrapingServiceServer) ring.InstanceDesc {
	t.Helper()

	l, err := net.Listen("tcp", "127.0.0.1:0")
	require.NoError(t, err)

	grpcServer := grpc.NewServer()
	agentproto.RegisterScrapingServiceServer(grpcServer, srv)

	go func() {
		_ = grpcServer.Serve(l)
	}()
	t.Cleanup(func() { grpcServer.Stop() })

	return ring.InstanceDesc{
		Addr:      l.Addr().String(),
		State:     ring.ACTIVE,
		Timestamp: math.MaxInt64,
	}
}

type mockFuncReadRing struct {
	http.Handler

	GetFunc           func(key uint32, op ring.Operation, bufDescs []ring.InstanceDesc, bufHosts, bufZones []string) (ring.ReplicationSet, error)
	GetAllHealthyFunc func(ring.Operation) (ring.ReplicationSet, error)
}

func (r *mockFuncReadRing) Get(key uint32, op ring.Operation, bufDescs []ring.InstanceDesc, bufHosts, bufZones []string) (ring.ReplicationSet, error) {
	if r.GetFunc != nil {
		return r.GetFunc(key, op, bufDescs, bufHosts, bufZones)
	}
	return ring.ReplicationSet{}, errors.New("not implemented")
}

func (r *mockFuncReadRing) GetAllHealthy(op ring.Operation) (ring.ReplicationSet, error) {
	if r.GetAllHealthyFunc != nil {
		return r.GetAllHealthyFunc(op)
	}
	return ring.ReplicationSet{}, errors.New("not implemented")
}

func testConfig(name string) *instance.Config {
	cfg := instance.DefaultConfig
	cfg.Name = name
	return &cfg
}
