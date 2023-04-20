package http

import (
	"github.com/weaveworks/common/server"
)

// Config is a wrapper around server.Config.
type Config struct {
	Server server.Config
}
