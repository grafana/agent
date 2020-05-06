package ha

import (
	"context"
	"errors"
	"math"
	"net"
	"sort"
	"strings"
	"testing"
	"time"

	"github.com/cortexproject/cortex/pkg/ring"
	"github.com/cortexproject/cortex/pkg/ring/kv"
	"github.com/cortexproject/cortex/pkg/ring/kv/etcd"
	"github.com/cortexproject/cortex/pkg/util/flagext"
	"github.com/cortexproject/cortex/pkg/util/test"
	"github.com/go-kit/kit/log"
	"github.com/golang/protobuf/ptypes/empty"
	"github.com/grafana/agent/pkg/agentproto"
	"github.com/grafana/agent/pkg/prometheus/ha/client"
	"github.com/grafana/agent/pkg/prometheus/instance"
	"github.com/stretchr/testify/require"
	"go.uber.org/atomic"
	"google.golang.org/grpc"
)

func TestServer_Reshard_On_Start(t *testing.T) {
	r := &mockFuncReadRing{}
	cm := newMockConfigManager()

	kv, closer, err := etcd.Mock(GetCodec())
	require.NoError(t, err)
	t.Cleanup(func() { closer.Close() })

	injectRingIngester(r)

	// Preconfigure some configs for the server to reshard and use.
	for _, name := range []string{"a", "b", "c"} {
		err := kv.CAS(context.Background(), name, func(_ interface{}) (interface{}, bool, error) {
			return &instance.Config{Name: name}, false, nil
		})
		require.NoError(t, err)
	}

	srv := newTestServer(r, kv, cm, time.Minute*60)
	defer func() { require.NoError(t, srv.Stop()) }()

	test.Poll(t, time.Second*5, []string{"a", "b", "c"}, func() interface{} {
		return getRunningConfigs(cm)
	})
}

func TestServer_NewConfig_Detection(t *testing.T) {
	r := &mockFuncReadRing{}
	cm := newMockConfigManager()

	kv, closer, err := etcd.Mock(GetCodec())
	require.NoError(t, err)
	t.Cleanup(func() { closer.Close() })

	injectRingIngester(r)

	srv := newTestServer(r, kv, cm, time.Minute*60)
	defer func() { require.NoError(t, srv.Stop()) }()

	// Wait for the server to finish joining before applying a new config.
	test.Poll(t, time.Second*5, true, func() interface{} {
		return srv.joined.Load()
	})
	err = kv.CAS(context.Background(), "a", func(_ interface{}) (interface{}, bool, error) {
		return &instance.Config{Name: "a"}, false, nil
	})
	require.NoError(t, err)

	test.Poll(t, time.Second*5, []string{"a"}, func() interface{} {
		return getRunningConfigs(cm)
	})
}

func TestServer_DeletedConfig_Detection(t *testing.T) {
	r := &mockFuncReadRing{}
	cm := newMockConfigManager()

	kv, closer, err := etcd.Mock(GetCodec())
	require.NoError(t, err)
	t.Cleanup(func() { closer.Close() })

	injectRingIngester(r)

	err = kv.CAS(context.Background(), "a", func(_ interface{}) (interface{}, bool, error) {
		return &instance.Config{Name: "a"}, false, nil
	})
	require.NoError(t, err)

	srv := newTestServer(r, kv, cm, time.Minute*60)
	defer func() { require.NoError(t, srv.Stop()) }()

	// Wait for the server to finish joining before deleting the config.
	test.Poll(t, time.Second*5, true, func() interface{} {
		return srv.joined.Load()
	})
	test.Poll(t, time.Second*5, []string{"a"}, func() interface{} {
		return getRunningConfigs(cm)
	})
	err = kv.Delete(context.Background(), "a")
	require.NoError(t, err)

	test.Poll(t, time.Second*10, []string{}, func() interface{} {
		return getRunningConfigs(cm)
	})
}

