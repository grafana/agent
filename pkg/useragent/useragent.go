// package useragent provides a consistent way to get a user agent for outbound http requests from Grafana Agent.
// The default User-Agent is `GrafanaAgent/$VERSION($MODE)`
// Where version is the build version of the agent and MODE is one of "static" or "flow".
package useragent

import (
	"fmt"
	"os"

	"github.com/grafana/agent/pkg/build"
)

func UserAgent() string {
	parenthesis := ""
	if mode := getRunMode(); mode != "" {
		parenthesis = fmt.Sprintf(" (%s)", mode)
	}
	return fmt.Sprintf("GrafanaAgent/%s%s", build.Version, parenthesis)
}

// getRunMode attempts to get agent mode.
// if an unknown value is found we will simply omit it.
func getRunMode() string {
	key, found := os.LookupEnv("AGENT_MODE")
	if !found {
		return "static"
	}
	switch key {
	case "flow":
		return "flow"
	case "static", "":
		return "static"
	default:
		return ""
	}
}
