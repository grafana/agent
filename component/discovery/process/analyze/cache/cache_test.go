//go:build linux

package cache

import (
	"io"
	"os"
	"testing"

	"github.com/grafana/agent/pkg/util"
	"github.com/stretchr/testify/require"
)

func copyFile(t *testing.T, src, dst string) {
	t.Helper()
	s, err := os.Open(src)
	if err != nil {
		t.Fatal(err)
	}
	defer s.Close()
	d, err := os.Create(dst)
	if err != nil {
		t.Fatal(err)
	}
	defer d.Close()
	_, err = io.Copy(d, s)
	if err != nil {
		t.Fatal(err)
	}
}

func TestCache(t *testing.T) {
	d := t.TempDir()
	copyFile(t, "/proc/self/exe", d+"/exe1")
	copyFile(t, "/proc/self/exe", d+"/exe2")
	err := os.Symlink(d+"/exe1", d+"/exe1-symlink")
	require.NoError(t, err)

	l := util.TestLogger(t)
	c := New(l)
	r1, err := c.AnalyzePIDPath(1, "1", d+"/exe1")
	require.NoError(t, err)
	r2, err := c.AnalyzePIDPath(1, "1", d+"/exe1")
	require.NoError(t, err)
	require.True(t, r1 == r2)

	r3, err := c.AnalyzePIDPath(2, "2", d+"/exe1-symlink")
	require.NoError(t, err)
	require.True(t, r1 == r3)

	require.Equal(t, 2, len(c.pids))
	require.Equal(t, 1, len(c.stats))
	require.Equal(t, 1, len(c.buildIDs))

	r4, err := c.AnalyzePIDPath(3, "3", d+"/exe2")
	require.NoError(t, err)
	require.True(t, r1 == r4)

	require.Equal(t, 3, len(c.pids))
	require.Equal(t, 2, len(c.stats))
	require.Equal(t, 1, len(c.buildIDs))

	c.GC(map[uint32]struct{}{1: {}, 2: {}, 3: {}})

	require.Equal(t, 3, len(c.pids))
	require.Equal(t, 2, len(c.stats))
	require.Equal(t, 1, len(c.buildIDs))

	c.GC(map[uint32]struct{}{2: {}, 3: {}})

	require.Equal(t, 2, len(c.pids))
	require.Equal(t, 2, len(c.stats))
	require.Equal(t, 1, len(c.buildIDs))

	r3, err = c.AnalyzePIDPath(2, "2", d+"/exe1-symlink")
	require.NoError(t, err)
	require.True(t, r1 == r3)

	r4, err = c.AnalyzePIDPath(3, "3", d+"/exe2")
	require.NoError(t, err)
	require.True(t, r1 == r4)

	c.GC(map[uint32]struct{}{3: {}})

	require.Equal(t, 1, len(c.pids))
	require.Equal(t, 1, len(c.stats))
	require.Equal(t, 1, len(c.buildIDs))

	c.GC(map[uint32]struct{}{})

	require.Equal(t, 0, len(c.pids))
	require.Equal(t, 0, len(c.stats))
	require.Equal(t, 0, len(c.buildIDs))
}
