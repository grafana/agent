package models

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/grafana/agent/pkg/integrations/v2/app_o11y_exporter/utils"
)

// Measurement holds the data for user provided measurements
type Measurement struct {
	Values    map[string]float64 `json:"values,omitempty"`
	Timestamp time.Time          `json:"timestamp,omitempty"`
	Type      string             `json:"type,omitempty"`
}

const (
	// MTypeWebVitals type for web vitals metrics
	MTypeWebVitals = "web-vitals"
	// MTypeCustom for custom metrics
	MTypeCustom = "custom"
)

// UnmarshalJSON implements the Unmarshaller interface
func (m *Measurement) UnmarshalJSON(data []byte) error {
	type MAlias Measurement
	aux := &struct{ *MAlias }{MAlias: (*MAlias)(m)}
	if err := json.Unmarshal(data, &aux); err != nil {
		return err
	}

	m.Values = aux.Values
	m.Timestamp = aux.Timestamp

	switch aux.Type {
	default:
		return fmt.Errorf("Unknown measurement type '%s'", aux.Type)
	case MTypeCustom:
		m.Type = MTypeCustom
	case MTypeWebVitals:
		m.Type = MTypeWebVitals
	}

	return nil
}

// KeyVal representation of the exception object
func (m Measurement) KeyVal() *utils.KeyVal {
	kv := utils.NewKeyVal()

	utils.KeyValAdd(kv, "timestamp", m.Timestamp.String())
	utils.KeyValAdd(kv, "kind", "measurement")
	utils.KeyValAdd(kv, "type", m.Type)
	for k, v := range m.Values {
		utils.KeyValAdd(kv, k, fmt.Sprintf("%f", v))
	}
	return kv
}
