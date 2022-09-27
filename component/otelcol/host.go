package otelcol

import (
	"github.com/go-kit/log"
	"github.com/go-kit/log/level"

	otelcomponent "go.opentelemetry.io/collector/component"
	otelconfig "go.opentelemetry.io/collector/config"
)

// host implements otelcomponent.Host for Grafana Agent Flow.
type host struct {
	log log.Logger

	extensions map[otelconfig.ComponentID]otelcomponent.Extension
	exporters  map[otelconfig.DataType]map[otelconfig.ComponentID]otelcomponent.Exporter
}

func newHost(l log.Logger) *host {
	return &host{log: l}
}

var _ otelcomponent.Host = (*host)(nil)

func (h *host) ReportFatalError(err error) {
	level.Error(h.log).Log("msg", "fatal error running component", "err", err)
}

func (h *host) GetFactory(kind otelcomponent.Kind, componentType otelconfig.Type) otelcomponent.Factory {
	// GetFactory is used for components to create other components. It's not
	// clear if we want to allow this right now, so it's disabled.
	return nil
}

func (h *host) GetExtensions() map[otelconfig.ComponentID]otelcomponent.Extension {
	return h.extensions
}

func (h *host) GetExporters() map[otelconfig.DataType]map[otelconfig.ComponentID]otelcomponent.Exporter {
	return h.exporters
}
