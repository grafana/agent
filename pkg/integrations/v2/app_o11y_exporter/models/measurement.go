package models

import (
	"encoding/json"
	"errors"
	"fmt"
	"time"
)

type Measurement struct {
	Values    map[string]float64 `json:"values,omitempty"`
	Timestamp time.Time          `json:"timestamp,omitempty"`
	Type      string             `json:"type,omitempty"`
}

const (
	MTYPE_WEBVITALS = "web-vitals"
	MTYPE_CUSTOM    = "custom"
)

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
		return errors.New(fmt.Sprintf("Unknown measurement type '%s'", aux.Type))
	case MTYPE_CUSTOM:
		m.Type = MTYPE_CUSTOM
	case MTYPE_WEBVITALS:
		m.Type = MTYPE_WEBVITALS
	}

	return nil
}
