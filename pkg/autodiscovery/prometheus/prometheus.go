package prometheus

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"strings"

	"github.com/google/uuid"
	"github.com/grafana/agent/component/discovery"
	"github.com/grafana/agent/pkg/autodiscovery"
	"github.com/grafana/agent/pkg/river"
	"github.com/prometheus/client_golang/prometheus/testutil/promlint"
)

// Config holds the predefined endpoints we look for running exporters at.
// If `metrics_prefix` is set, we also try to verify that the metrics have at
// least one of the prefixes attached to them; this can be useful to verify
// that we have auto-discovered the correct exporter and not some other that is
// running on a custom port.
//
// https://github.com/prometheus/prometheus/wiki/Default-port-allocations
type Config struct {
	Exporters []Exporter `river:"exporter,block,optional"`
}

// Exporter defines a single exporter we look for.
type Exporter struct {
	URL      string   `river:"url,attr"`
	Name     string   `river:"name,attr,optional"`
	Prefixes []string `river:"metrics_prefix,attr,optional"`
}

// Prometheus is an autodiscovery mechanism for Prometheus exporters.
// NOTE(@tpaschalis) Looks like for these mechanisms we could use Config
// directly to implement this interface. I'm keeping this intermediate layer in
// for now, just in case we need to carry anything else around.
type Prometheus struct {
	Exporters []Exporter
}

func (p *Prometheus) String() string {
	return "prometheus"
}

// New creates a new auto-discovery Postgres mechanism instance.
func New() (*Prometheus, error) {
	bb, err := os.ReadFile("pkg/autodiscovery/prometheus/prometheus.river")
	if err != nil {
		return nil, err
	}

	var cfg Config
	err = river.Unmarshal(bb, &cfg)
	if err != nil {
		return nil, err
	}

	return &Prometheus{
		Exporters: cfg.Exporters,
	}, nil
}

// Run check whether a Postgres instance is running, and if so, returns a
// `prometheus.exporter.postgres` component that can read metrics from it.
func (pg *Prometheus) Run() (*autodiscovery.Result, error) {
	var httpClient http.Client

	res := &autodiscovery.Result{}

	for _, exp := range pg.Exporters {
		rsp, err := httpClient.Get(exp.URL)
		if err != nil {
			continue
		}
		defer rsp.Body.Close()

		// Oh, so there _is_ something running on that port.
		bb, err := ioutil.ReadAll(rsp.Body)
		if err != nil {
			fmt.Fprintln(os.Stderr, "error reading response body:", exp.URL, err)
			continue
		}

		l := promlint.New(bytes.NewReader(bb))
		problems, err := l.Lint()
		if err != nil {
			// fmt.Fprintln(os.Stderr, "error while linting:", exp.URL, err)
			continue
		}

		// Commenting this out for now as the Agent's own metrics report a problem.
		// {agent_component_controller_running_components_total non-counter metrics should not have "_total" suffix}
		// {agent_resources_process_start_time_seconds counter metrics should have "_total" suffix}]
		//
		// if len(problems) > 0 {
		// 	fmt.Fprintln(os.Stderr, "the endpoint's metrics had some problems", problems)
		// }
		_ = problems

		// This endpoint correctly exposes Prometheus metrics, let's scrape it.

		// If a list of prefixes was defined, look for it.
		// The following is a bit weak, but can work as a start.
		// Should we parse the response and specifically only keep metrics that
		// start with one of the defined prefixes?
		// p, err := textparse.New(); for et, err := p.Next; ...
		if len(exp.Prefixes) > 0 {
			prefixMatch := false
			for _, prefix := range exp.Prefixes {
				if strings.Contains(string(bb), prefix) {
					prefixMatch = true
				}
			}
			if !prefixMatch {
				continue
			}
		}

		if exp.Name == "" {
			exp.Name = uuid.NewString()[:6]
		}

		res.MetricsTargets = append(res.MetricsTargets,
			discovery.Target{"__address__": exp.URL, "component": exp.Name},
		)

	}
	return res, nil
}
