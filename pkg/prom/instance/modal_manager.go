package instance

import (
	"fmt"
	"sync"

	"github.com/go-kit/kit/log"
	"github.com/go-kit/kit/log/level"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

// Mode controls how instances are created.
type Mode string

// Types of instance modes
var (
	ModeDistinct Mode = "distinct"
	ModeShared   Mode = "shared"

	DefaultMode = ModeShared
)

// UnmarshalYAML unmarshals a string to a Mode. Fails if the string is
// unrecognized.
func (m *Mode) UnmarshalYAML(unmarshal func(interface{}) error) error {
	*m = DefaultMode

	var plain string
	if err := unmarshal(&plain); err != nil {
		return err
	}

	switch plain {
	case string(ModeDistinct):
		*m = ModeDistinct
		return nil
	case string(ModeShared):
		*m = ModeShared
		return nil
	default:
		return fmt.Errorf("unsupported instance_mode '%s'. supported values 'shared', 'distinct'", plain)
	}
}

// ModalManager runs instances by either grouping them or running them fully
// separately.
type ModalManager struct {
	mut     sync.RWMutex
	mode    Mode
	configs map[string]Config

	currentActiveConfigs prometheus.Gauge

	log log.Logger

	// Next is the underlying manager this manager wraps.
	next        Manager
	modeManager Manager
}

// NewModalManager creates a new ModalManager.
func NewModalManager(reg prometheus.Registerer, l log.Logger, next Manager, mode Mode) (*ModalManager, error) {
	if mode == "" {
		mode = DefaultMode
	}

	currentActiveConfigs := promauto.With(reg).NewGauge(prometheus.GaugeOpts{
		Name: "agent_prometheus_active_configs",
		Help: "Current number of active configs being used by the agent.",
	})

	mm := ModalManager{
		next:                 next,
		log:                  l,
		currentActiveConfigs: currentActiveConfigs,
		configs:              make(map[string]Config),
	}
	if err := mm.SetMode(mode); err != nil {
		return nil, err
	}
	return &mm, nil
}

// SetMode updates the mode ModalManager is running in. Changing the mode is
// an expensive operation; all underlying configs must be stopped and then
// reapplied.
func (m *ModalManager) SetMode(newMode Mode) error {
	m.mut.Lock()
	defer m.mut.Unlock()

	if m.mode == newMode {
		return nil
	}
	m.mode = newMode

	// Stop the old mode before changing it. It won't exist if this is the first
	// time calling SetMode from NewModalManager.
	if m.modeManager != nil {
		m.modeManager.Stop()
	}

	switch m.mode {
	case ModeDistinct:
		m.modeManager = m.next
	case ModeShared:
		m.modeManager = NewGroupManager(m.next)
	default:
		panic("unknown mode " + m.mode)
	}

	// Re-apply configs to the new mode.
	var firstError error
	for name, cfg := range m.configs {
		err := m.modeManager.ApplyConfig(cfg)
		if err != nil {
			level.Error(m.log).Log("msg", "failed to apply config when changing modes", "name", name, "err", err)
		}
		if firstError == nil && err != nil {
			firstError = err
		}
	}

	return firstError
}

// ListInstances implements Manager.
func (m *ModalManager) ListInstances() map[string]ManagedInstance {
	m.mut.RLock()
	defer m.mut.RUnlock()
	return m.modeManager.ListInstances()
}

// ListConfigs implements Manager.
func (m *ModalManager) ListConfigs() map[string]Config {
	m.mut.RLock()
	defer m.mut.RUnlock()
	return m.modeManager.ListConfigs()
}

// ApplyConfig implements Manager.
func (m *ModalManager) ApplyConfig(c Config) error {
	m.mut.Lock()
	defer m.mut.Unlock()

	if err := m.modeManager.ApplyConfig(c); err != nil {
		return err
	}

	if _, existingConfig := m.configs[c.Name]; !existingConfig {
		m.currentActiveConfigs.Inc()
	}
	m.configs[c.Name] = c

	return nil
}

// DeleteConfig implements Manager.
func (m *ModalManager) DeleteConfig(name string) error {
	m.mut.Lock()
	defer m.mut.Unlock()

	if err := m.modeManager.DeleteConfig(name); err != nil {
		return err
	}

	if _, existingConfig := m.configs[name]; existingConfig {
		m.currentActiveConfigs.Dec()
		delete(m.configs, name)
	}
	return nil
}

// Stop implements Manager.
func (m *ModalManager) Stop() {
	m.mut.Lock()
	defer m.mut.Unlock()

	m.modeManager.Stop()
	m.currentActiveConfigs.Set(0)
	m.configs = make(map[string]Config)
}
