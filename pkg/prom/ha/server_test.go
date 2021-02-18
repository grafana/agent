package ha

import (
	"context"
	"errors"
	"math"
	"net"
	"sort"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/cortexproject/cortex/pkg/ring"
	"github.com/cortexproject/cortex/pkg/ring/kv"
	"github.com/cortexproject/cortex/pkg/util/flagext"
	"github.com/cortexproject/cortex/pkg/util/test"
	"github.com/go-kit/kit/log"
	"github.com/golang/protobuf/ptypes/empty"
	"github.com/grafana/agent/pkg/agentproto"
	"github.com/grafana/agent/pkg/prom/ha/client"
	"github.com/grafana/agent/pkg/prom/instance"
	"github.com/prometheus/prometheus/config"
	"github.com/stretchr/testify/require"
	"go.uber.org/atomic"
	"google.golang.org/grpc"
)

func TestServer_Reshard_On_Start(t *testing.T) {
	r := &mockFuncReadRing{}
	im := newFakeInstanceManager()

	kv := newMockKV(true)
	injectRingIngester(r)

	// Preconfigure some configs for the server to reshard and use.
	for _, name := range []string{"a", "b", "c"} {
		err := kv.CAS(context.Background(), name, func(_ interface{}) (interface{}, bool, error) {
			return &instance.Config{Name: name}, false, nil
		})
		require.NoError(t, err)
	}

	srv := newTestServer(r, kv, im, time.Minute*60)
	defer func() { require.NoError(t, srv.Stop()) }()

	test.Poll(t, time.Second*5, []string{"a", "b", "c"}, func() interface{} {
		return getRunningConfigs(im)
	})
}

func TestServer_NewConfig_Detection(t *testing.T) {
	r := &mockFuncReadRing{}
	im := newFakeInstanceManager()

	kv := newMockKV(true)
	injectRingIngester(r)

	srv := newTestServer(r, kv, im, time.Minute*60)
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
		return getRunningConfigs(im)
	})
}

func TestServer_DeletedConfig_Detection(t *testing.T) {
	r := &mockFuncReadRing{}
	im := newFakeInstanceManager()

	kv := newMockKV(true)
	injectRingIngester(r)

	err := kv.CAS(context.Background(), "a", func(_ interface{}) (interface{}, bool, error) {
		return &instance.Config{Name: "a"}, false, nil
	})
	require.NoError(t, err)

	srv := newTestServer(r, kv, im, time.Minute*60)
	defer func() { require.NoError(t, srv.Stop()) }()

	// Wait for the server to finish joining before deleting the config.
	test.Poll(t, time.Second*5, true, func() interface{} {
		return srv.joined.Load()
	})
	test.Poll(t, time.Second*5, []string{"a"}, func() interface{} {
		return getRunningConfigs(im)
	})
	err = kv.Delete(context.Background(), "a")
	require.NoError(t, err)

	test.Poll(t, time.Second*10, []string{}, func() interface{} {
		return getRunningConfigs(im)
	})
}

