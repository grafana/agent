package apache

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"

	"github.com/grafana/agent/component/discovery"
	"github.com/grafana/agent/pkg/autodiscovery"
	"github.com/grafana/agent/pkg/river"
	"github.com/mitchellh/go-ps"
)

type Config struct {
	Binary     string   `river:"binary,attr"`
	ScrapeURIs []string `river:"scrape_uris,attr,optional"`
	Extensions []string `river:"ext,attr,optional"`
}

type Apache struct {
	binary     string
	scrapeURIs []string
	ext        []string
}

func (a *Apache) String() string {
	return "apache-http"
}

func New() (*Apache, error) {
	bb, err := os.ReadFile("pkg/autodiscovery/apache/apache.river")
	if err != nil {
		return nil, err
	}

	var cfg Config
	err = river.Unmarshal(bb, &cfg)
	if err != nil {
		return nil, err
	}

	return &Apache{
		binary:     cfg.Binary,
		scrapeURIs: cfg.ScrapeURIs,
		ext:        cfg.Extensions,
	}, nil
}

// Run check whether a Apache instance is running, and if so, returns a
// `prometheus.exporter.apache` component that can read metrics from it.
func (m *Apache) Run() (*autodiscovery.Result, error) {
	procs, err := ps.Processes()
	if err != nil {
		return nil, fmt.Errorf("could not read processes from host system: %w", err)
	}

	pid := -1
	for _, p := range procs {
		if p.Executable() == m.binary {
			pid = p.Pid()
			break
		}
	}
	if pid == -1 {
		return nil, fmt.Errorf("no running instance of process '%s' was found", m.binary)
	}

	// Apache is running on the host system, so we'll try to return _something_.
	res := &autodiscovery.Result{}
	var lsof autodiscovery.LSOF

	fns, err := autodiscovery.GetOpenFilenames(lsof, pid, m.ext...)
	if err != nil {
		return nil, err
	}
	for fn, _ := range fns {
		res.LogfileTargets = append(res.LogfileTargets,
			discovery.Target{"__path__": fn, "component": "apache"},
		)
	}

	// Let's try to use the configuration to connect using predefined URIs.
	for _, uri := range m.scrapeURIs {
		resp, err := http.Get(uri)
		if err != nil {
			continue
		}
		if !isRealServerStatusPage(resp) {
			continue
		}

		res.RiverConfig = fmt.Sprintf(`prometheus.exporter.apache "default" {
  scrape_uri = "%s"
}`, uri)
		res.MetricsExport = "prometheus.exporter.apache.default.targets"
		return res, nil
	}

	// Our predefined configurations didn't work; but MySQL is running.
	// Let's return a Flow component template for the user to fill out.
	res.RiverConfig = `prometheus.exporter.apache "default" {
    // NOTE: Agent Autodiscovery could not automatically configure an Apache exporter.
    // To set up an Apache exporter, please either set "scrape_uri" explicitly
    // or set up the AGENT_APACHE_SERVER_STATUS_URI environment variable and restart the Agent.
    scrape_uri = env("AGENT_APACHE_SERVER_STATUS_URI")
}`
	res.MetricsExport = "prometheus.exporter.apache.default.targets"

	return res, nil
}

func isRealServerStatusPage(httpResp *http.Response) bool {
	metrics := []string{"ServerVersion: ",
		"ServerMPM: ",
		"Server Built: ",
		"CurrentTime: ",
		"RestartTime: ",
		"ParentServerConfigGeneration: ",
		"ParentServerMPMGeneration: ",
		"ServerUptimeSeconds: ",
		"ServerUptime: ",
		"Load1: ",
		"Load5: ",
		"Load15: ",
		"Total Accesses: ",
		"Total kBytes: ",
		"Total Duration: ",
		"CPUUser: ",
		"CPUSystem: ",
		"CPUChildrenUser: ",
		"CPUChildrenSystem: ",
		"CPULoad: ",
		"Uptime: ",
		"ReqPerSec: ",
		"BytesPerSec: ",
		"BytesPerReq: ",
		"DurationPerReq: ",
		"BusyWorkers: ",
		"IdleWorkers: "}

	respBodyBytes, err := io.ReadAll(httpResp.Body)
	if err != nil {
		//TODO: Log the error?
		return false
	}
	respBody := string(respBodyBytes)

	for _, metric := range metrics {
		if !strings.Contains(respBody, metric) {
			return false
		}
	}
	return true
}
