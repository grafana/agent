package asprof

import (
	"archive/tar"
	"bytes"
	"io"
	"os"
	"path/filepath"

	"github.com/klauspost/compress/gzip"
)

func extractTarGZ(buf []byte, dst string) error {
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
		dir := filepath.Join(dst, filepath.Dir(header.Name))
		filename := filepath.Join(dir, fileInfo.Name())
		if fileInfo.IsDir() {
			continue
		}
		err = os.MkdirAll(dir, 0755)
		if err != nil {
			return err
		}

		buffer, err := io.ReadAll(tarReader)
		if err != nil {
			return err
		}
		//tyodo make sure dst is not a symlink

		file, err := os.Create(filename)
		if err != nil {
			return err
		}
		defer file.Close()

		_, err = file.Write(buffer)
		if err != nil {
			return err
		}

		err = file.Close()
		if err != nil {
			return err
		}
	}

	return nil
}
