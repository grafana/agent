package instance

import (
	"fmt"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestGroupManager_ListInstances_Configs(t *testing.T) {
	gm := NewGroupManager(newFakeManager())

	// Create two configs in the same group and one in another
	// group.
	configs := []string{
		`
name: configA
host_filter: false
scrape_configs: []
remote_write: []
wal_truncate_frequency: 1m
remote_flush_deadline: 1m
write_stale_on_shutdown: false`,
		`
name: configB
host_filter: false
scrape_configs: []
remote_write: []
wal_truncate_frequency: 1m
remote_flush_deadline: 1m
write_stale_on_shutdown: false`,
		`
name: configC
host_filter: false
scrape_configs: []
remote_write: 
- url: http://localhost:9090
wal_truncate_frequency: 1m
remote_flush_deadline: 1m
write_stale_on_shutdown: false`,
	}

	for _, cfg := range configs {
		c := testUnmarshalConfig(t, cfg)
		err := gm.ApplyConfig(c)
		require.NoError(t, err)
	}

	// ListInstances should return our grouped instances
	insts := gm.ListInstances()
	require.Equal(t, 2, len(insts))

	// ...but ListConfigs should return the ungrouped configs.
	confs := gm.ListConfigs()
	require.Equal(t, 3, len(confs))
	require.Containsf(t, confs, "configA", "configA not in confs")
	require.Containsf(t, confs, "configB", "configB not in confs")
	require.Containsf(t, confs, "configC", "configC not in confs")
}

func testUnmarshalConfig(t *testing.T, cfg string) Config {
	c, err := UnmarshalConfig(strings.NewReader(cfg))
	require.NoError(t, err)
	return *c
}

func TestGroupManager_ApplyConfig(t *testing.T) {
	t.Run("updating existing config within group", func(t *testing.T) {
		inner := newFakeManager()
		gm := NewGroupManager(inner)
		err := gm.ApplyConfig(testUnmarshalConfig(t, `
name: configA
host_filter: false
scrape_configs: []
remote_write: []
wal_truncate_frequency: 1m
remote_flush_deadline: 1m
write_stale_on_shutdown: false
`))
		require.NoError(t, err)
		require.Equal(t, 1, len(gm.groups))
		require.Equal(t, 1, len(gm.groupLookup))

		err = gm.ApplyConfig(testUnmarshalConfig(t, `
name: configA
host_filter: false
scrape_configs: 
- job_name: test_job 
  static_configs:
    - targets: [127.0.0.1:12345]
remote_write: []
wal_truncate_frequency: 1m
remote_flush_deadline: 1m
write_stale_on_shutdown: false
`))
		require.NoError(t, err)
		require.Equal(t, 1, len(gm.groups))
		require.Equal(t, 1, len(gm.groupLookup))

		// Check the underlying grouped config and make sure it was updated.
		expect := testUnmarshalConfig(t, fmt.Sprintf(`
name: %s
host_filter: false
scrape_configs:
- job_name: test_job
  static_configs:
    - targets: [127.0.0.1:12345]
remote_write: []
wal_truncate_frequency: 1m
remote_flush_deadline: 1m
write_stale_on_shutdown: false
`, gm.groupLookup["configA"]))
		actual := inner.ListConfigs()[gm.groupLookup["configA"]]
		require.Equal(t, expect, actual)
	})
}

func newFakeManager() Manager {
	instances := make(map[string]ManagedInstance)
	configs := make(map[string]Config)

	return &MockManager{
		ListInstancesFunc: func() map[string]ManagedInstance {
			return instances
		},
		ListConfigsFunc: func() map[string]Config {
			return configs
		},
		ApplyConfigFunc: func(c Config) error {
			instances[c.Name] = &mockInstance{}
			configs[c.Name] = c
			return nil
		},
		DeleteConfigFunc: func(name string) error {
			delete(instances, name)
			delete(configs, name)
			return nil
		},
		StopFunc: func() {},
	}
}

func Test_hashConfig(t *testing.T) {
	t.Run("name and scrape configs are ignored", func(t *testing.T) {
		configAText := `
name: configA
host_filter: false
scrape_configs: []
remote_write: []
wal_truncate_frequency: 1m
remote_flush_deadline: 1m
write_stale_on_shutdown: false`

		configBText := `
name: configB
host_filter: false
scrape_configs: 
- job_name: test_job 
  static_configs:
    - targets: [127.0.0.1:12345]
remote_write: []
wal_truncate_frequency: 1m
remote_flush_deadline: 1m
write_stale_on_shutdown: false`

		hashA, hashB := getHashesFromConfigs(t, configAText, configBText)
		require.Equal(t, hashA, hashB)
	})

	t.Run("remote_writes are unordered", func(t *testing.T) {
		configAText := `
name: configA
host_filter: false
scrape_configs: []
remote_write: 
- url: http://localhost:9009/api/prom/push1
- url: http://localhost:9009/api/prom/push2
wal_truncate_frequency: 1m
remote_flush_deadline: 1m
write_stale_on_shutdown: false`

		configBText := `
name: configB
host_filter: false
scrape_configs: []
remote_write: 
- url: http://localhost:9009/api/prom/push2
- url: http://localhost:9009/api/prom/push1
wal_truncate_frequency: 1m
remote_flush_deadline: 1m
write_stale_on_shutdown: false`

		hashA, hashB := getHashesFromConfigs(t, configAText, configBText)
		require.Equal(t, hashA, hashB)
	})

	t.Run("remote_writes must match", func(t *testing.T) {
		configAText := `
name: configA
host_filter: false
scrape_configs: []
remote_write: 
- url: http://localhost:9009/api/prom/push1
- url: http://localhost:9009/api/prom/push2
wal_truncate_frequency: 1m
remote_flush_deadline: 1m
write_stale_on_shutdown: false`

		configBText := `
name: configB
host_filter: false
scrape_configs: []
remote_write: 
- url: http://localhost:9009/api/prom/push1
- url: http://localhost:9009/api/prom/push1
wal_truncate_frequency: 1m
remote_flush_deadline: 1m
write_stale_on_shutdown: false`

		hashA, hashB := getHashesFromConfigs(t, configAText, configBText)
		require.NotEqual(t, hashA, hashB)
	})

	t.Run("other fields must match", func(t *testing.T) {
		configAText := `
name: configA
host_filter: true
scrape_configs: []
remote_write: []
wal_truncate_frequency: 1m
remote_flush_deadline: 1m
write_stale_on_shutdown: false`

		configBText := `
name: configB
host_filter: false
scrape_configs: []
remote_write: []
wal_truncate_frequency: 1m
remote_flush_deadline: 1m
write_stale_on_shutdown: false`

		hashA, hashB := getHashesFromConfigs(t, configAText, configBText)
		require.NotEqual(t, hashA, hashB)
	})
}

func getHashesFromConfigs(t *testing.T, configAText, configBText string) (string, string) {
	configA := testUnmarshalConfig(t, configAText)
	configB := testUnmarshalConfig(t, configBText)

	hashA, err := hashConfig(configA)
	require.NoError(t, err)

	hashB, err := hashConfig(configB)
	require.NoError(t, err)

	return hashA, hashB
}

func Test_groupConfigs(t *testing.T) {
	configAText := `
name: configA
host_filter: false
scrape_configs: 
- job_name: test_job 
  static_configs:
    - targets: [127.0.0.1:12345]
remote_write: 
- url: http://localhost:9009/api/prom/push1
- url: http://localhost:9009/api/prom/push2
wal_truncate_frequency: 1m
remote_flush_deadline: 1m
write_stale_on_shutdown: false`

	configBText := `
name: configB
host_filter: false
scrape_configs: 
- job_name: test_job2
  static_configs:
    - targets: [127.0.0.1:12345]
remote_write: 
- url: http://localhost:9009/api/prom/push2
- url: http://localhost:9009/api/prom/push1
wal_truncate_frequency: 1m
remote_flush_deadline: 1m
write_stale_on_shutdown: false`

	configA := testUnmarshalConfig(t, configAText)
	configB := testUnmarshalConfig(t, configBText)

	groupName, err := hashConfig(configA)
	require.NoError(t, err)

	expectText := fmt.Sprintf(`
name: %s
host_filter: false
scrape_configs: 
- job_name: test_job 
  static_configs:
    - targets: [127.0.0.1:12345]
- job_name: test_job2
  static_configs:
    - targets: [127.0.0.1:12345]
remote_write: 
- url: http://localhost:9009/api/prom/push1
- url: http://localhost:9009/api/prom/push2
wal_truncate_frequency: 1m
remote_flush_deadline: 1m
write_stale_on_shutdown: false`, groupName)

	expect, err := UnmarshalConfig(strings.NewReader(expectText))
	require.NoError(t, err)

	group := groupedConfigs{
		"configA": configA,
		"configB": configB,
	}
	actual, err := groupConfigs(groupName, group)
	require.NoError(t, err)
	require.Equal(t, *expect, actual)

	// Consistency check: groupedConfigs is a map and we want to always have
	// groupConfigs return the same thing regardless of how the map
	// is iterated over. Run through groupConfigs a bunch of times and
	// make sure it always returns the same thing.
	for i := 0; i < 100; i++ {
		actual, err = groupConfigs(groupName, group)
		require.NoError(t, err)
		require.Equal(t, *expect, actual)
	}
}
