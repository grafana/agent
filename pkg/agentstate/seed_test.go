package agentstate_test

import (
	"encoding/json"
	"os"
	"testing"
	"time"

	"github.com/grafana/agent/pkg/agentstate"
	"github.com/prometheus/common/version"
	"github.com/stretchr/testify/require"
)

func TestInitNewFile(t *testing.T) {
	filepath := t.TempDir() + "/agent_seed_test.json"
	asc := agentstate.NewAgentSeedController(filepath)
	err := asc.Init()
	require.NoError(t, err)

	require.NotEmpty(t, asc.AgentSeed.UID)
	require.NotEmpty(t, asc.AgentSeed.CreatedAt)
	require.Equal(t, version.Version, asc.AgentSeed.Version)
}

func TestInitExistingFile(t *testing.T) {
	filepath := t.TempDir() + "/agent_seed_test.json"
	existingSeed := agentstate.AgentSeed{
		UID:       "existing-uid",
		Version:   "v1.2.3",
		CreatedAt: time.Now().Add(-24 * time.Hour), // 24 hours ago
	}

	existingData, err := json.Marshal(existingSeed)
	require.NoError(t, err)
	err = os.WriteFile(filepath, existingData, 0644)
	require.NoError(t, err)

	asc := agentstate.NewAgentSeedController(filepath)
	err = asc.Init()
	require.NoError(t, err)

	require.Equal(t, existingSeed.UID, asc.AgentSeed.UID)
	require.Equal(t, existingSeed.Version, asc.AgentSeed.Version)
	require.Equal(t, existingSeed.CreatedAt.Format(time.RFC3339), asc.AgentSeed.CreatedAt.Format(time.RFC3339))
}
