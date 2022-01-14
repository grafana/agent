package metrics

import (
	"context"
	"fmt"
	"net/http"
	"sort"
	"time"

	"github.com/gorilla/mux"
	"github.com/grafana/agent/pkg/metrics/cluster/configapi"
	"github.com/prometheus/common/model"
	"github.com/prometheus/prometheus/pkg/labels"
	"github.com/rfratto/ckit"
	"github.com/rfratto/ckit/httpgrpc"
	"google.golang.org/grpc"
)

const (
	discoveryJobsAPIEndpoint    = "/agent/api/v1/metrics/discovery/jobs"
	discoveryTargetsAPIEndpoint = "/agent/api/v1/metrics/discovery/targets"
	scrapeTargetsAPIEndpoint    = "/agent/api/v1/metrics/scrape/targets"
)

// WireAPI wires up HTTP API routes for the metrics subsystem.
func (m *Metrics) WireAPI(r *mux.Router) {
	r.HandleFunc(discoveryJobsAPIEndpoint, m.listDiscoveryJobsHandler).Methods(http.MethodGet)
	r.HandleFunc(discoveryTargetsAPIEndpoint, m.listDiscoveryTargetsHandler).Methods(http.MethodGet)
	r.HandleFunc(scrapeTargetsAPIEndpoint, m.scrapeTargetsHandler).Methods(http.MethodGet)
}

type discoveryJob struct {
	Node     string `json:"node_name,omitempty"`
	Instance string `json:"instance_name"`
	Name     string `json:"job_name"`
}

func (m *Metrics) listDiscoveryJobsHandler(w http.ResponseWriter, r *http.Request) {
	var (
		peers     = m.node.Peers()
		localOnly = r.URL.Query().Get("remote") == "0"
	)

	withNodeName := func(name string, jobs []discoveryJob) []discoveryJob {
		for i := range jobs {
			jobs[i].Node = name
		}
		return jobs
	}

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
			_ = configapi.WriteError(w, http.StatusInternalServerError, err)
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

func queryPeer(ctx context.Context, p *ckit.Peer, endpoint string, v interface{}) error {
	cc, err := grpc.Dial(p.ApplicationAddr, grpc.WithInsecure())
	if err != nil {
		return fmt.Errorf("failed to connect to peer %q: %w", p.Name, err)
	}

	cli := http.Client{Transport: httpgrpc.ClientTransport(cc)}
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

type discoveryTarget struct {
	Node        string         `json:"node,omitempty"`
	Instance    string         `json:"instance"`
	TargetGroup string         `json:"target_group"`
	Labels      model.LabelSet `json:"labels"`
}

func (m *Metrics) listDiscoveryTargetsHandler(w http.ResponseWriter, r *http.Request) {
	var (
		peers     = m.node.Peers()
		localOnly = r.URL.Query().Get("remote") == "0"
	)

	withNodeName := func(name string, targets []discoveryTarget) []discoveryTarget {
		for i := range targets {
			targets[i].Node = name
		}
		return targets
	}

	var targets []discoveryTarget

	for _, p := range peers {
		if p.Self {
			targets = append(targets, withNodeName(p.Name, m.discoverers.getDiscoveryTargets())...)
			continue
		}
		if localOnly {
			// Skip over remote peers when the query is only for the local node.
			continue
		}

		var peerTargets []discoveryTarget
		err := queryPeer(r.Context(), &p, discoveryTargetsAPIEndpoint+"?remote=0", &peerTargets)
		if err != nil {
			_ = configapi.WriteError(w, http.StatusInternalServerError, err)
			return
		}
		targets = append(targets, peerTargets...)
	}

	// Sort slice by instance, target group, and address label. Node is not used
	// for sorting.
	sort.Slice(targets, func(i, j int) bool {
		switch {
		case targets[i].Instance != targets[j].Instance:
			return targets[i].Instance < targets[j].Instance
		case targets[i].TargetGroup != targets[j].TargetGroup:
			return targets[i].TargetGroup < targets[j].TargetGroup
		default:
			return targets[i].Labels[model.AddressLabel] < targets[j].Labels[model.AddressLabel]
		}
	})

	_ = configapi.WriteResponse(w, http.StatusOK, targets)
}

type scrapeTarget struct {
	Node        string `json:"node,omitempty"`
	Instance    string `json:"instance"`
	TargetGroup string `json:"target_group"`

	Endpoint         string        `json:"endpoint"`
	State            string        `json:"state"`
	Labels           labels.Labels `json:"labels"`
	DiscoveredLabels labels.Labels `json:"discovered_labels"`
	LastScrape       *time.Time    `json:"last_scrape"`
	ScrapeDuration   int64         `json:"scrape_duration_ms"`
	ScrapeError      string        `json:"scrape_error"`
}

func (m *Metrics) scrapeTargetsHandler(w http.ResponseWriter, r *http.Request) {
	var (
		peers     = m.node.Peers()
		localOnly = r.URL.Query().Get("remote") == "0"
	)

	withNodeName := func(name string, targets []scrapeTarget) []scrapeTarget {
		for i := range targets {
			targets[i].Node = name
		}
		return targets
	}

	var targets []scrapeTarget

	for _, p := range peers {
		if p.Self {
			targets = append(targets, withNodeName(p.Name, m.scrapers.getScrapeTargets())...)
			continue
		}
		if localOnly {
			// Skip over remote peers when the query is only for the local node.
			continue
		}

		var peerTargets []scrapeTarget
		err := queryPeer(r.Context(), &p, scrapeTargetsAPIEndpoint+"?remote=0", &peerTargets)
		if err != nil {
			_ = configapi.WriteError(w, http.StatusInternalServerError, err)
			return
		}
		targets = append(targets, peerTargets...)
	}

	// Sort slice by instance, target group, and job label, and instance label.
	// Node name is not used for sorting.
	sort.Slice(targets, func(i, j int) bool {
		switch {
		case targets[i].Instance != targets[j].Instance:
			return targets[i].Instance < targets[j].Instance
		case targets[i].TargetGroup != targets[j].TargetGroup:
			return targets[i].TargetGroup < targets[j].TargetGroup
		case targets[i].Labels.Get(model.JobLabel) != targets[j].Labels.Get(model.JobLabel):
			return targets[i].Labels.Get(model.JobLabel) < targets[j].Labels.Get(model.JobLabel)
		default:
			return targets[i].Labels.Get(model.InstanceLabel) < targets[j].Labels.Get(model.InstanceLabel)
		}
	})

	_ = configapi.WriteResponse(w, http.StatusOK, targets)
}
