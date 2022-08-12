// Package api implements the HTTP API used for the Grafana Agent Flow UI.
//
// The API is internal only; it is not stable and shouldn't be relied on
// externally.
package api

// Unless otherwise specified, API methods should be JSON.
//
// API methods needed:
//
// /api/v0/web/components
//
// Return list of components, where each component contains:
//   * component ID
//   * component name (metrics.remote_write)
//   * component label
//   * health info
//   * component IDs of components being referenced by this component
//   * component IDs of components referencing this component
//
// Arguments, exports, and debug info are *not* included.
//
// /api/v0/web/component/{id}
//
// Return details on a component:
//   * component name (metrics.remote_write)
//   * Arguments
//   * Exports
//   * Debug info
//   * Health info
//   * Dependencies
//   * Dependants
//
// /api/v0/web/component/{id}/raw
//
// Return raw evaluated River text for component
//
// /api/v0/web/status/build-info
//
//   Go runtime, build information (like the Prometheus page)
//
// /api/v0/web/status/flags
//
//   Command-line flags used to launch application
//
// /api/v0/web/status/config-file
//
//   Parsed config file (*not* evaluated config file)
