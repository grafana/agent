package ha

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"sort"
	"testing"

	"github.com/cortexproject/cortex/pkg/ring"
	"github.com/cortexproject/cortex/pkg/ring/kv/consul"
	"github.com/go-kit/kit/log"
	"github.com/grafana/agent/pkg/agentproto"
	"github.com/grafana/agent/pkg/prometheus/instance"
	"github.com/stretchr/testify/require"
	"gopkg.in/yaml.v2"
)

func TestServer_Reshard(t *testing.T) {
	// Resharding should do the following:
	//	- All configs in the store should be applied
	//	- All configs not in the store but in the existing ConfigManager should be deleted
	mockCm := newMockConfigManager()
	for _, name := range []string{"keep_a", "keep_b", "remove_a", "remove_b"} {
		mockCm.ApplyConfig(instance.Config{Name: name})
	}

	mockKv := consul.NewInMemoryClient(GetCodec())
	for _, name := range []string{"keep_a", "keep_b", "new_a", "new_b"} {
		err := mockKv.CAS(context.Background(), name, func(in interface{}) (out interface{}, retry bool, err error) {
			return &instance.Config{Name: name}, true, nil
		})
		require.NoError(t, err)
	}

	srv := Server{kv: mockKv, cm: mockCm}
	_, err := srv.Reshard(context.Background(), &agentproto.ReshardRequest{})
	require.NoError(t, err)

	expect := []string{
		"keep_a",
		"keep_b",
		"new_a",
		"new_b",
	}
	var actual []string
	for k := range mockCm.ListConfigs() {
		actual = append(actual, k)
	}
	sort.Strings(actual)
	require.Equal(t, expect, actual)
}

func TestShardingConfigManager(t *testing.T) {
	logger := log.NewNopLogger()

	t.Run("applies owned config", func(t *testing.T) {
		mockCm := newMockConfigManager()
		mockRing := &mockFuncReadRing{}
		mockRing.GetFunc = func(_ uint32, _ ring.Operation, _ []ring.IngesterDesc) (ring.ReplicationSet, error) {
			return ring.ReplicationSet{
				Ingesters: []ring.IngesterDesc{{Addr: "same_machine"}},
			}, nil
		}

		cm := NewShardingConfigManager(logger, mockCm, mockRing, "same_machine")
		cm.ApplyConfig(instance.Config{Name: "test"})

		require.Equal(t, 1, len(mockCm.cfgs))
	})

	t.Run("ignores apply of unowned config", func(t *testing.T) {
		mockCm := newMockConfigManager()
		mockRing := &mockFuncReadRing{}
		mockRing.GetFunc = func(_ uint32, _ ring.Operation, _ []ring.IngesterDesc) (ring.ReplicationSet, error) {
			return ring.ReplicationSet{
				Ingesters: []ring.IngesterDesc{{Addr: "remote"}},
			}, nil
		}

		cm := NewShardingConfigManager(logger, mockCm, mockRing, "same_machine")
		cm.ApplyConfig(instance.Config{Name: "test"})

		require.Equal(t, 0, len(mockCm.cfgs))
	})

	t.Run("properly hashes config", func(t *testing.T) {
		var hashes []uint32

		mockCm := newMockConfigManager()
		mockRing := &mockFuncReadRing{}
		mockRing.GetFunc = func(key uint32, _ ring.Operation, _ []ring.IngesterDesc) (ring.ReplicationSet, error) {
			hashes = append(hashes, key)
			return ring.ReplicationSet{
				Ingesters: []ring.IngesterDesc{{Addr: "remote"}},
			}, nil
		}

		// Each config here should be given a different hash when checked against the ring
		cm := NewShardingConfigManager(logger, mockCm, mockRing, "same_machine")
		cm.ApplyConfig(instance.Config{Name: "test1"})
		cm.ApplyConfig(instance.Config{Name: "test2"})

		require.Len(t, hashes, 2)
		require.NotEqual(t, hashes[0], hashes[1])
	})

	t.Run("deletes previously owned config on apply", func(t *testing.T) {
		returnRingAddr := "same_machine"

		mockCm := newMockConfigManager()
		mockRing := &mockFuncReadRing{}
		mockRing.GetFunc = func(_ uint32, _ ring.Operation, _ []ring.IngesterDesc) (ring.ReplicationSet, error) {
			return ring.ReplicationSet{
				Ingesters: []ring.IngesterDesc{{Addr: returnRingAddr}},
			}, nil
		}

		cm := NewShardingConfigManager(logger, mockCm, mockRing, "same_machine")
		cm.ApplyConfig(instance.Config{Name: "test"})
		require.Equal(t, 1, len(mockCm.cfgs))

		// Pretend the ring changed and that ring doesn't hash to us anymore.
		// The next apply should delete it.
		returnRingAddr = "not_localhost"

		cm.ApplyConfig(instance.Config{Name: "test"})
		require.Equal(t, 0, len(mockCm.cfgs), "unowned config was not deleted")
	})

	t.Run("doesn't reapply unchanged config", func(t *testing.T) {
		mockCm := newMockConfigManager()
		mockRing := &mockFuncReadRing{}
		mockRing.GetFunc = func(_ uint32, _ ring.Operation, _ []ring.IngesterDesc) (ring.ReplicationSet, error) {
			return ring.ReplicationSet{
				Ingesters: []ring.IngesterDesc{{Addr: "same_machine"}},
			}, nil
		}

		cm := NewShardingConfigManager(logger, mockCm, mockRing, "same_machine")
		cm.ApplyConfig(instance.Config{Name: "test"})

		require.Equal(t, 1, len(mockCm.cfgs))

		// Internally delete the config and try to reapply; our wrapper should ignore
		// it since the hash hasn't changed from the last time it was applied.
		delete(mockCm.cfgs, "test")
		cm.ApplyConfig(instance.Config{Name: "test"})
		require.Equal(t, 0, len(mockCm.cfgs), "unchanged config got reapplied")
	})

	t.Run("reapplies changed config", func(t *testing.T) {
		mockCm := newMockConfigManager()
		mockRing := &mockFuncReadRing{}
		mockRing.GetFunc = func(_ uint32, _ ring.Operation, _ []ring.IngesterDesc) (ring.ReplicationSet, error) {
			return ring.ReplicationSet{
				Ingesters: []ring.IngesterDesc{{Addr: "same_machine"}},
			}, nil
		}

		cm := NewShardingConfigManager(logger, mockCm, mockRing, "same_machine")
		cm.ApplyConfig(instance.Config{Name: "test"})

		require.Equal(t, 1, len(mockCm.cfgs))

		cm.ApplyConfig(instance.Config{Name: "test", HostFilter: true})
		require.Equal(t, 1, len(mockCm.cfgs))
		require.True(t, mockCm.cfgs["test"].HostFilter)
	})

	t.Run("ignores deletes of unowned config", func(t *testing.T) {
		mockCm := newMockConfigManager()
		mockRing := &mockFuncReadRing{}
		cm := NewShardingConfigManager(logger, mockCm, mockRing, "same_machine")

		mockCm.ApplyConfig(instance.Config{Name: "test"})
		err := cm.DeleteConfig("test")
		require.NoError(t, err)

		require.Equal(t, 1, len(mockCm.cfgs), "untracked config was not ignored")
	})

	t.Run("deletes owned config", func(t *testing.T) {
		mockCm := newMockConfigManager()
		mockRing := &mockFuncReadRing{}
		mockRing.GetFunc = func(_ uint32, _ ring.Operation, _ []ring.IngesterDesc) (ring.ReplicationSet, error) {
			return ring.ReplicationSet{
				Ingesters: []ring.IngesterDesc{{Addr: "same_machine"}},
			}, nil
		}

		cm := NewShardingConfigManager(logger, mockCm, mockRing, "same_machine")
		cm.ApplyConfig(instance.Config{Name: "test"})

		err := cm.DeleteConfig("test")
		require.NoError(t, err)
		require.Equal(t, 0, len(mockCm.cfgs), "owned config was not deleted")
	})
}

