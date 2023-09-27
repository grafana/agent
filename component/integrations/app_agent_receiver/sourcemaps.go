package app_agent_receiver

import (
	"github.com/go-kit/log"
	"github.com/go-sourcemap/sourcemap"
	"github.com/grafana/agent/component/integrations/app_agent_receiver/internal/payload"
)

// sourceMapsStore is an interface for a sourcemap service capable of
// transforming minified source locations to the original source location.
type sourceMapsStore interface {
	GetSourceMap(sourceURL string, release string) (*sourcemap.Consumer, error)
}

func transformException(log log.Logger, store sourceMapsStore, ex *payload.Exception, release string) *payload.Exception
