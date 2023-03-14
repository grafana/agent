package runner

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"strings"
	"text/template"

	"github.com/grafana/agent/pkg/autodiscovery"
	"github.com/grafana/agent/pkg/autodiscovery/apache"
	"github.com/grafana/agent/pkg/autodiscovery/consul"
	"github.com/grafana/agent/pkg/autodiscovery/kubernetes"
	"github.com/grafana/agent/pkg/autodiscovery/mysql"
	"github.com/grafana/agent/pkg/autodiscovery/postgres"
	"github.com/grafana/agent/pkg/autodiscovery/prometheus"
	"github.com/grafana/agent/pkg/autodiscovery/redis"
	"github.com/grafana/agent/pkg/river/parser"
	"github.com/grafana/agent/pkg/river/printer"
)

type templateInput struct {
	Components     []string
	MetricsExports []string
	MetricsTargets []string
	LogsTargets    []string
}

type AutodiscT string

const (
	AUTODISC_MYSQL      AutodiscT = "mysql"
	AUTODISC_POSTGRES   AutodiscT = "postgres"
	AUTODISC_CONSUL     AutodiscT = "consul"
	AUTODISC_PROMETHEUS AutodiscT = "prometheus"
	AUTODISC_KUBERNETES AutodiscT = "kubernetes"
	AUTODISC_REDIS      AutodiscT = "redis"
	AUTODISC_APACHE     AutodiscT = "apache"
)

var allMechanisms = []AutodiscT{
	AUTODISC_MYSQL,
	AUTODISC_POSTGRES,
	AUTODISC_CONSUL,
	AUTODISC_PROMETHEUS,
	AUTODISC_KUBERNETES,
	AUTODISC_REDIS,
	AUTODISC_APACHE,
}

type Autodiscovery struct {
	// Discoverers which we were explicitly told to ignore, e.g. via Agent Management
	IgnoreList map[AutodiscT]struct{}
}

func createMechanism(discovererId AutodiscT) (autodiscovery.Mechanism, error) {
	switch discovererId {
	case AUTODISC_MYSQL:
		return mysql.New()
	case AUTODISC_POSTGRES:
		return postgres.New()
	case AUTODISC_CONSUL:
		return consul.New()
	case AUTODISC_PROMETHEUS:
		return prometheus.New()
	case AUTODISC_KUBERNETES:
		return kubernetes.New()
	case AUTODISC_REDIS:
		return redis.New()
	case AUTODISC_APACHE:
		return apache.New()
	}
	return nil, fmt.Errorf("unknown discoverer")
}

// Do do doo do doo.
func (a *Autodiscovery) Do(wr io.Writer) []AutodiscT {
	// "mechanisms" are the discoverers which we need to run.
	// Usually these are all the available discoverers
	// bar the ones in the IgnoreList and the ones for exporters
	// already used in the River config.
	usedMechanisms := make([]AutodiscT, 0)
	var mechanisms []autodiscovery.Mechanism
	for _, mechId := range allMechanisms {
		if _, ok := a.IgnoreList[mechId]; ok {
			continue
		}
		mech, err := createMechanism(mechId)
		if err != nil {
			fmt.Fprintf(os.Stderr,
				"failed to create a %s auto discovery mechanism: %s\n", mechId, err)
			continue
		}
		mechanisms = append(mechanisms, mech)
		//TODO: "usedMechanisms" should not include mechanisms which failed to init or to run
		usedMechanisms = append(usedMechanisms, mechId)
	}

	results := make([]*autodiscovery.Result, 0, len(mechanisms))
	for _, f := range mechanisms {
		res, err := f.Run()
		if err != nil {
			//TODO: Also print out the name of the mechanism
			fmt.Fprintf(os.Stderr,
				"failed to run auto discovery mechanism: %s\n", err)
			continue
		}
		results = append(results, res)
	}

	input := BuildTemplateInput(results)

	//TODO: Check RenderConfig for errors?
	// We don't have to return it. Maybe log a warning and continue silently?
	RenderConfig(wr, input)

	//TODO: If the agent already has a River config, can we merge this new one and the existing one?

	// Reasons not to use a mechanism:
	// * in the ignore list
	// * already in the user's configuration
	// * failed to create
	// * failed to run
	return usedMechanisms
}

