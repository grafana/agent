package agentseed

import (
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"runtime"
	"time"

	"github.com/google/uuid"
	"github.com/prometheus/common/version"
)

// AgentSeed identifies a unique agent
type AgentSeed struct {
	UID       string    `json:"UID"`
	CreatedAt time.Time `json:"created_at"`
	Version   string    `json:"version"`
}

const filename = "agent_seed.json"

var savedSeed *AgentSeed

// Get will return a unique uuid for this agent.
// Seed will be saved in agent_seed.json
// If path is not empty, that will be the "preferred" place to read and save it.
// If it is empty, we will fall back to $APPDATA on windows or /tmp on *nix systems.
func Get(path string) (seed *AgentSeed, err error) {
	if savedSeed != nil {
		return savedSeed, nil
	}
	defer func() {
		if err == nil && seed != nil {
			savedSeed = seed
		}
	}()
	paths := []string{}
	if path != "" {
		paths = append(paths, filepath.Join(path, filename))
	}
	paths = append(paths, legacyPath())
	for i, p := range paths {
		if fileExists(p) {
			if seed, err = readSeedFile(p); err == nil {
				if i == 0 {
					// we found it at the preferred path. Just return it
					return seed, err
				} else {
					return seed, writeSeedFile(seed, paths[0])
				}
			}
		}
	}
	seed = &AgentSeed{
		UID:       uuid.NewString(),
		Version:   version.Version,
		CreatedAt: time.Now(),
	}
	return seed, writeSeedFile(seed, paths[0])
}

// readSeedFile reads the agent seed file
func readSeedFile(path string) (*AgentSeed, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	seed := &AgentSeed{}
	err = json.Unmarshal(data, seed)
	if err != nil {
		return nil, err
	}
	return seed, nil
}

func legacyPath() string {
	// windows
	if runtime.GOOS == "windows" {
		return filepath.Join(os.Getenv("APPDATA"), filename)
	}
	// linux/mac
	return filepath.Join("/tmp", filename)
}

func fileExists(path string) bool {
	_, err := os.Stat(path)
	return !errors.Is(err, os.ErrNotExist)
}

// writeSeedFile writes the agent seed file
func writeSeedFile(seed *AgentSeed, path string) error {
	data, err := json.Marshal(*seed)
	if err != nil {
		return err
	}
	return os.WriteFile(path, data, 0644)
}
