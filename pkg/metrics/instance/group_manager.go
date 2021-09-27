package instance

import (
	"crypto/md5"
	"encoding/hex"
	"errors"
	"fmt"

	"github.com/go-kit/kit/log"
	"github.com/go-kit/kit/log/level"

	"sort"
	"sync"

	"github.com/prometheus/prometheus/config"
)

// A GroupManager wraps around another Manager and groups all incoming Configs
// into a smaller set of configs, causing less managed instances to be spawned.
//
// Configs are grouped by all settings for a Config *except* scrape configs.
// Any difference found in any flag will cause a Config to be placed in another
// group. One exception to this rule is that remote_writes are compared
// unordered, but the sets of remote_writes should otherwise be identical.
//
// GroupManagers drastically improve the performance of the Agent when a
// significant number of instances are spawned, as the overhead of each
// instance having its own service discovery, WAL, and remote_write can be
// significant.
//
// The config names of instances within the group will be represented by
// that group's hash of settings.
type GroupManager struct {
	inner Manager

	mtx sync.Mutex

	// activeConfigs is a list of all the configs currently being used in the group
	activeConfigs []Config

	// groups is a map of group name to the grouped configs.
	groups map[string]groupedConfigs

	// groupLookup is a map of config name to group name.
	groupLookup map[string]string

	mergedConfigHashes map[string]int

	log log.Logger
}

// ApplyConfigs is used to batch configurations for performance instead of the singular ApplyConfig
func (m *GroupManager) ApplyConfigs(configs []Config) error {
	m.mtx.Lock()
	defer m.mtx.Unlock()
	return m.applyConfigs(configs, false)
}

// applyConfigs is used to take the currently running set of configs, plus the new ones. (update/append configs).
// isRollback determines whether we are rolling back any changes and if so avoid recursion in calling rollback again
func (m *GroupManager) applyConfigs(configs []Config, isRollback bool) error {
	if len(configs) == 0 {
		return nil
	}
	failed := make([]BatchFailure, 0)

	// This will form our master set of configurations to apply by their group, this includes any that are currently
	// running but not in the configs parameter.
	groupsInConfigs := make(map[string]groupedConfigs)

	// Combined group of current and new configs
	combinedConfigs := m.getCombinedConfigs(configs)

	oldConfigs := make([]Config, 0)
	// In case of an error in applying the new configs, we need to grab the old config for rollback
	copy(oldConfigs, m.activeConfigs)
	activeConfigs := make([]Config, 0)
	groupLookup := make(map[string]string)
	// Iterate through the combined configurations
	for _, c := range combinedConfigs {
		groupName, err := hashConfig(c)
		if err != nil {
			failed = append(failed, BatchFailure{
				Err:    err,
				Config: c,
			})
			level.Error(m.log).Log("err", fmt.Sprintf("failed to get group name for config %s: %s", c.Name, err))
			continue
		}
		if _, exists := groupsInConfigs[groupName]; !exists {
			groupsInConfigs[groupName] = make(groupedConfigs)
		}
		groupsInConfigs[groupName][c.Name] = c
		groupLookup[c.Name] = groupName
		activeConfigs = append(activeConfigs, c)
	}
	// If we have group that no longer exist then we need to delete them
	for groupName := range m.groups {
		if _, exists := groupsInConfigs[groupName]; !exists {
			if err := m.inner.DeleteConfig(groupName); err != nil {
				level.Error(m.log).Log("err", fmt.Sprintf("failed to delete group named  %s with error %s", groupName, err))
			}
		}
	}
	groups := make(map[string]groupedConfigs)
	mergedHashes := make(map[string]int)
	// Now that we have grouped all the new and existing configurations we can apply them
	for groupName, grouped := range groupsInConfigs {
		mergedConfig, err := groupConfigs(groupName, grouped)
		if err != nil {
			level.Error(m.log).Log("err", fmt.Sprintf("failed to group configs with groupname  %s with error %s", groupName, err))
		}
		// If we have the exact same config no need to reload it
		newHash, err := createMergedConfigHash(mergedConfig)
		if err != nil {
			level.Error(m.log).Log("err", fmt.Sprintf("failed to create merged hash named  %s with error %s", groupName, err))
			continue
		}
		groups[groupName] = grouped
		mergedHashes[newHash] = 0
		if _, exists := m.mergedConfigHashes[newHash]; exists {
			continue
		}

		// Something terrible happened so roll back to the old configurations, this ensures that at least the agent
		// always has a valid configuration. If we are already in a rollback then dont try again
		if !isRollback {
			m.rollbackConfigs(oldConfigs)
			return CreateBatchApplyErrorOrNil(failed, err)
		}
		err = m.inner.ApplyConfig(mergedConfig)
		if err != nil && !isRollback {
			m.rollbackConfigs(oldConfigs)
			return CreateBatchApplyErrorOrNil(failed, err)
		}
	}
	m.groups = groups
	m.activeConfigs = activeConfigs
	m.groupLookup = groupLookup
	m.mergedConfigHashes = mergedHashes

	return CreateBatchApplyErrorOrNil(failed, nil)
}