// BuildTemplateInput ...
func BuildTemplateInput(input []*autodiscovery.Result) templateInput {
	res := templateInput{}

	for _, r := range input {
		res.Components = append(res.Components, r.RiverConfig)
		if r.MetricsExport != "" {
			res.MetricsExports = append(res.MetricsExports, r.MetricsExport)
		}
		if len(r.MetricsTargets) > 0 {
			for _, mt := range r.MetricsTargets {
				res.MetricsTargets = append(res.MetricsTargets, mt.RiverString())
			}
		}
		if len(r.LogfileTargets) > 0 {
			for _, lt := range r.LogfileTargets {
				res.LogsTargets = append(res.LogsTargets, lt.RiverString())
			}
		}
	}

	return res
}

func RenderConfig(wr io.Writer, input templateInput) error {
	tmpl := template.New("cfg.river")
	tmpl = template.Must(tmpl.Parse(templateStr))

	rawBuf := new(bytes.Buffer)
	tmpl.Execute(rawBuf, input)

	return PretifyRiver(wr, rawBuf.Bytes())
}

// TODO: The main formatting logic was copied form riverfmt. Place this
// function in a shared package, and have riverfmt use it from the same shared place.
func PretifyRiver(wr io.Writer, riverCfg []byte) error {
	ast, err := parser.ParseFile("", riverCfg)
	if err != nil {
		return err
	}

	var prettyBuf bytes.Buffer
	if err := printer.Fprint(&prettyBuf, ast); err != nil {
		return err
	}

	wr.Write(prettyBuf.Bytes())
	return nil
}

var templateStr = `
prometheus.scrape "default" {
  targets = concat({{if .MetricsExports }}{{range .MetricsExports}}
  {{if .}}  {{.}},{{end}}{{end}}{{end}}
  {{if .MetricsTargets }} [{{range .MetricsTargets}}
      {{.}},{{end}}
    ],{{end}}
  )
  forward_to = [prometheus.remote_write.default.receiver]
}

prometheus.remote_write "default" {
  endpoint {
    url = env("GRAFANACLOUD_METRICS_URL")

    basic_auth {
      username = env("GRAFANACLOUD_METRICS_USER")
      password = env("GRAFANACLOUD_APIKEY")
    }
  }
}

{{if .LogsTargets }}
loki.source.file "default" {
  targets = [{{range .LogsTargets}}
      {{.}},{{end}}
  ]

  forward_to = [loki.write.default.receiver]
}

loki.write "default" {
  endpoint {
    url = env("GRAFANACLOUD_LOGS_URL")
    basic_auth {
      username = env("GRAFANACLOUD_LOGS_USER")
      password = env("GRAFANACLOUD_APIKEY")
    }
  }
}
{{end}}

{{ range .Components}} {{if .}}
{{.}}
{{ end }}{{end}}
`

