package wal

import (
	"testing"
	"time"

	"github.com/grafana/agent/pkg/flow/logging"
	"github.com/stretchr/testify/require"
)

func TestDBWriteGet(t *testing.T) {
	l := logging.New(nil)
	database, err := newDb(t.TempDir(), l)
	require.NoError(t, err)
	var insert string
	insert = "hello world"
	key, err := database.writeRecordWithAutoKey(&insert, 5*time.Minute)
	require.NoError(t, err)
	require.True(t, key > 0)
	var result string
	found, err := database.getRecordByUint(key, &result)
	require.NoError(t, err)
	require.True(t, found)
	require.True(t, result == "hello world")
}
