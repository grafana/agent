// package useragent provides a consistent way to get a user agent for outbound http requests from Grafana Agent.
// The default User-Agent is `GrafanaAgent/$VERSION($MODE)`
// Where version is the build version of the agent and MODE is one of "static" or "flow".
package useragent

import (
	"fmt"
	"os"
	"runtime"
	"strings"

	"github.com/grafana/agent/pkg/build"
)

const (
	deployModeEnv = "AGENT_DEPLOY_MODE"
	modeEnv       = "AGENT_MODE"
)

// settable by tests
var goos = runtime.GOOS

func Get() string {
	parenthesis := ""
	metadata := []string{}
	if mode := getRunMode(); mode != "" {
		metadata = append(metadata, mode)
	}
	metadata = append(metadata, goos)
	if op := getDeployMode(); op != "" {
		metadata = append(metadata, op)
	}
	if len(metadata) > 0 {
		parenthesis = fmt.Sprintf(" (%s)", strings.Join(metadata, "; "))
	}
	return fmt.Sprintf("GrafanaAgent/%s%s", build.Version, parenthesis)
}

// getRunMode attempts to get agent mode, using `unknown` for invalid values.
func getRunMode() string {
	key := os.Getenv(modeEnv)
	switch key {
	case "flow":
		return "flow"
	case "static", "":
		return "static"
	default:
		return "unknown"
	}
}

func getDeployMode() string {
	op := os.Getenv(deployModeEnv)
	// only return known modes. Use "binary" as a default catch-all.
	switch op {
	case "operator", "helm", "docker", "deb", "rpm", "brew":
		return op
	}
	return "binary"
}
