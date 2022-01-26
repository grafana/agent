package models

import (
	"fmt"
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
	utils.MergeKeyVal(kv, m.Trace.KeyVal())
	for k, v := range m.Values {
		utils.KeyValAdd(kv, k, fmt.Sprintf("%f", v))
	}
	return kv
}
