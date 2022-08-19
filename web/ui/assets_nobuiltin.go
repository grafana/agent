//go:build !builtinassets
// +build !builtinassets

package ui

import "net/http"

// Assets contains the UI's assets.
func Assets() http.FileSystem {
	return http.Dir("./web/ui/build")
}
