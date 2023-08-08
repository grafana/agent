package agentstate

import (
	"encoding/json"
	"errors"
	"os"
	"time"

	"github.com/google/uuid"
	"github.com/prometheus/common/version"
)

type AgentSeedController struct {
	AgentSeed *AgentSeed
	filepath  string
}

// AgentSeed identifies a unique agent
type AgentSeed struct {
	UID       string    `json:"UID"`
	CreatedAt time.Time `json:"created_at"`
	Version   string    `json:"version"`
}

func NewAgentSeedController(filePath string) *AgentSeedController {
	return &AgentSeedController{
		filepath:  filePath,
		AgentSeed: &AgentSeed{},
	}
}

func (asc *AgentSeedController) Init() error {
	if fileExists(asc.filepath) {
		return asc.readSeedFile()
	}

	asc.AgentSeed = &AgentSeed{
		UID:       uuid.NewString(),
		Version:   version.Version,
		CreatedAt: time.Now(),
	}

	return asc.writeSeedFile()
}

func fileExists(path string) bool {
	_, err := os.Stat(path)
	return !errors.Is(err, os.ErrNotExist)
}

func (asc *AgentSeedController) readSeedFile() error {
	data, err := os.ReadFile(asc.filepath)
	if err != nil {
		return err
	}
	asc.AgentSeed = &AgentSeed{}
	return json.Unmarshal(data, asc.AgentSeed)
}

func (asc *AgentSeedController) writeSeedFile() error {
	data, err := json.Marshal(asc.AgentSeed)
	if err != nil {
		return err
	}
	return os.WriteFile(asc.filepath, data, 0644)
}
