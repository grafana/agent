package ha

import (
	"context"
	"sort"
	"testing"

	"github.com/cortexproject/cortex/pkg/ring"
	"github.com/cortexproject/cortex/pkg/ring/kv/consul"
	"github.com/go-kit/kit/log"
	"github.com/grafana/agent/pkg/agentproto"
	"github.com/grafana/agent/pkg/prom/instance"
	"github.com/stretchr/testify/require"
)

func TestServer_Reshard(t *testing.T) {
	// Resharding should do the following:
	//	- All configs in the store should be applied
	//	- All configs not in the store but in the existing InstanceManager should be deleted
	fakeIm := newFakeInstanceManager()

	mockKv := consul.NewInMemoryClient(GetCodec())
	for _, name := range []string{"keep_a", "keep_b", "new_a", "new_b"} {
		err := mockKv.CAS(context.Background(), name, func(in interface{}) (out interface{}, retry bool, err error) {
			return &instance.Config{Name: name}, true, nil
		})
		require.NoError(t, err)
	}

	fakeRing := mockFuncReadRing{
		GetFunc: func(key uint32, op ring.Operation, bufDescs []ring.InstanceDesc, bufHosts, bufZones []string) (ring.ReplicationSet, error) {
			return ring.ReplicationSet{Ingesters: []ring.InstanceDesc{{Addr: "test-server"}}}, nil
		},
	}

	srv := Server{
		logger: log.NewNopLogger(),

		kv: mockKv,
		im: fakeIm,

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
