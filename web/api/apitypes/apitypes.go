package apitypes

import (
	"encoding/json"
	"time"
)

// ComponentInfo represents a component in flow.
type ComponentInfo struct {
	Name         string           `json:"name,omitempty"`
	Type         string           `json:"type,omitempty"`
	ID           string           `json:"id,omitempty"`
	Label        string           `json:"label,omitempty"`
	References   []string         `json:"referencesTo"`
	ReferencedBy []string         `json:"referencedBy"`
	Health       *ComponentHealth `json:"health"`
	Original     string           `json:"original"`
	Arguments    json.RawMessage  `json:"arguments,omitempty"`
	Exports      json.RawMessage  `json:"exports,omitempty"`
	DebugInfo    json.RawMessage  `json:"debugInfo,omitempty"`
}

// ComponentHealth represents the health of a component.
type ComponentHealth struct {
	State       string    `json:"state"`
	Message     string    `json:"message"`
	UpdatedTime time.Time `json:"updatedTime"`
}
