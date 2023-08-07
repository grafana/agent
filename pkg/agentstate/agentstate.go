package agentstate

import (
	"encoding/json"
	"time"
)

type AgentState struct {
	ID     string
	Labels map[string]string
}

type Component struct {
	ID              string            `parquet:"id"`
	ModuleID        string            `parquet:"module_id"`
	Health          Health            `parquet:"health"`
	ComponentDetail []ComponentDetail `parquet:"component_detail"`
}

type Health struct {
	Health     string    `parquet:"state"`
	Message    string    `parquet:"message"`
	UpdateTime time.Time `parquet:"update_time"`
}

type ComponentDetail struct {
	ID         uint            `parquet:"id,delta"`
	ParentID   uint            `parquet:"parent_id,delta"`
	Name       string          `parquet:"name,dict"`
	Label      string          `parquet:"label,dict"`
	RiverType  string          `parquet:"river_type,dict"`
	RiverValue json.RawMessage `parquet:"river_value,json"`
}
