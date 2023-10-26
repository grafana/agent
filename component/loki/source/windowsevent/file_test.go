//go:build windows

package windowsevent

import (
	"fmt"
	"github.com/stretchr/testify/require"
	"os"
	"path/filepath"
	"testing"
)

func TestBookmarkFile(t *testing.T) {
	dir := t.TempDir()
	bf, err := NewBookmarkFile(dir, "")
	require.NoError(t, err)
	require.NotNil(t, bf)
	require.True(t, bf.index == 1)
	err = bf.Put("<xml></xml>")
	require.NoError(t, err)
	require.True(t, bf.index == 2)
}

func TestBookmarkFileWithLegacy(t *testing.T) {
	dir := t.TempDir()
	legacy := t.TempDir()
	legacyPath := filepath.Join(legacy, "bookmark.xml")
	err := os.WriteFile(legacyPath, []byte("<xml>-1</xml>"), 0600)
	require.NoError(t, err)
	bf, err := NewBookmarkFile(dir, legacyPath)

	require.NoError(t, err)
	require.NotNil(t, bf)
	require.True(t, bf.index == 1)

	// Legacy should no longer be there.
	_, err = os.Stat(legacyPath)
	require.ErrorIs(t, err, os.ErrNotExist)

	// bookmark.1.xml should have been created from the legacy item.
	content, err := os.ReadFile(filepath.Join(dir, "bookmark.1.xml"))
	require.NoError(t, err)
	require.True(t, string(content) == "<xml>-1</xml>")
	err = bf.Put("<xml>2</xml>")
	require.NoError(t, err)
	require.True(t, bf.index == 2)

	// Bookmark 1 should not exist
	_, err = os.ReadFile(filepath.Join(dir, "bookmark.1.xml"))
	require.ErrorIs(t, err, os.ErrNotExist)

	// Bookmark 2 should  exist
	content, err = os.ReadFile(filepath.Join(dir, "bookmark.2.xml"))
	require.NoError(t, err)
	require.True(t, string(content) == "<xml>2</xml>")
}

func TestMultipleBookmarks(t *testing.T) {
	dir := t.TempDir()
	for i := 1; i <= 10; i++ {
		writeBookmark(t, dir, i)
	}
	bf, err := NewBookmarkFile(dir, "")
	require.NoError(t, err)
	require.NotNil(t, bf)
	require.True(t, bf.index == 10)
	require.True(t, bf.currentPath == "bookmark.10.xml")
	files, err := os.ReadDir(dir)
	require.NoError(t, err)
	require.Len(t, files, 1)
}

func TestMultipleBookmarksWithInvalidXML(t *testing.T) {
	dir := t.TempDir()
	for i := 1; i <= 10; i++ {
		writeBookmark(t, dir, i)
	}
	// Make 10 an invalid file.
	err := os.WriteFile(filepath.Join(dir, "bookmark.10.xml"), []byte("bad"), 0600)
	require.NoError(t, err)
	bf, err := NewBookmarkFile(dir, "")
	require.NoError(t, err)
	require.NotNil(t, bf)
	require.True(t, bf.index == 9)
	require.True(t, bf.currentPath == "bookmark.9.xml")
	files, err := os.ReadDir(dir)
	require.NoError(t, err)
	require.Len(t, files, 1)
}

func writeBookmark(t *testing.T, dir string, index int) {
	fileName := fmt.Sprintf("bookmark.%d.xml", index)
	err := os.WriteFile(filepath.Join(dir, fileName), []byte(fmt.Sprintf("<xml>%d</xml>", index)), 0600)
	require.NoError(t, err)
}
