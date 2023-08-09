package flow

import (
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
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

func RetrieveAgentSeed(fp string) (*AgentSeed, error) {
	if fileExists(fp) {
		return readSeedFile(fp)
	}

	return writeSeedFile(fp)
}

func fileExists(fp string) bool {
	_, err := os.Stat(fp)
	return !errors.Is(err, os.ErrNotExist)
}

func readSeedFile(fp string) (*AgentSeed, error) {
	data, err := os.ReadFile(fp)
	if err != nil {
		return nil, err
	}
	agentSeed := &AgentSeed{}
	err = json.Unmarshal(data, agentSeed)
	return agentSeed, err
}

func writeSeedFile(fp string) (*AgentSeed, error) {
	agentSeed := &AgentSeed{
		UID:       uuid.NewString(),
		Version:   version.Version,
		CreatedAt: time.Now(),
	}

	data, err := json.Marshal(agentSeed)
	if err != nil {
		return nil, err
	}

	err = os.MkdirAll(filepath.Dir(fp), 0755)
	if err != nil {
		return nil, err
	}

	err = os.WriteFile(fp, data, 0644)
	return agentSeed, err
}
