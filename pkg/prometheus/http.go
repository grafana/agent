package prometheus

import (
	"net/http"

	"github.com/go-kit/kit/log/level"
	"github.com/gorilla/mux"
	"github.com/grafana/agent/pkg/prometheus/ha/configapi"
)

// WireAPI adds API routes to the provided mux router.
func (a *Agent) WireAPI(r *mux.Router) {
	if a.cfg.ServiceConfig.Enabled {
		a.ha.WireAPI(r)
	}

	r.HandleFunc("/agent/api/v1/instances", a.ListInstancesHandler).Methods("GET")
}

// ListInstances writes the set of currently running instances to the http.ResponseWriter.
func (a *Agent) ListInstancesHandler(w http.ResponseWriter, r *http.Request) {
	var instanceNames []string
	for k := range a.cm.ListConfigs() {
		instanceNames = append(instanceNames, k)
	}
	err := configapi.WriteResponse(w, http.StatusOK, instanceNames)
	if err != nil {
		level.Error(a.logger).Log("msg", "failed to write response", "err", err)
	}
}
