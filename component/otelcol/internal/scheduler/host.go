package scheduler

import (
	"github.com/go-kit/log"
	"github.com/go-kit/log/level"

	otelcomponent "go.opentelemetry.io/collector/component"
	otelconfig "go.opentelemetry.io/collector/config"
)

// Host implements otelcomponent.Host for Grafana Agent Flow.
type Host struct {
	log log.Logger

	extensions map[otelconfig.ComponentID]otelcomponent.Extension
	exporters  map[otelconfig.DataType]map[otelconfig.ComponentID]otelcomponent.Exporter
}

// NewHost creates a new Host.
func NewHost(l log.Logger, opts ...HostOption) *Host {
	h := &Host{log: l}
	for _, opt := range opts {
		opt(h)
	}
	return h
}

// HostOption customizes behavior of the Host.
type HostOption func(*Host)

// WithHostExtensions provides a custom set of extensions to the Host.
func WithHostExtensions(extensions map[otelconfig.ComponentID]otelcomponent.Extension) HostOption {
	return func(h *Host) {
		h.extensions = extensions
	}
}

// WithHostExporters provides a custom set of exporters to the Host.
func WithHostExporters(exporters map[otelconfig.DataType]map[otelconfig.ComponentID]otelcomponent.Exporter) HostOption {
	return func(h *Host) {
		h.exporters = exporters
	}
}

var _ otelcomponent.Host = (*Host)(nil)

// ReportFatalError implements otelcomponent.Host.
func (h *Host) ReportFatalError(err error) {
	level.Error(h.log).Log("msg", "fatal error running component", "err", err)
}

// GetFactory implements otelcomponent.Host.
func (h *Host) GetFactory(kind otelcomponent.Kind, componentType otelconfig.Type) otelcomponent.Factory {
	// GetFactory is used for components to create other components. It's not
	// clear if we want to allow this right now, so it's disabled.
	return nil
}

// GetExtensions implements otelcomponent.Host.
func (h *Host) GetExtensions() map[otelconfig.ComponentID]otelcomponent.Extension {
	return h.extensions
}

// GetExporters implements otelcomponent.Host.
func (h *Host) GetExporters() map[otelconfig.DataType]map[otelconfig.ComponentID]otelcomponent.Exporter {
	return h.exporters
}
