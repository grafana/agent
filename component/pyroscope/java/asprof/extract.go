package asprof

import (
	"archive/tar"
	"bytes"
	"crypto/sha1"
	"encoding/hex"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	//"path/filepath"

	"github.com/klauspost/compress/gzip"
	"golang.org/x/sys/unix"
)

func readTarGZ(buf []byte, cb func(fi fs.FileInfo, data []byte) error) error {
	gzipReader, err := gzip.NewReader(bytes.NewReader(buf))
	if err != nil {
		return err
	}
	defer gzipReader.Close()

	tarReader := tar.NewReader(gzipReader)

	for {
		header, err := tarReader.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			return err
		}

		fileInfo := header.FileInfo()
		if fileInfo.IsDir() {
			continue
		}
		buffer, err := io.ReadAll(tarReader)
		if err != nil {
			return err
		}
		err = cb(fileInfo, buffer)
		if err != nil {
			return err
		}
	}

	return nil
}

func getLibAndLauncher(targz []byte) (lib []byte, launcher []byte, err error) {
	err = readTarGZ(targz, func(fi fs.FileInfo, data []byte) error {
		if fi.Name() == "libasyncProfiler.dylib" || fi.Name() == "libasyncProfiler.so" {
			lib = data
			return nil
		}
		if fi.Name() == "profiler.sh" || fi.Name() == "asprof" {
			launcher = data
		}
		return nil
	})
	if lib == nil || launcher == nil {
		return nil, nil, fmt.Errorf("failed to find libasyncProfiler in tar.gz")
	}
	return lib, launcher, err
}

var race = func(stage, extra string) {}

const extractPerm = 0755
const tmpDirMarker = "grafana-agent-asprof"

func (d *Distribution) write(dstPath string, lib, launcher []byte) error {
	dstFile, err := os.Open(dstPath)
	if err != nil {
		return err
	}
	defer dstFile.Close()
	sum := sha1.Sum(d.targz)
	distName := fmt.Sprintf("%s-%s-%s", tmpDirMarker,
		strings.TrimSuffix(d.fname, ".tar.gz"),
		hex.EncodeToString(sum[:]))
	distDirPath := filepath.Join(dstPath, distName)
	distDir, err := os.Open(distDirPath)
	if err == nil {
		return d.verifyExtracted(distDir, lib, launcher)
	}

	race("mkdir dist", distName)

	err = unix.Mkdirat(int(dstFile.Fd()), distName, extractPerm)
	if err != nil {
		return fmt.Errorf("failed to create directory %s: %w", distDirPath, err)
	}
	newDirFD, err := unix.Openat(int(dstFile.Fd()), distName, unix.O_DIRECTORY, 0)
	if err != nil {
		return fmt.Errorf("failed to open directory %s: %w", distDirPath, err)
	}
	distDir = os.NewFile(uintptr(newDirFD), distDirPath)
	if err := validateParent(dstFile, distDir); err != nil {
		return fmt.Errorf("failed to validate parent directory %s: %w", distDirPath, err)
	}
	stat, err := distDir.Stat()
	if err != nil {
		return fmt.Errorf("failed to stat directory %s: %w", distDirPath, err)
	}
	if !stat.IsDir() {
		return fmt.Errorf("file %s is not a directory", distDirPath)
	}
	if stat.Mode().Perm() != extractPerm {
		return fmt.Errorf("directory %s has wrong permissions %s", distDirPath, stat.Mode().Perm())
	}
	err = os.MkdirAll(filepath.Join(distDirPath, "bin"), extractPerm)
	if err != nil {
		return fmt.Errorf("failed to create bin directory %s: %w", distDirPath, err)
	}
	err = os.MkdirAll(filepath.Join(distDirPath, "lib"), extractPerm)
	if err != nil {
		return fmt.Errorf("failed to create lib directory %s: %w", distDirPath, err)
	}
	err = WriteNonExistingFile(filepath.Join(distDirPath, d.LibPath()), lib, extractPerm)
	if err != nil {
		return fmt.Errorf("failed to write file %s : %w", d.LibPath(), err)
	}
	err = WriteNonExistingFile(filepath.Join(distDirPath, d.AsprofPath()), launcher, extractPerm)
	if err != nil {
		return fmt.Errorf("failed to write file %s : %w", d.AsprofPath(), err)
	}
	d.extractedDir = distDirPath
	return nil
}

func validateParent(parent, child *os.File) error {
	parentPath, err := readLinkFD(parent)
	if err != nil {
		return fmt.Errorf("readlinkfd %s %w", parent.Name(), err)
	}
	childPath, err := readLinkFD(child)
	if err != nil {
		return fmt.Errorf("readlinkfd %s %w", child.Name(), err)
	}
	if !strings.HasPrefix(childPath, parentPath+"/") {
		return fmt.Errorf("parent %s is not a parent of child %s", parentPath, childPath)
	}
	return nil
}

func (d *Distribution) verifyExtracted(distDir *os.File, lib, launcher []byte) error {
	distDirPath := distDir.Name()

	prevLib, err := os.ReadFile(filepath.Join(distDirPath, d.LibPath()))
	if err != nil {
		return fmt.Errorf("failed to read file %s : %w", d.LibPath(), err)
	}
	prevLauncher, err := os.ReadFile(filepath.Join(distDirPath, d.AsprofPath()))
	if err != nil {
		return fmt.Errorf("failed to read file %s : %w", d.AsprofPath(), err)
	}
	if !bytes.Equal(lib, prevLib) {
		return fmt.Errorf("file %s %s already exists and is different", d.LibPath(), distDirPath)
	}
	if !bytes.Equal(launcher, prevLauncher) {
		return fmt.Errorf("file %s %s already exists and is different", d.AsprofPath(), distDirPath)
	}

	d.extractedDir = distDirPath
	return nil
}

func WriteNonExistingFile(name string, data []byte, perm os.FileMode) error {
	f, err := os.OpenFile(name, os.O_WRONLY|os.O_CREATE|os.O_EXCL, perm)
	if err != nil {
		return err
	}
	_, err = f.Write(data)
	if err1 := f.Close(); err1 != nil && err == nil {
		err = err1
	}
	return err
}
