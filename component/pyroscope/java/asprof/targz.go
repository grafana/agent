package asprof

import (
	"archive/tar"
	"bufio"
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
			err = os.MkdirAll(filename, 0755)
			if err != nil {
				return err
			}
			continue
		}
		err = os.MkdirAll(dir, 0755)
		if err != nil {
			return err
		}

		file, err := os.Create(filename)
		if err != nil {
			return err
		}

		writer := bufio.NewWriter(file)

		buffer := make([]byte, 4096)
		for {
			n, err := tarReader.Read(buffer)
			if err != nil && err != io.EOF {
				panic(err)
			}
			if n == 0 {
				break
			}

			_, err = writer.Write(buffer[:n])
			if err != nil {
				return err
			}
		}

		err = writer.Flush()
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
