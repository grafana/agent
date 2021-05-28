package config

import (
	"bytes"
	"io"
	"io/fs"
	"path"
	"strings"

	jsonnet "github.com/google/go-jsonnet"
)

// FSImporter implements jsonnet.Importer for a fs.FS.
type FSImporter struct {
	fs fs.FS

	cache map[string]jsonnet.Contents
}

// NewFSImporter creates a new jsonnet VM Importer that uses the given fs.
func NewFSImporter(f fs.FS) *FSImporter {
	return &FSImporter{
		fs:    f,
		cache: make(map[string]jsonnet.Contents),
	}
}

// Import implements jsonnet.Importer.
func (i *FSImporter) Import(importedFrom, importedPath string) (contents jsonnet.Contents, foundAt string, err error) {
	cleanedPath := path.Clean(
		path.Join(path.Dir(importedFrom), importedPath),
	)
	cleanedPath = strings.TrimPrefix(cleanedPath, "./")

	// jsonnet expects the same "foundAt" to always return the same instance of
	// contents, so we need to return a cache here.
	if c, ok := i.cache[cleanedPath]; ok {
		return c, cleanedPath, nil
	}

	f, err := i.fs.Open(cleanedPath)
	if err != nil {
		err = jsonnet.RuntimeError{Msg: err.Error()}
		return
	}

	var buf bytes.Buffer
	if _, copyErr := io.Copy(&buf, f); copyErr != nil {
		err = jsonnet.RuntimeError{Msg: copyErr.Error()}
		return
	}

	contents = jsonnet.MakeContents(buf.String())
	i.cache[cleanedPath] = contents
	return contents, cleanedPath, nil
}
