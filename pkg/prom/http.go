package prom

import (
	"net/http"
	"sort"

	"github.com/go-kit/kit/log/level"
	"github.com/gorilla/mux"
	"github.com/grafana/agent/pkg/prom/ha/configapi"
)

// WireAPI adds API routes to the provided mux router.
func (a *Agent) WireAPI(r *mux.Router) {
	if a.cfg.ServiceConfig.Enabled {
		a.ha.WireAPI(r)
	}

	r.HandleFunc("/agent/api/v1/instances", a.ListInstancesHandler).Methods("GET")
}

// ListInstances writes the set of currently running instances to the http.ResponseWriter.
func (a *Agent) ListInstancesHandler(w http.ResponseWriter, _ *http.Request) {
	cfgs := a.cm.ListConfigs()
	instanceNames := make([]string, 0, len(cfgs))
	for k := range cfgs {
		instanceNames = append(instanceNames, k)
	}
	sort.Strings(instanceNames)

	err := configapi.WriteResponse(w, http.StatusOK, instanceNames)
	if err != nil {
		level.Error(a.logger).Log("msg", "failed to write response", "err", err)
	}
}