func TestServer_Reshard_On_Interval(t *testing.T) {
	r := &mockFuncReadRing{}
	cm := newMockConfigManager()

	// use a poll only KV so watch events aren't detected - we want this
	// test to make sure polling the KV store works for resharding.
	kv := newPollOnlyKV()
	injectRingIngester(r)

	srv := newTestServer(r, kv, cm, time.Millisecond*250)
	defer func() { require.NoError(t, srv.Stop()) }()

	// Wait for the server to finish joining before applying a new config.
	test.Poll(t, time.Second*5, true, func() interface{} {
		return srv.joined.Load()
	})
	err := kv.CAS(context.Background(), "a", func(_ interface{}) (interface{}, bool, error) {
		return &instance.Config{Name: "a"}, false, nil
	})
	require.NoError(t, err)

	test.Poll(t, time.Second*5, []string{"a"}, func() interface{} {
		return getRunningConfigs(cm)
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
	cm := newMockConfigManager()
	kv := newPollOnlyKV()

	// Inject the GetFunc to always return the local node but override GetAll to
	// return our custom fake nodes.
	injectRingIngester(r)
	r.GetAllFunc = func() (ring.ReplicationSet, error) {
		return ring.ReplicationSet{
			Ingesters: []ring.IngesterDesc{
				{Addr: "test"},
				agent1Desc,
				agent2Desc,
			},
		}, nil
	}

	// Launch a local agent. It should connect over gRPC to our fake agents
	// and tell them to reshard on startup and after TransferOut is called.
	srv := newTestServer(r, kv, cm, time.Minute*60)

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
	r.GetFunc = func(_ uint32, _ ring.Operation, _ []ring.IngesterDesc) (ring.ReplicationSet, error) {
		return ring.ReplicationSet{
			Ingesters: []ring.IngesterDesc{{Addr: "test"}},
		}, nil
	}

	r.GetAllFunc = func() (ring.ReplicationSet, error) {
		return ring.ReplicationSet{
			Ingesters: []ring.IngesterDesc{{Addr: "test"}},
		}, nil
	}
}

func newTestServer(r readRing, kv kv.Client, cm ConfigManager, reshard time.Duration) *Server {
	var (
		cfg          Config
		clientConfig client.Config
	)
	flagext.DefaultValues(&cfg, &clientConfig)
	cfg.ReshardInterval = reshard

	logger := log.NewNopLogger()
	closer := func() error { return nil }

	return newServer(cfg, clientConfig, logger, cm, "test", r, kv, closer)
}

func getRunningConfigs(cm ConfigManager) []string {
	configs := cm.ListConfigs()
	configKeys := make([]string, 0, len(configs))
	for n := range configs {
		configKeys = append(configKeys, n)
	}
	sort.Strings(configKeys)
	return configKeys
}

type pollOnlyKV struct {
	keys map[string]interface{}
}

func newPollOnlyKV() *pollOnlyKV {
	return &pollOnlyKV{
		keys: make(map[string]interface{}),
	}
}

func (kv pollOnlyKV) List(ctx context.Context, prefix string) ([]string, error) {
	keys := make([]string, 0, len(kv.keys))
	for k := range kv.keys {
		if strings.HasPrefix(k, prefix) {
			keys = append(keys, k)
		}
	}
	return keys, nil
}

func (kv pollOnlyKV) Get(ctx context.Context, key string) (interface{}, error) {
	return kv.keys[key], nil
}

func (kv pollOnlyKV) Delete(ctx context.Context, key string) error {
	delete(kv.keys, key)
	return nil
}

func (kv pollOnlyKV) CAS(ctx context.Context, key string, f func(in interface{}) (out interface{}, retry bool, err error)) error {
	old := kv.keys[key]
	new, _, err := f(old)
	if err != nil {
		return err
	}
	kv.keys[key] = new
	return nil
}

func (kv pollOnlyKV) WatchKey(ctx context.Context, _ string, _ func(interface{}) bool) {
	// WatchKey does nothing - pollOnlyKV can only be used for polling.
	<-ctx.Done()
}

func (kv pollOnlyKV) WatchPrefix(ctx context.Context, _ string, _ func(string, interface{}) bool) {
	// WatchPrefix does nothing - pollOnlyKV can only be used for polling.
	<-ctx.Done()
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
// against it. The ring.IngesterDesc to add to a ring implementation is returned.
//
// The gRPC server will be stopped when the test exits.
func startScrapingServiceServer(t *testing.T, srv agentproto.ScrapingServiceServer) ring.IngesterDesc {
	t.Helper()

	l, err := net.Listen("tcp", "127.0.0.1:0")
	require.NoError(t, err)

	grpcServer := grpc.NewServer()
	agentproto.RegisterScrapingServiceServer(grpcServer, srv)

	go func() {
		_ = grpcServer.Serve(l)
	}()
	t.Cleanup(func() { grpcServer.Stop() })

	return ring.IngesterDesc{
		Addr:      l.Addr().String(),
		State:     ring.ACTIVE,
		Timestamp: math.MaxInt64,
	}
}
