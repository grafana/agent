package agentstate

import (
	"encoding/json"
	"time"
)

type AgentState struct {
	Labels map[string]string
}

func NewAgentState(labels map[string]string) AgentState {
	return AgentState{
		Labels: labels,
	}
}

type Component struct {
	ID        string            `parquet:"id"`
	ModuleID  string            `parquet:"module_id"`
	Health    Health            `parquet:"health"`
	Arguments []ComponentDetail `parquet:"arguments"`
	Exports   []ComponentDetail `parquet:"exports"`
	DebugInfo []ComponentDetail `parquet:"debug_info"`
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
	Label      string          `parquet:"label,optional"`
	RiverType  string          `parquet:"river_type,dict"`
	RiverValue json.RawMessage `parquet:"river_value,json"`
}
