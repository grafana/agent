//go:build linux

package asprof

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

// extracting to /tmp
// /tmp dir should be sticky or owned 0700 by the current user
// /tmp/dist-... dir should be owned 0700 by the current user and should not be a symlink
// the rest should use mkdirAt, openAt

// test /tmp/dist-... is not symlink to /proc/conatinerpid/root/tmp/dist-
// test /tmp/dist-... is not symlink to /../../../foo

// write skippable tests with uid=0
func TestStickyDir(t *testing.T) {
	dir := "/tmp"
	p := NewProfiler(dir, EmbeddedArchive)
	p.tmpDirMarker = fmt.Sprintf("grafana-agent-asprof-%s", uuid.NewString())
	t.Logf("tmpDirMarker: %s", p.tmpDirMarker)
	err := p.ExtractDistributions()
	assert.NoError(t, err)
}

func TestOwnedDir(t *testing.T) {
	dir := tempDir(t)
	err := os.Chmod(dir, 0755)
	assert.NoError(t, err)
	p := NewProfiler(dir, EmbeddedArchive)
	err = p.ExtractDistributions()
	assert.NoError(t, err)
}

func TestOwnedDirWrongPermission(t *testing.T) {
	dir := tempDir(t)
	err := os.Chmod(dir, 0777)
	assert.NoError(t, err)
	p := NewProfiler(dir, EmbeddedArchive)
	err = p.ExtractDistributions()
	assert.Error(t, err)
}

func TestDistSymlink(t *testing.T) {
	// check if /tmp/dist-... is a symlink
	td := []bool{true, false}
	for _, glibc := range td {
		t.Run(fmt.Sprintf("glibc=%t", glibc), func(t *testing.T) {
			root := tempDir(t)
			err := os.Chmod(root, 0755)
			assert.NoError(t, err)
			manipulated := tempDir(t)
			err = os.Chmod(manipulated, 0755)
			assert.NoError(t, err)
			p := NewProfiler(root, EmbeddedArchive)
			muslDistName, glibcDistName := p.getDistNames()

			if glibc {
				err = os.Symlink(manipulated, filepath.Join(root, muslDistName))
				assert.NoError(t, err)
			} else {
				err = os.Symlink(manipulated, filepath.Join(root, glibcDistName))
				assert.NoError(t, err)
			}

			err = p.ExtractDistributions()
			t.Logf("expected %s", err)
			assert.Error(t, err)
		})
	}
}

//this one requires root and docker(or mount namespace)
//func TestDistSymlinkIntoProcRoot(t *testing.T) {
//	if os.Getuid() != 0 {
//		t.Fail()
//	}
//	td := []bool{true, false}
//	victim := "/proc/5186/root/"
//	for _, glibc := range td {
//		t.Run(fmt.Sprintf("glibc=%t", glibc), func(t *testing.T) {
//			// check if /tmp/dist-... is a symlink to /proc/pid/root/tmp/dist-
//			root := tempDir(t)
//			err := os.Chmod(root, 0755)
//			assert.NoError(t, err)
//
//			p := NewProfiler(root, embeddedArchive)
//			muslDistName, glibcDistName := p.getDistNames()
//			fake := ""
//			if glibc {
//				fake = filepath.Join(victim, root, glibcDistName)
//				err = os.Symlink(fake, filepath.Join(root, glibcDistName))
//				assert.NoError(t, err)
//			} else {
//				fake = filepath.Join(victim, root, muslDistName)
//				err = os.Symlink(fake, filepath.Join(root, muslDistName))
//				assert.NoError(t, err)
//			}
//			err = mkdirAll(fake)
//			assert.NoError(t, err)
//
//			err = p.ExtractDistributions()
//			t.Logf("expected error %s", err)
//			assert.Error(t, err)
//		})
//	}
//}

// this one requires root and docker(or mount namespace)
//func TestFlag(t *testing.T) {
//	if os.Getuid() != 0 {
//		t.Fail()
//	}
//	procFlag := "/proc/20837/root"
//	f, err := os.Open(procFlag)
//	assert.NoError(t, err)
//
//	err = writeFile(f, "asd/qwe", []byte("FLAG"))
//	assert.Error(t, err)
//	file, err := os.ReadFile("/asd/qwe")
//	assert.Error(t, err)
//	assert.NotContains(t, string(file), "FLAG")
//
//}

//func mkdirAll(fake string) error {
//	parts := strings.Split(fake, string(filepath.Separator))
//	it := "/"
//	for i := 0; i < len(parts); i++ {
//		it = filepath.Join(it, parts[i])
//		if _, err := os.Stat(it); err != nil {
//			err := os.Mkdir(it, 0755)
//			if err != nil {
//				return err
//			}
//		}
//	}
//	return nil
//}

func tempDir(t *testing.T) string {
        t.Helper()
	dir, err := os.MkdirTemp("", "asprof-test")
	assert.NoError(t, err)
	t.Logf("dir: %s", dir)
	return dir
}

// copying the lib to /proc/pid/root/tmp/dist-.../libasyncProfiler.so
