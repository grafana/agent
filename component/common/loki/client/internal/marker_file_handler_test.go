package internal

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/go-kit/log"
	"github.com/stretchr/testify/require"
)

func TestMarkerFileHandler(t *testing.T) {
	logger := log.NewLogfmtLogger(os.Stdout)
	getTempDir := func(t *testing.T) string {
		dir := t.TempDir()
		require.NoError(t, os.Chmod(dir, MarkerFileMode), "failed to change temporary dir mode")
		return dir
	}

	t.Run("reads the last segment from existing marker file", func(t *testing.T) {
		dir := getTempDir(t)
		fh, err := NewMarkerFileHandler(logger, dir)
		require.NoError(t, err)

		// write first something to marker
		markerFile := filepath.Join(dir, MarkerFolderName, MarkerFileName)
		bs, err := EncodeMarkerV1(10)
		require.NoError(t, err)
		err = os.WriteFile(markerFile, bs, MarkerFileMode)
		require.NoError(t, err)

		require.Equal(t, 10, fh.LastMarkedSegment())
	})

	t.Run("marks segment, and then reads value from it", func(t *testing.T) {
		dir := getTempDir(t)
		fh, err := NewMarkerFileHandler(logger, dir)
		require.NoError(t, err)

		fh.MarkSegment(12)
		require.Equal(t, 12, fh.LastMarkedSegment())
	})

	t.Run("marker file is created with correct permissions", func(t *testing.T) {
		dir := getTempDir(t)
		fh, err := NewMarkerFileHandler(logger, dir)
		require.NoError(t, err)

		fh.MarkSegment(12)
		stats, err := os.Stat(filepath.Join(dir, MarkerFolderName, MarkerFileName))
		require.NoError(t, err)
		require.Equal(t, MarkerFileMode, stats.Mode().Perm())
	})
}
