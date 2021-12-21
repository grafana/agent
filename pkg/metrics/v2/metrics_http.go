package metrics

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"sort"

	"github.com/gorilla/mux"
	"github.com/grafana/agent/pkg/cluster/httpclient"
	"github.com/grafana/agent/pkg/metrics/cluster/configapi"
	"github.com/prometheus/common/model"
	"github.com/rfratto/ckit"
	"github.com/weaveworks/common/httpgrpc"
	"google.golang.org/grpc"
)

const (
	discoveryJobsAPIEndpoint    = "/agent/api/v1/metrics/discovery/jobs"
	discoveryTargetsAPIEndpoint = "/agent/api/v1/metrics/discovery/targets"
)

// WireAPI wires up HTTP API routes for the metrics subsystem.
func (m *Metrics) WireAPI(r *mux.Router) {
	r.HandleFunc(discoveryJobsAPIEndpoint, m.listDiscoveryJobsHandler).Methods(http.MethodGet)
	r.HandleFunc(discoveryTargetsAPIEndpoint, m.listDiscoveryTargetsHandler).Methods(http.MethodGet)
}

func (m *Metrics) listDiscoveryJobsHandler(w http.ResponseWriter, r *http.Request) {
	var (
		peers     = m.node.Peers()
		localOnly = r.URL.Query().Get("remote") == "0"
	)

	var jobs []discoveryJob

	for _, p := range peers {
		if p.Self {
			jobs = append(jobs, withNodeName(p.Name, m.discoverers.getDiscoveryJobs())...)
			continue
		}
		if localOnly {
			// Skip over remote peers when the query is only for the local node.
			continue
		}

		var peerJobs []discoveryJob
		err := queryPeer(r.Context(), &p, discoveryJobsAPIEndpoint+"?remote=0", &peerJobs)
		if err != nil {
			configapi.WriteError(w, http.StatusInternalServerError, err)
			return
		}
		jobs = append(jobs, peerJobs...)
	}

	// Sort slice by instance and job name. Node name is not used when sorting.
	sort.Slice(jobs, func(i, j int) bool {
		switch {
		case jobs[i].Instance != jobs[j].Instance:
			return jobs[i].Instance < jobs[j].Instance
		default:
			return jobs[i].Name < jobs[j].Name
		}
	})

	_ = configapi.WriteResponse(w, http.StatusOK, jobs)
}

func withNodeName(name string, jobs []discoveryJob) []discoveryJob {
	for i := range jobs {
		jobs[i].Node = name
	}
	return jobs
}

func queryPeer(ctx context.Context, p *ckit.Peer, endpoint string, v interface{}) error {
	cc, err := grpc.Dial(p.ApplicationAddr, grpc.WithInsecure())
	if err != nil {
		return fmt.Errorf("failed to connect to peer %q: %w", p.Name, err)
	}

	cli := http.Client{Transport: httpclient.RoundTripper{Client: httpgrpc.NewHTTPClient(cc)}}
	url := fmt.Sprintf("http://%s%s", p.ApplicationAddr, endpoint)
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return fmt.Errorf("failed to request %s from peer %q: %w", endpoint, p.Name, err)
	}

	resp, err := cli.Do(req)
	if err != nil {
		return fmt.Errorf("failed to perform request %s against peer %q: %w", endpoint, p.Name, err)
	}
	return configapi.UnmarshalAPIResponse(resp.Body, v)
}

type discoveryJob struct {
	Node     string `json:"node_name"`
	Instance string `json:"instance_name"`
	Name     string `json:"job_name"`
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
