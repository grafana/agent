//go:build !builtinassets

package ui

import (
	"net/http"
	"path/filepath"
)

// Assets contains the UI's assets.
func Assets() http.FileSystem {
	assetsDir := filepath.Join(".", "web", "ui", "build")
	return http.Dir(assetsDir)
}