func InstallIntegrations(apiToken string, integrations ...string) error {
	client := &http.Client{}
	baseURL := "https://grafana.com/api"
	integrationsAPIURL := "https://integrations-api-eu-west.grafana.net"

	// Get stack ID
	req, err := http.NewRequest("GET", baseURL+"/instances/tpaschalis", nil)
	if err != nil {
		return err
	}
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", apiToken))
	req.Header.Set("User-Agent", "deployment_tools:/scripts/gcom")
	res, _ := client.Do(req)
	bb, _ := ioutil.ReadAll(res.Body)

	var rsp map[string]interface{}
	_ = json.Unmarshal(bb, &rsp)
	instanceID := rsp["id"]
	grafanaURL := rsp["url"].(string)
	instanceURL := fmt.Sprintf("%s/instances/%d", baseURL, int(instanceID.(float64)))
	integrationsInstanceURL := fmt.Sprintf("%s/v2/stacks/%d", integrationsAPIURL, int(instanceID.(float64)))

	// Generate an API Key for the hosted Grafana instance
	// TODO(@tpaschalis): Remove key afterwards?
	body := []byte(`{"name": "autodiscovery-install-integrations22", "role": "admin", "secondsToLive": 300}`)
	req, err = http.NewRequest("POST", instanceURL+"/api/auth/keys", bytes.NewReader(body))
	if err != nil {
		return err
	}
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", apiToken))
	req.Header.Set("User-Agent", "deployment_tools:/scripts/gcom")
	req.Header.Set("Content-Type", "application/json")
	res, _ = client.Do(req)
	bb, _ = ioutil.ReadAll(res.Body)
	_ = json.Unmarshal(bb, &rsp)
	grafanaAPIKey := rsp["key"]
	_ = grafanaAPIKey

	allDashboardData := make([]map[string]interface{}, 0)

	// Get dashboard infos for required integrations.
	for _, integration := range integrations {
		req, err = http.NewRequest("GET", integrationsInstanceURL+"/integrations/"+integration+"/dashboards", nil)
		if err != nil {
			return err
		}
		req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", apiToken))
		res, _ = client.Do(req)
		bb, _ = ioutil.ReadAll(res.Body)
		_ = json.Unmarshal(bb, &rsp)
		dashboardData := rsp["data"].([]interface{})
		for _, dd := range dashboardData {
			allDashboardData = append(allDashboardData, dd.(map[string]interface{}))
		}
	}

	// Create all required folders.
	for _, dd := range allDashboardData {
		// folderName := dd["dashboard_folder"].(string)
		folderName := dd["folder_name"].(string)
		uid := strings.Replace(folderName, " ", "-", -1)

		body := []byte(fmt.Sprintf(`{"title": "%s", "uid": "%s"}`, folderName, uid))
		req, err = http.NewRequest("POST", grafanaURL+"/api/folders", bytes.NewReader(body))
		if err != nil {
			return err
		}
		req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", grafanaAPIKey))
		req.Header.Set("User-Agent", "deployment_tools:/scripts/gcom")
		req.Header.Set("Content-Type", "application/json")
		res, _ = client.Do(req)
		bb, _ = ioutil.ReadAll(res.Body)
	}

	// Install all dashboards
	for _, dd := range allDashboardData {
		dashboardJSON, err := json.Marshal(dd["dashboard"])
		if err != nil {
			return err
		}
		folderName := dd["folder_name"].(string)
		uid := strings.Replace(folderName, " ", "-", -1)
		overwrite := dd["overwrite"]

		body := []byte(fmt.Sprintf(`{"dashboard": %s, "folderUid": "%s", "overwrite": %t, "message": "creating dashboard from the Cloud Connections plugin"}`, dashboardJSON, uid, overwrite))
		req, err = http.NewRequest("POST", grafanaURL+"/api/dashboards/db", bytes.NewReader(body))
		if err != nil {
			return err
		}
		req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", grafanaAPIKey))
		req.Header.Set("User-Agent", "deployment_tools:/scripts/gcom")
		req.Header.Set("Content-Type", "application/json")
		res, _ = client.Do(req)
		bb, _ = ioutil.ReadAll(res.Body)
	}

	// Install all integrations
	for _, integration := range integrations {
		req, err = http.NewRequest("POST", integrationsInstanceURL+"/integrations/"+integration+"/install", nil)
		if err != nil {
			return err
		}
		req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", apiToken))
		req.Header.Set("User-Agent", "deployment_tools:/scripts/gcom")
		req.Header.Set("Content-Type", "application/json")
		res, _ = client.Do(req)
		bb, _ = ioutil.ReadAll(res.Body)
	}

	fmt.Println("All done! Navigate to the following link to see your new Grafana Cloud integrations in action! :tada:")
	fmt.Println(grafanaURL + "/dashboards")

	return nil
}
