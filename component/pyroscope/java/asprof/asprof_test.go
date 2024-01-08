//go:build darwin || linux

package asprof

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestDistributionExtract(t *testing.T) {
	d := *DistributionForProcess(os.Getpid())
	tmpDir := t.TempDir()
	t.Log(tmpDir)

	err := d.write(tmpDir, tmpDirMarker)
	assert.NoError(t, err)
	assert.NotEmpty(t, d.extractedDir)
	assert.FileExists(t, d.Launcher())
	assert.FileExists(t, d.LibPath())
	libStat1, err := os.Stat(d.LibPath())
	assert.NoError(t, err)

	d = *DistributionForProcess(os.Getpid()) // extracting second time should just verify
	err = d.write(tmpDir, tmpDirMarker)
	assert.NoError(t, err)
	assert.NotEmpty(t, d.extractedDir)
	assert.FileExists(t, filepath.Join(d.Launcher()))
	assert.FileExists(t, filepath.Join(d.LibPath()))
	libStat2, err := os.Stat(d.LibPath())
	require.NoError(t, err)
	require.Equal(t, libStat1.ModTime(), libStat2.ModTime())

	file, err := os.OpenFile(d.Launcher(), os.O_RDWR, 0)
	require.NoError(t, err)
	defer file.Close()
	file.Write([]byte("hello")) //modify binary
	file.Close()

	d = *DistributionForProcess(os.Getpid()) // extracting second time should just verify
	err = d.write(tmpDir, tmpDirMarker)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "already exists and is different")
	assert.Empty(t, d.extractedDir)

}

func TestDistributionExtractRace(t *testing.T) {
	tmpDir := t.TempDir()
	race = func(stage, extra string) {
		switch stage {
		case "mkdir dist":
			err := os.MkdirAll(filepath.Join(tmpDir, extra), 0755)
			assert.NoError(t, err)
		default:
			t.Fatal("unexpected stage")
		}
	}
	defer func() {
		race = func(stage, extra string) {

		}
	}()
	d := *DistributionForProcess(os.Getpid())
	t.Log(tmpDir)
	err := d.write(tmpDir, tmpDirMarker)
	assert.Error(t, err)
	assert.Empty(t, d.extractedDir)
	assert.NoFileExists(t, filepath.Join(d.Launcher()))
	assert.NoFileExists(t, filepath.Join(d.LibPath()))
}
