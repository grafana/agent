package module

import (
	"github.com/grafana/agent/pkg/cluster"
	"github.com/grafana/agent/pkg/flow/logging"
	"github.com/grafana/agent/pkg/flow/tracing"
	"github.com/prometheus/client_golang/prometheus"
)

// Options holds static options fora module.
type Options struct {
	// ControllerID is an identifier used to represent the controller.
	// ControllerID is used to generate a globally unique display name for
	// components in a binary where multiple controllers are used.
	//
	// If running multiple Flow controllers, each controller must have a
	// different value for ControllerID to be able to differentiate between
	// components in telemetry data.
	ControllerID string

	// LogSink to use for controller logs and components. A no-op logger will be
	// created if this is nil.
	LogSink *logging.Sink

	// Tracer for components to use. A no-op tracer will be created if this is
	// nil.
	Tracer *tracing.Tracer

	// Clusterer for implementing distributed behavior among components running
	// on different nodes.
	Clusterer *cluster.Clusterer

	// Directory where components can write data. Constructed components will be
	// given a subdirectory of DataPath using the local ID of the component.
	//
	// If running multiple Flow controllers, each controller must have a
	// different value for DataPath to prevent components from colliding.
	DataPath string

	// Reg is the prometheus register to use
	Reg prometheus.Registerer

	// HTTPPathPrefix is the path prefix given to managed components. May be
	// empty. When provided, it should be an absolute path.
	//
	// Components will be given a path relative to HTTPPathPrefix using their
	// local ID.
	//
	// If running multiple Flow controllers, each controller must have a
	// different value for HTTPPathPrefix to prevent components from colliding.
	HTTPPathPrefix string

	// HTTPListenAddr is the base address that the server is listening on.
	// The controller does not itself listen here, but some components
	// need to know this to set the correct targets.
	HTTPListenAddr string
}