func TestConfigHash_Secrets_BasicAuth(t *testing.T) {
	configTemplate := `name: 'test'
host_filter: false
scrape_configs:
  - job_name: process-1
    static_configs:
      - targets: ['process-1:80']
        labels:
          cluster: 'local'
          origin: 'agent'
remote_write:
  - name: test-abcdef
    url: http://cortex:9090/api/prom/push
    basic_auth:
      username: test_username
      password: %s`

	configA := fmt.Sprintf(configTemplate, "password_a")
	configB := fmt.Sprintf(configTemplate, "password_b")

	var inA instance.Config
	err := yaml.Unmarshal([]byte(configA), &inA)
	require.NoError(t, err)

	hashA, err := configHash(&inA)
	require.NoError(t, err)

	var inB instance.Config
	err = yaml.Unmarshal([]byte(configB), &inB)
	require.NoError(t, err)

	hashB, err := configHash(&inB)
	require.NoError(t, err)

	require.NotEqual(t, hashA, hashB, "secrets were not hashed separately")
}

func TestConfigHash_Secrets_BearerToken(t *testing.T) {
	configTemplate := `name: 'test'
host_filter: false
scrape_configs:
  - job_name: process-1
    static_configs:
      - targets: ['process-1:80']
        labels:
          cluster: 'local'
          origin: 'agent'
remote_write:
  - name: test-abcdef
    url: http://cortex:9090/api/prom/push
    bearer_token: %s`

	configA := fmt.Sprintf(configTemplate, "bearer_a")
	configB := fmt.Sprintf(configTemplate, "bearer_b")

	var inA instance.Config
	err := yaml.Unmarshal([]byte(configA), &inA)
	require.NoError(t, err)

	hashA, err := configHash(&inA)
	require.NoError(t, err)

	var inB instance.Config
	err = yaml.Unmarshal([]byte(configB), &inB)
	require.NoError(t, err)

	hashB, err := configHash(&inB)
	require.NoError(t, err)

	require.NotEqual(t, hashA, hashB, "secrets were not hashed separately")
}

type mockFuncReadRing struct {
	http.Handler

	GetFunc    func(key uint32, op ring.Operation, buf []ring.IngesterDesc) (ring.ReplicationSet, error)
	GetAllFunc func() (ring.ReplicationSet, error)
}

func (r *mockFuncReadRing) Get(key uint32, op ring.Operation, buf []ring.IngesterDesc) (ring.ReplicationSet, error) {
	if r.GetFunc != nil {
		return r.GetFunc(key, op, buf)
	}
	return ring.ReplicationSet{}, errors.New("not implemented")
}

func (r *mockFuncReadRing) GetAll() (ring.ReplicationSet, error) {
	if r.GetAllFunc != nil {
		return r.GetAllFunc()
	}
	return ring.ReplicationSet{}, errors.New("not implemented")
}
