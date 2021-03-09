//+build ignore

package ha

import (
	"context"
	"sort"
	"sync"
	"testing"

	"github.com/cortexproject/cortex/pkg/ring"
	"github.com/go-kit/kit/log"
	"github.com/grafana/agent/pkg/agentproto"
	"github.com/grafana/agent/pkg/prom/instance"
	"github.com/grafana/agent/pkg/prom/instance/configstore"
	"github.com/stretchr/testify/require"
)

func TestServer_Reshard(t *testing.T) {
	// Resharding should do the following:
	//	- All configs in the store should be applied
	//	- All configs not in the store but in the existing InstanceManager should be deleted
	fakeIm := newFakeInstanceManager()

	mockStore := storeWithKeys("keep_a", "keep_b", "new_a", "new_b")

	fakeRing := mockFuncReadRing{
		GetFunc: func(key uint32, op ring.Operation, bufDescs []ring.InstanceDesc, bufHosts, bufZones []string) (ring.ReplicationSet, error) {
			return ring.ReplicationSet{Ingesters: []ring.InstanceDesc{{Addr: "test-server"}}}, nil
		},
	}

	srv := Server{
		logger: log.NewNopLogger(),

		store: mockStore,
		im:    fakeIm,

		ring: &fakeRing,
		addr: "test-server",

		// Pass fake configs that were applied in a previous run. remove_a and remove_b
		// don't exist
		configs: map[string]struct{}{
			"keep_a":   {},
			"keep_b":   {},
			"remove_a": {},
			"remove_b": {},
		},
	}

	_, err := srv.Reshard(context.Background(), &agentproto.ReshardRequest{})
	require.NoError(t, err)

	expect := []string{"keep_a", "keep_b", "new_a", "new_b"}
	var actual []string
	for k := range fakeIm.ListConfigs() {
		actual = append(actual, k)
	}
	sort.Strings(actual)
	require.Equal(t, expect, actual)
}

func TestServer_Ownership(t *testing.T) {
	// Resharding should do the following:
	//	- All configs in the store should be applied
	//	- All configs not in the store but in the existing InstanceManager should be deleted
	fakeIm := newFakeInstanceManager()

	mockStore := storeWithKeys("owned", "unowned")

	var (
		ownedHash = keyHash("owned")
	)

	fakeRing := mockFuncReadRing{
		GetFunc: func(key uint32, op ring.Operation, bufDescs []ring.InstanceDesc, bufHosts, bufZones []string) (ring.ReplicationSet, error) {
			switch key {
			case ownedHash:
				return ring.ReplicationSet{Ingesters: []ring.InstanceDesc{{Addr: "test-server"}}}, nil
			default:
				return ring.ReplicationSet{Ingesters: []ring.InstanceDesc{{Addr: "someone-else"}}}, nil
			}
		},
	}

	srv := Server{
		logger: log.NewNopLogger(),

		store: mockStore,
		im:    fakeIm,

		ring: &fakeRing,
		addr: "test-server",
	}

	_, err := srv.Reshard(context.Background(), &agentproto.ReshardRequest{})
	require.NoError(t, err)

	expect := []string{"owned"}
	var actual []string
	for k := range fakeIm.ListConfigs() {
		actual = append(actual, k)
	}
	sort.Strings(actual)
	require.Equal(t, expect, actual)
}

func storeWithKeys(keys ...string) *configstore.Mock {
	return &configstore.Mock{
		AllFunc: func(ctx context.Context, keep func(key string) bool) (<-chan instance.Config, error) {
			ch := make(chan instance.Config)

			go func() {
				for _, key := range keys {
					if keep(key) {
						ch <- *testConfig(key)
					}
				}
				close(ch)
			}()

			return ch, nil
		},

		WatchFunc: func() <-chan configstore.WatchEvent {
			// Return a watcher that will never emit any events.
			return make(<-chan configstore.WatchEvent)
		},
	}
}

func newFakeInstanceManager() instance.Manager {
	var mut sync.Mutex
	var cfgs = make(map[string]instance.Config)

	return &instance.MockManager{
		// ListInstances isn't used in this package, so we won't bother to try to
		// fake it here.
		ListInstancesFunc: func() map[string]instance.ManagedInstance { return nil },

		ListConfigsFunc: func() map[string]instance.Config {
			mut.Lock()
			defer mut.Unlock()

			cp := make(map[string]instance.Config, len(cfgs))
			for k, v := range cfgs {
				cp[k] = v
			}
			return cp
		},

		ApplyConfigFunc: func(c instance.Config) error {
			mut.Lock()
			defer mut.Unlock()
			cfgs[c.Name] = c
			return nil
		},

		DeleteConfigFunc: func(name string) error {
			mut.Lock()
			defer mut.Unlock()
			delete(cfgs, name)
			return nil
		},

		StopFunc: func() {},
	}
}
