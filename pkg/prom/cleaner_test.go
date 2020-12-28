package prom

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/go-kit/kit/log"
	"github.com/grafana/agent/pkg/prom/instance"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Basic logger to write to stderr to help debug test failures
func getLogger() log.Logger {
	return log.NewLogfmtLogger(os.Stderr)
}

func TestWALCleaner_getAllStorageNoRoot(t *testing.T) {
	walRoot := filepath.Join(os.TempDir(), "getAllStorageNoRoot")
	cleaner := NewWALCleaner(getLogger(), &mockManager{}, walRoot)

	// Bogus WAL root that doesn't exist. Method should return no results
	wals := cleaner.getAllStorage()

	assert.Empty(t, wals)
}

func TestWALCleaner_getAllStorageSuccess(t *testing.T) {
	walRoot, err := ioutil.TempDir(os.TempDir(), "getAllStorageSuccess")
	require.NoError(t, err)
	defer os.RemoveAll(walRoot)

	walDir := filepath.Join(walRoot, "instance-1")
	err = os.MkdirAll(walDir, 0755)
	require.NoError(t, err)

	cleaner := NewWALCleaner(getLogger(), &mockManager{}, walRoot)
	wals := cleaner.getAllStorage()

	assert.Equal(t, []string{walDir}, wals)
}

func TestWALCleaner_getAbandonedStorageBeforeCutoff(t *testing.T) {
	walRoot, err := ioutil.TempDir(os.TempDir(), "getAbandonedStorageBeforeCutoff")
	require.NoError(t, err)
	defer os.RemoveAll(walRoot)

	walDir := filepath.Join(walRoot, "instance-1")
	err = os.MkdirAll(walDir, 0755)
	require.NoError(t, err)

	all := []string{walDir}
	managed := make(map[string]bool)
	now := time.Now()

	cleaner := NewWALCleaner(
		getLogger(),
		&mockManager{},
		walRoot,
		WithCleanerMinAge(5*time.Minute),
	)

	cleaner.walOperations = &mockWalOperations{
		mtime: now,
		err:   nil,
	}

	// Last modification time on our WAL directory is the same as "now"
	// so there shouldn't be any results even though it's not part of the
	// set of "managed" directories.
	abandoned := cleaner.getAbandonedStorage(all, managed, now)
	assert.Empty(t, abandoned)
}

func TestWALCleaner_getAbandonedStorageAfterCutoff(t *testing.T) {
	walRoot, err := ioutil.TempDir(os.TempDir(), "getAbandonedStorageAfterCutoff")
	require.NoError(t, err)
	defer os.RemoveAll(walRoot)

	walDir := filepath.Join(walRoot, "instance-1")
	err = os.MkdirAll(walDir, 0755)
	require.NoError(t, err)

	all := []string{walDir}
	managed := make(map[string]bool)
	now := time.Now()

	cleaner := NewWALCleaner(
		getLogger(),
		&mockManager{},
		walRoot,
		WithCleanerMinAge(5*time.Minute),
	)

	cleaner.walOperations = &mockWalOperations{
		mtime: now.Add(-30 * time.Minute),
		err:   nil,
	}

	// Last modification time on our WAL directory is 30 minutes in the past
	// compared to "now" and we've set the cutoff for our cleaner to be 5
	// minutes: our WAL directory should show up as abandoned
	abandoned := cleaner.getAbandonedStorage(all, managed, now)
	assert.Equal(t, []string{walDir}, abandoned)
}

func TestWALCleaner_CleanupStorage(t *testing.T) {
	walRoot, err := ioutil.TempDir(os.TempDir(), "CleanupStorage")
	require.NoError(t, err)
	defer os.RemoveAll(walRoot)

	walDir := filepath.Join(walRoot, "instance-1")
	err = os.MkdirAll(walDir, 0755)
	require.NoError(t, err)

	now := time.Now()
	cleaner := NewWALCleaner(
		getLogger(),
		&mockManager{},
		walRoot,
		WithCleanerMinAge(5*time.Minute),
	)

	cleaner.walOperations = &mockWalOperations{
		mtime: now.Add(-30 * time.Minute),
		err:   nil,
	}

	// Last modification time on our WAL directory is 30 minutes in the past
	// compared to "now" and we've set the cutoff for our cleaner to be 5
	// minutes: our WAL directory should be removed since it's abandoned
	cleaner.CleanupStorage()
	_, err = os.Stat(walDir)
	assert.Error(t, err)
	assert.True(t, os.IsNotExist(err))
}

type mockManager struct{}

func (m *mockManager) ListInstances() map[string]instance.ManagedInstance {
	return make(map[string]instance.ManagedInstance)
}

func (m *mockManager) ListConfigs() map[string]instance.Config {
	return make(map[string]instance.Config)
}

func (m *mockManager) ApplyConfig(_ instance.Config) error {
	return nil
}

func (m *mockManager) DeleteConfig(_ string) error {
	return nil
}

func (m *mockManager) Stop() {
	// nop
}

type mockWalOperations struct {
	mtime time.Time
	err   error
}

func (m *mockWalOperations) lastModified(path string) (time.Time, error) {
	return m.mtime, m.err
}
