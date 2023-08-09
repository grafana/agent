package flow

import (
	"encoding/json"
	"errors"
	"os"
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

func RetrieveAgentSeed(filepath string) (*AgentSeed, error) {
	if fileExists(filepath) {
		return readSeedFile(filepath)
	}

	return writeSeedFile(filepath)
}

func fileExists(path string) bool {
	_, err := os.Stat(path)
	return !errors.Is(err, os.ErrNotExist)
}

func readSeedFile(filepath string) (*AgentSeed, error) {
	data, err := os.ReadFile(filepath)
	if err != nil {
		return nil, err
	}
	agentSeed := &AgentSeed{}
	err = json.Unmarshal(data, agentSeed)
	return agentSeed, err
}

func writeSeedFile(filepath string) (*AgentSeed, error) {
	agentSeed := &AgentSeed{
		UID:       uuid.NewString(),
		Version:   version.Version,
		CreatedAt: time.Now(),
	}

	data, err := json.Marshal(agentSeed)
	if err != nil {
		return nil, err
	}

	err = os.WriteFile(filepath, data, 0644)
	return agentSeed, err
}
