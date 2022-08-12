// Package api implements the HTTP API used for the Grafana Agent Flow UI.
//
// The API is internal only; it is not stable and shouldn't be relied on
// externally.
package api

// All API methods should be JSON.
//
// API methods needed:
//
// /api/v0/web/components
//
// Return list of components ID and health status. Arguments and exports
// should not be included.
//
// /api/v0/web/component/{id}
//
// Return details on a component:
//   * Arguments
//   * Exports
//   * Debug info
//   * Health info
//   * Dependencies
//   * Dependants
//
//  /api/v0/web/dag
//
//  Return the DAG to be interpreted by Javascript for rendering.
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
