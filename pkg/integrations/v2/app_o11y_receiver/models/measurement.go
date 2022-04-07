package models

import (
	"fmt"
	"sort"
	"time"

	"github.com/grafana/agent/pkg/integrations/v2/app_o11y_receiver/utils"
)

// Measurement holds the data for user provided measurements
type Measurement struct {
	Values    map[string]float64 `json:"values,omitempty"`
	Timestamp time.Time          `json:"timestamp,omitempty"`
	Trace     TraceContext       `json:"trace,omitempty"`
}

// KeyVal representation of the exception object
func (m Measurement) KeyVal() *utils.KeyVal {
	kv := utils.NewKeyVal()

	utils.KeyValAdd(kv, "timestamp", m.Timestamp.String())
	utils.KeyValAdd(kv, "kind", "measurement")

	keys := make([]string, 0, len(m.Values))
	for k := range m.Values {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	for _, k := range keys {
		utils.KeyValAdd(kv, k, fmt.Sprintf("%f", m.Values[k]))
	}
	utils.MergeKeyVal(kv, m.Trace.KeyVal())
	return kv
}