func createMergedConfigHash(mergedConfig Config) (string, error) {
	bb, err := MarshalConfig(&mergedConfig, false)
	if err != nil {
		return "", err
	}
	hash := md5.Sum(bb)
	return hex.EncodeToString(hash[:]), nil
}

func (m *GroupManager) rollbackConfigs(oldConfigs []Config) {
	// If restoring a config fails, we've left the Agent in a really bad
	// state: the new config can't be applied and the old config can't be
	// brought back. Just crash and let the Agent start fresh.
	//
	// Restoring the config _shouldn't_ fail here since applies only fail
	// if the config is invalid. Since the config was running before, it
	// should already be valid. If it does happen to fail, though, the
	// internal state is left corrupted since we've completely lost a
	// config.
	if err := m.applyConfigs(oldConfigs, true); err != nil {
		level.Error(m.log).Log("err", fmt.Sprintf("failed to rollback configs with error %s", err))
		panic(err)
	}
}

func (m *GroupManager) getCombinedConfigs(newConfigs []Config) map[string]Config {
	// There is also a known issue in that this DOES not handle deleted configs
	combinedConfigsMap := make(map[string]Config)

	// Preload all our existing configs
	for _, ac := range m.activeConfigs {
		combinedConfigsMap[ac.Name] = ac
	}
	for _, nc := range newConfigs {
		// If new config shares the name with an active config we assume they are the same config and should
		// use the new config
		combinedConfigsMap[nc.Name] = nc
	}
	return combinedConfigsMap
}

// groupedConfigs holds a set of grouped configs, keyed by the config name.
// They are stored in a map rather than a slice to make overriding an existing
// config within the group less error prone.
type groupedConfigs map[string]Config

// Copy returns a shallow copy of the groupedConfigs.
func (g groupedConfigs) Copy() groupedConfigs {
	res := make(groupedConfigs, len(g))
	for k, v := range g {
		res[k] = v
	}
	return res
}

// NewGroupManager creates a new GroupManager for combining instances of the
// same "group."
func NewGroupManager(log log.Logger, inner Manager) *GroupManager {
	return &GroupManager{
		inner:              inner,
		groups:             make(map[string]groupedConfigs),
		groupLookup:        make(map[string]string),
		mergedConfigHashes: make(map[string]int),
		log:                log,
	}
}

// GetInstance gets the underlying grouped instance for a given name.
func (m *GroupManager) GetInstance(name string) (ManagedInstance, error) {
	m.mtx.Lock()
	defer m.mtx.Unlock()

	group, ok := m.groupLookup[name]
	if !ok {
		return nil, fmt.Errorf("instance %s does not exist", name)
	}

	inst, err := m.inner.GetInstance(group)
	if err != nil {
		return nil, fmt.Errorf("failed to get instance for %s: %w", name, err)
	}
	return inst, nil
}

// ListInstances returns all currently grouped managed instances. The key
// will be the group's hash of shared settings.
func (m *GroupManager) ListInstances() map[string]ManagedInstance {
	return m.inner.ListInstances()
}

// ListConfigs returns the UNGROUPED instance configs with their original
// settings. To see the grouped instances, call ListInstances instead.
func (m *GroupManager) ListConfigs() map[string]Config {
	m.mtx.Lock()
	defer m.mtx.Unlock()

	cfgs := make(map[string]Config)
	for _, groupedConfigs := range m.groups {
		for _, cfg := range groupedConfigs {
			cfgs[cfg.Name] = cfg
		}
	}
	return cfgs
}

// ApplyConfig will determine the group of the Config before applying it to
// the group. If no group exists, one will be created. If a group already
// exists, the group will have its settings merged with the Config and
// will be updated.
func (m *GroupManager) ApplyConfig(c Config) error {
	cfgs := []Config{c}
	err := m.ApplyConfigs(cfgs)
	var bae BatchApplyError

	if err != nil && errors.As(err, &bae) {
		if len(bae.Failed) > 0 {
			return bae.Failed[0].Err
		}
	}

	return nil
}

// DeleteConfig will remove a Config from its associated group. If there are
// no more Configs within that group after this Config is deleted, the managed
// instance will be stopped. Otherwise, the managed instance will be updated
// with the new grouped Config that doesn't include the removed one.
func (m *GroupManager) DeleteConfig(name string) error {
	m.mtx.Lock()
	defer m.mtx.Unlock()
	return m.deleteConfig(name)
}

