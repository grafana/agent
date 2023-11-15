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
	operatorEnv = "AGENT_OPERATOR"
	modeEnv     = "AGENT_MODE"
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
	if op := getOperator(); op != "" {
		metadata = append(metadata, op)
	}
	if len(metadata) > 0 {
		parenthesis = fmt.Sprintf(" (%s)", strings.Join(metadata, "; "))
	}
	return fmt.Sprintf("GrafanaAgent/%s%s", build.Version, parenthesis)
}

// getRunMode attempts to get agent mode, using `unknown` for invalid values.
func getRunMode() string {
	key, found := os.LookupEnv(modeEnv)
	if !found {
		return "static"
	}
	switch key {
	case "flow":
		return "flow"
	case "static", "":
		return "static"
	default:
		return "unknown"
	}
}

func getOperator() string {
	op := os.Getenv(operatorEnv)
	if op == "1" {
		return "operator"
	}
	return ""
}
