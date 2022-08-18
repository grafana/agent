package ui

import "net/http"

// Assets contains the UI's assets.
var Assets = func() http.FileSystem {
	return http.Dir("./web/ui/build")
}