func (m *GroupManager) deleteConfig(name string) error {
	groupName, ok := m.groupLookup[name]
	if !ok {
		return fmt.Errorf("config does not exist")
	}

	// Grab a copy of the stored group and delete our entry. We can
	// persist it after we successfully remove the config.
	group := m.groups[groupName].Copy()
	delete(group, name)

	if len(group) == 0 {
		// We deleted the last remaining config in that group; we can delete it in
		// its entirety now.
		if err := m.inner.DeleteConfig(groupName); err != nil {
			return fmt.Errorf("failed to delete empty group %s after removing config %s: %w", groupName, name, err)
		}
	} else {
		// We deleted the config but there's still more in the group; apply the new
		// group that holds the remainder of the configs (minus the one we just
		// deleted).
		mergedConfig, err := groupConfigs(groupName, group)
		if err != nil {
			return fmt.Errorf("failed to regroup configs without %s: %w", name, err)
		}

		err = m.inner.ApplyConfig(mergedConfig)
		if err != nil {
			return fmt.Errorf("failed to apply new group without %s: %w", name, err)
		}
	}

	// Update the stored group and remove the entry from the lookup table.
	if len(group) == 0 {
		delete(m.groups, groupName)
	} else {
		m.groups[groupName] = group
	}

	delete(m.groupLookup, name)
	return nil
}

// Stop stops the Manager and all of its managed instances.
func (m *GroupManager) Stop() {
	m.mtx.Lock()
	defer m.mtx.Unlock()

	m.inner.Stop()
	m.groupLookup = make(map[string]string)
	m.groups = make(map[string]groupedConfigs)
}

// hashConfig determines the hash of a Config used for grouping. It ignores
// the name and scrape_configs and also orders remote_writes by name prior to
// hashing.
func hashConfig(c Config) (string, error) {
	// We need a deep copy since we're going to mutate the remote_write
	// pointers.
	groupable, err := c.Clone()
	if err != nil {
		return "", err
	}

	// Ignore name and scrape configs when hashing
	groupable.Name = ""
	groupable.ScrapeConfigs = nil

	// Assign names to remote_write configs if they're not present already.
	// This is also done in AssignDefaults but is duplicated here for the sake
	// of simplifying responsibility of GroupManager.
	for _, cfg := range groupable.RemoteWrite {
		if cfg != nil {
			// We don't care if the names are different, just that the other settings
			// are the same. Blank out the name here before hashing the remote
			// write config.
			cfg.Name = ""

			hash, err := getHash(cfg)
			if err != nil {
				return "", err
			}
			cfg.Name = hash[:6]
		}
	}

	// Now sort remote_writes by name and nil-ness.
	sort.Slice(groupable.RemoteWrite, func(i, j int) bool {
		switch {
		case groupable.RemoteWrite[i] == nil:
			return true
		case groupable.RemoteWrite[j] == nil:
			return false
		default:
			return groupable.RemoteWrite[i].Name < groupable.RemoteWrite[j].Name
		}
	})

	bb, err := MarshalConfig(&groupable, false)
	if err != nil {
		return "", err
	}
	hash := md5.Sum(bb)
	return hex.EncodeToString(hash[:]), nil
}

// groupConfig creates a grouped Config where all fields are copied from
// the first config except for scrape_configs, which are appended together.
func groupConfigs(groupName string, grouped groupedConfigs) (Config, error) {
	if len(grouped) == 0 {
		return Config{}, fmt.Errorf("no configs")
	}

	// Move the map into a slice and sort it by name so this function
	// consistently does the same thing.
	cfgs := make([]Config, 0, len(grouped))
	for _, cfg := range grouped {
		cfgs = append(cfgs, cfg)
	}
	sort.Slice(cfgs, func(i, j int) bool { return cfgs[i].Name < cfgs[j].Name })

	combined, err := cfgs[0].Clone()
	if err != nil {
		return Config{}, err
	}
	combined.Name = groupName
	combined.ScrapeConfigs = []*config.ScrapeConfig{}

	// Assign all remote_write configs in the group a consistent set of remote_names.
	// If the grouped configs are coming from the scraping service, defaults will have
	// been applied and the remote names will be prefixed with the old instance config name.
	for _, rwc := range combined.RemoteWrite {
		// Blank out the existing name before getting the hash so it is doesn't take into
		// account any existing name.
		rwc.Name = ""

		hash, err := getHash(rwc)
		if err != nil {
			return Config{}, err
		}

		rwc.Name = groupName[:6] + "-" + hash[:6]
	}

	// Combine all the scrape configs. It's possible that two different ungrouped
	// configs had a matching job name, but this will be detected and rejected
	// (as it should be) when the underlying Manager eventually validates the
	// combined config.
	//
	// TODO(rfratto): should we prepend job names with the name of the original
	// config? (e.g., job_name = "config_name/job_name").
	for _, cfg := range cfgs {
		combined.ScrapeConfigs = append(combined.ScrapeConfigs, cfg.ScrapeConfigs...)
	}

	return combined, nil
}
