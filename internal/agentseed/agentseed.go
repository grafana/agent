package agentseed

import (
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"runtime"
	"time"

	"github.com/go-kit/log"
	"github.com/google/uuid"
	"github.com/grafana/agent/pkg/flow/logging/level"
	"github.com/prometheus/common/version"
)

// AgentSeed identifies a unique agent
type AgentSeed struct {
	UID       string    `json:"UID"`
	CreatedAt time.Time `json:"created_at"`
	Version   string    `json:"version"`
}

const filename = "agent_seed.json"

var dataDir = ""
var logger log.Logger

var savedSeed *AgentSeed

// Init should be called by an app entrypoint as soon as it can to configure where the unique seed will be stored.
// dir is the directory where we will read and store agent_seed.json
// If left empty it will default to $APPDATA or /tmp
func Init(dir string, log log.Logger) {
	dataDir = dir
	logger = log
}

// Get will return a unique uuid for this agent.
// Seed will be saved in agent_seed.json
// If path is not empty, that will be the "preferred" place to read and save it.
// If it is empty, we will fall back to $APPDATA on windows or /tmp on *nix systems to read the file.
func Get() (seed *AgentSeed) {
	// TODO: This will just log errors and always return a valid seed.
	// If we wanted to have it return an error for some reason, we could change this api
	// worst case, we generate a new seed if we can't read/write files, and it is only good for the lifetime
	// of this agent.
	if savedSeed != nil {
		return savedSeed
	}
	var err error
	// list of paths in preference order.
	// we will always write to the first path
	paths := []string{}
	if dataDir != "" {
		paths = append(paths, filepath.Join(dataDir, filename))
	}
	defer func() {
		// as a fallback, gen and save a new uid
		if seed == nil || seed.UID == "" {
			seed = &AgentSeed{
				UID:       uuid.NewString(),
				Version:   version.Version,
				CreatedAt: time.Now(),
			}
			writeSeedFile(seed, paths[0])
		}
		// cache seed for future calls
		savedSeed = seed
	}()
	paths = append(paths, legacyPath())
	for i, p := range paths {
		if fileExists(p) {
			if seed, err = readSeedFile(p); err == nil {
				if i == 0 {
					// we found it at the preferred path. Just return it
					return seed
				} else {
					writeSeedFile(seed, paths[0])
					return seed
				}
			}
		}
	}

	return seed
}

// readSeedFile reads the agent seed file
func readSeedFile(path string) (*AgentSeed, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		level.Error(logger).Log("msg", "Reading seed file", "err", err)
		return nil, err
	}
	seed := &AgentSeed{}
	err = json.Unmarshal(data, seed)
	if err != nil {
		level.Error(logger).Log("msg", "Decoding seed file", "err", err)
		return nil, err
	}
	if seed.UID == "" {
		level.Error(logger).Log("msg", "Seed file has empty uid")
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
		level.Error(logger).Log("msg", "Encoding seed file", "err", err)
		return err
	}
	err = os.WriteFile(path, data, 0644)
	if err != nil {
		level.Error(logger).Log("msg", "Writing seed file", "err", err)
		return err
	}
	return err
}
