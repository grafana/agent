package usagestats

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"runtime"
	"time"

	"github.com/prometheus/common/version"
)

var (
	httpClient    = http.Client{Timeout: 5 * time.Second}
	usageStatsURL = "https://stats.grafana.org/agent-usage-report"
)

// Report is the payload to be sent to stats.grafana.org
type Report struct {
	UsageStatsID   string                 `json:"usageStatsId"`
	CreatedAt      time.Time              `json:"createdAt"`
	Interval       time.Time              `json:"interval"`
	Version        string                 `json:"version"`
	Metrics        map[string]interface{} `json:"metrics"`
	Os             string                 `json:"os"`
	Arch           string                 `json:"arch"`
	DeploymentMode string                 `json:"deploymentMode"`
}

func sendReport(ctx context.Context, seed *AgentSeed, interval time.Time, metrics map[string]interface{}) error {
	report := Report{
		UsageStatsID:   seed.UID,
		CreatedAt:      seed.CreatedAt,
		Version:        version.Version,
		Os:             runtime.GOOS,
		Arch:           runtime.GOARCH,
		Interval:       interval,
		Metrics:        metrics,
		DeploymentMode: getDeployMode(),
	}
	out, err := json.MarshalIndent(report, "", " ")
	if err != nil {
		return err
	}
	req, err := http.NewRequest(http.MethodPost, usageStatsURL, bytes.NewBuffer(out))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := httpClient.Do(req.WithContext(ctx))
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		data, err := io.ReadAll(resp.Body)
		if err != nil {
			return err
		}
		return fmt.Errorf("failed to send usage stats: %s  body: %s", resp.Status, string(data))
	}
	return nil
}

func getDeployMode() string {
	op := os.Getenv("AGENT_DEPLOY_MODE")
	// only return known modes. Use "binary" as a default catch-all.
	switch op {
	case "operator", "helm", "docker", "deb", "rpm", "brew":
		return op
	}
	return "binary"
}
