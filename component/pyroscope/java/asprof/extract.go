package asprof

import (
	"archive/tar"
	"bytes"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	//"path/filepath"

	"github.com/klauspost/compress/gzip"
)

func readTarGZ(buf []byte, cb func(name string, fi fs.FileInfo, data []byte) error) error {
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
		err = cb(header.Name, fileInfo, buffer)
		if err != nil {
			return err
		}
	}

	return nil
}

//var race = func(stage, extra string) {}

const extractPerm = 0755
const tmpDirMarker = "grafana-agent-asprof"

func writeFile(dir, path string, data []byte) error {
	dstPath := filepath.Join(dir, path)
	if err := os.MkdirAll(filepath.Dir(dstPath), extractPerm); err != nil {
		return fmt.Errorf("failed to create directory %s: %w", filepath.Dir(dstPath), err)
	}
	if f, err := os.Open(dstPath); err == nil {
		defer f.Close()
		prevData, err := io.ReadAll(f)
		if err != nil {
			return fmt.Errorf("failed to read file %s : %w", dstPath, err)
		}
		if !bytes.Equal(prevData, data) {
			return fmt.Errorf("file %s already exists and is different", dstPath)
		}
		return nil
	}
	return WriteNonExistingFile(dstPath, data, extractPerm)
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
