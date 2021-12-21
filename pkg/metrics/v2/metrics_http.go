package metrics

import (
	"encoding/json"
	"net/http"
	"sort"

	"github.com/gorilla/mux"
	"github.com/grafana/agent/pkg/metrics/cluster/configapi"
	"github.com/prometheus/common/model"
)

// WireAPI wires up HTTP API routes for the metrics subsystem.
func (m *Metrics) WireAPI(r *mux.Router) {
	r.HandleFunc("/agent/api/v1/metrics/discovery/jobs", m.listDiscoveryJobsHandler).Methods(http.MethodGet)
	r.HandleFunc("/agent/api/v1/metrics/discovery/targets", m.listDiscoveryTargetsHandler).Methods(http.MethodGet)
}

func (m *Metrics) listDiscoveryJobsHandler(w http.ResponseWriter, _ *http.Request) {
	jobs := m.discoverers.getDiscoveryJobs()
	_ = configapi.WriteResponse(w, http.StatusOK, jobs)
}

func (m *Metrics) listDiscoveryTargetsHandler(w http.ResponseWriter, _ *http.Request) {
	jobs := m.discoverers.getDiscoveryTargets()
	_ = configapi.WriteResponse(w, http.StatusOK, jobs)
}

type discoveryJobs struct {
	Instances []discoveryJobsInstance `json:"instances"`
}

type discoveryJobsInstance struct {
	Name string   `json:"name"`
	Jobs []string `json:"jobs"`
}

type discoveryTargets struct {
	Instances []discoveryTargetsInstance `json:"instances"`
}

type discoveryTargetsInstance struct {
	Name   string
	Groups targetGroups
}

func (d discoveryTargetsInstance) MarshalJSON() ([]byte, error) {
	type targetInfo struct {
		TargetGroup string         `json:"target_group"`
		Labels      model.LabelSet `json:"labels"`
	}

	var info []targetInfo

	for groupName, groups := range d.Groups {
		for _, group := range groups {
			for _, target := range group.Targets {
				info = append(info, targetInfo{
					TargetGroup: groupName,
					Labels:      group.Labels.Merge(target),
				})
			}
		}
	}
	sort.Slice(info, func(i, j int) bool {
		// sort by target group, then address label.
		var (
			iGroup   = info[i].TargetGroup
			iAddress = string(info[i].Labels[model.AddressLabel])

			jGroup   = info[j].TargetGroup
			jAddress = string(info[j].Labels[model.AddressLabel])
		)

		switch {
		case iGroup != jGroup:
			return iGroup < jGroup
		default:
			return iAddress < jAddress
		}
	})

	type out struct {
		Name    string       `json:"name"`
		Targets []targetInfo `json:"targets"`
	}
	return json.Marshal(out{
		Name:    d.Name,
		Targets: info,
	})
}
