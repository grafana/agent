package flow_test

import (
	"encoding/json"
	"os"
	"testing"
	"time"

	"github.com/grafana/agent/pkg/flow"
	"github.com/prometheus/common/version"
	"github.com/stretchr/testify/require"
)

func TestInitNewFile(t *testing.T) {
	filepath := t.TempDir() + "/agent_seed_test.json"
	as, err := flow.RetrieveAgentSeed(filepath)
	require.NoError(t, err)

	require.NotEmpty(t, as.UID)
	require.NotEmpty(t, as.CreatedAt)
	require.Equal(t, version.Version, as.Version)
}

func TestInitExistingFile(t *testing.T) {
	filepath := t.TempDir() + "/agent_seed_test.json"
	existingSeed := flow.AgentSeed{
		UID:       "existing-uid",
		Version:   "v1.2.3",
		CreatedAt: time.Now().Add(-24 * time.Hour), // 24 hours ago
	}

	existingData, err := json.Marshal(existingSeed)
	require.NoError(t, err)
	err = os.WriteFile(filepath, existingData, 0644)
	require.NoError(t, err)

	as, err := flow.RetrieveAgentSeed(filepath)
	require.NoError(t, err)

	require.Equal(t, existingSeed.UID, as.UID)
	require.Equal(t, existingSeed.Version, as.Version)
	require.Equal(t, existingSeed.CreatedAt.Format(time.RFC3339), as.CreatedAt.Format(time.RFC3339))
}