func TestServer_Reshard_On_Interval(t *testing.T) {
	r := &mockFuncReadRing{}
	im := newFakeInstanceManager()

	// use a poll only KV so watch events aren't detected - we want this
	// test to make sure polling the KV store works for resharding.
	kv := newMockKV(false)
	injectRingIngester(r)

	srv := newTestServer(r, kv, im, time.Millisecond*250)
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
	kv := newMockKV(false)

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
	srv := newTestServer(r, kv, im, time.Minute*60)

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

func newTestServer(r ReadRing, kv kv.Client, im instance.Manager, reshard time.Duration) *Server {
	var (
		cfg          Config
		clientConfig client.Config
	)
	flagext.DefaultValues(&cfg, &clientConfig)
	cfg.ReshardInterval = reshard

	logger := log.NewNopLogger()
	closer := func() error { return nil }

	return newServer(cfg, &config.DefaultGlobalConfig, clientConfig, logger, im, "test", r, kv, closer)
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

type mockKV struct {
	keys       map[string]interface{}
	allowWatch bool

	mut  *sync.Mutex
	cond *sync.Cond
}

func newMockKV(allowWatch bool) *mockKV {
	kv := &mockKV{
		keys:       make(map[string]interface{}),
		allowWatch: allowWatch,
		mut:        &sync.Mutex{},
	}
	kv.cond = sync.NewCond(kv.mut)
	return kv
}

func (kv *mockKV) List(ctx context.Context, prefix string) ([]string, error) {
	kv.mut.Lock()
	defer kv.mut.Unlock()

	keys := make([]string, 0, len(kv.keys))
	for k := range kv.keys {
		if strings.HasPrefix(k, prefix) {
			keys = append(keys, k)
		}
	}
	return keys, nil
}

func (kv *mockKV) Get(ctx context.Context, key string) (interface{}, error) {
	kv.mut.Lock()
	defer kv.mut.Unlock()

	return kv.keys[key], nil
}

func (kv *mockKV) Delete(ctx context.Context, key string) error {
	kv.mut.Lock()
	defer kv.mut.Unlock()

	delete(kv.keys, key)
	kv.cond.Broadcast()
	return nil
}

func (kv *mockKV) CAS(ctx context.Context, key string, f func(in interface{}) (out interface{}, retry bool, err error)) error {
	kv.mut.Lock()
	defer kv.mut.Unlock()

	old := kv.keys[key]
	new, _, err := f(old)
	if err != nil {
		return err
	}
	kv.keys[key] = new

	kv.cond.Broadcast()
	return nil
}

func (kv *mockKV) WatchKey(ctx context.Context, key string, f func(interface{}) bool) {
	if !kv.allowWatch {
		// Do nothing, watching isn't allowed
		<-ctx.Done()
		return
	}

	// When the context is canceled, wake up all watchers so they can check to see
	// if they need to exit.
	go func() {
		<-ctx.Done()
		kv.cond.Broadcast()
	}()

	kv.cond.L.Lock()

	// Retrieve the key's initial value before waiting.
	prev := kv.keys[key]

	for {
		// Wait for a key to change. If the key we're watching has a new value,
		// call f and then save it for the next wakeup.
		kv.cond.Wait()
		if v := kv.keys[key]; v != prev {
			_ = f(v)
			prev = v
		}
		kv.cond.L.Unlock()

		// We might've been woken up by the context being canceled, check that.
		if ctx.Err() != nil {
			break
		}

		// Relock the mutex for the next wait cycle
		kv.mut.Lock()
	}
}

func (kv *mockKV) WatchPrefix(ctx context.Context, prefix string, f func(string, interface{}) bool) {
	if !kv.allowWatch {
		// Do nothing, watching isn't allowed
		<-ctx.Done()
		return
	}

	// When the context is canceled, wake up all watchers so they can check to see
	// if they need to exit.
	go func() {
		<-ctx.Done()
		kv.cond.Broadcast()
	}()

	kv.mut.Lock()

	// Retrieve the initial value for everything within the prefix
	// before waiting.
	cache := map[string]interface{}{}
	for k, v := range kv.keys {
		if strings.HasPrefix(k, prefix) {
			cache[k] = v
		}
	}

	for {
		// Wait for a key to change. If any of the keys in the prefix we're watching
		// has a new value, call f and save its value for the next wakeup.
		kv.cond.Wait()
		for k, v := range kv.keys {
			if cached, ok := cache[k]; ok && cached != v {
				_ = f(k, v)
				cache[k] = v
			} else if !ok && strings.HasPrefix(k, prefix) {
				// New value to watch
				_ = f(k, v)
				cache[k] = v
			}
		}
		// Check for deleted keys
		for k := range cache {
			if _, exist := kv.keys[k]; !exist {
				_ = f(k, nil)
				delete(cache, k)
			}
		}
		kv.mut.Unlock()

		// We might've been woken up by the context being canceled, check that.
		if ctx.Err() != nil {
			break
		}

		// Relock the mutex for the next wait cycle
		kv.mut.Lock()
	}
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
