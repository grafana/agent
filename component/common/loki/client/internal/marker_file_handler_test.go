package internal

import (
	"os"
	"path"
	"testing"

	"github.com/go-kit/log"
	"github.com/stretchr/testify/require"
)

func TestMarkerFileHandler(t *testing.T) {
	dir := t.TempDir()
	fh, err := NewMarkerFileHandler(log.NewNopLogger(), dir)
	require.NoError(t, err)

	// write first something to marker
	markerFile := path.Join(dir, "remote", "segment")
	err = os.WriteFile(markerFile, []byte("10"), 0o666)
	require.NoError(t, err)

	require.Equal(t, 10, fh.LastMarkedSegment())

	// mark segment and re-check
	fh.MarkSegment(12)
	require.Equal(t, 12, fh.LastMarkedSegment())
	bs, err := os.ReadFile(markerFile)
	require.NoError(t, err)
	require.Equal(t, "12", string(bs))
}
