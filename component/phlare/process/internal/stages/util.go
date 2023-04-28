package stages

// This package is ported over from grafana/loki/clients/pkg/logentry/stages.
// We aim to port the stages in steps, to avoid introducing huge amounts of
// new code without being able to slowly review, examine and test them.

import (
	"fmt"
	"strconv"
)

var (
	// Debug is used to wrap debug log statements, the go-kit logger won't let us introspect the current log level
	// so this global is used for that purpose. This allows us to skip allocations of log messages at the
	// debug level when debug level logging is not enabled. Log level allocations can become very expensive
	// as we log numerous log entries per log line at debug level.
	Debug = false

	// Inspect is used to debug promtail pipelines by showing diffs between pipeline stages
	Inspect = false
)

const (
	ErrTimestampContainsYear = "timestamp '%s' is expected to not contain the year date component"
)

// getString will convert the input variable to a string if possible
func getString(unk interface{}) (string, error) {
	switch i := unk.(type) {
	case float64:
		return strconv.FormatFloat(i, 'f', -1, 64), nil
	case float32:
		return strconv.FormatFloat(float64(i), 'f', -1, 32), nil
	case int64:
		return strconv.FormatInt(i, 10), nil
	case int32:
		return strconv.FormatInt(int64(i), 10), nil
	case int:
		return strconv.Itoa(i), nil
	case uint64:
		return strconv.FormatUint(i, 10), nil
	case uint32:
		return strconv.FormatUint(uint64(i), 10), nil
	case uint:
		return strconv.FormatUint(uint64(i), 10), nil
	case string:
		return unk.(string), nil
	case bool:
		if i {
			return "true", nil
		}
		return "false", nil
	default:
		return "", fmt.Errorf("Can't convert %v to string", unk)
	}
}

func stringsContain(values []string, search string) bool {
	for _, v := range values {
		if search == v {
			return true
		}
	}

	return false
}
