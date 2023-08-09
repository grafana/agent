package riveragentstate

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

// Various concrete types used to marshal River values.
type (
	// jsonStatement is a statement within a River body.
	jsonStatement interface{ isStatement() }

	// A jsonBody is a collection of statements.
	jsonBody = []jsonStatement

	// jsonBlock represents a River block as JSON. jsonBlock is a jsonStatement.
	jsonBlock struct {
		Name  string          `json:"name"`
		Type  string          `json:"type"` // Always "block"
		Label string          `json:"label,omitempty"`
		Body  []jsonStatement `json:"body"`
	}

	// jsonAttr represents a River attribute as JSON. jsonAttr is a
	// jsonStatement.
	jsonAttr struct {
		Name  string    `json:"name"`
		Type  string    `json:"type"` // Always "attr"
		Value jsonValue `json:"value"`
	}

	// jsonValue represents a single River value as JSON.
	jsonValue struct {
		Type  string      `json:"type"`
		Value interface{} `json:"value"`
	}

	// jsonObjectField represents a field within a River object.
	jsonObjectField struct {
		Key   string      `json:"key"`
		Value interface{} `json:"value"`
	}
)

func (jsonBlock) isStatement() {}
func (jsonAttr) isStatement()  {}
