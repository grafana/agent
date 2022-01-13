package metrics

import (
	"github.com/gorilla/mux"
	"github.com/grafana/agent/pkg/metrics/instance"
	"google.golang.org/grpc"
)

// Subsystem is the interface exposed by the metrics subsystem.
type Subsystem interface {
	InstanceManager() instance.Manager
	Validate(*instance.Config) error
	WireAPI(*mux.Router)
	WireGRPC(*grpc.Server)
	Stop()
}
