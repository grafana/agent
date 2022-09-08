package encoding

// Field represents a value in river.
type Field struct {
	Name  string      `json:"name,omitempty"`
	Type  string      `json:"type,omitempty"`
	Value interface{} `json:"value,omitempty"`
}

// ValueField represents a value in river.
type ValueField struct {
	Type  string      `json:"type,omitempty"`
	Value interface{} `json:"value,omitempty"`
}

func (vf *ValueField) hasValue() bool {
	if vf == nil {
		return false
	}
	return vf.Value != nil
}

// KeyField represents a map backed field.
type KeyField struct {
	Field `json:",omitempty"`
	Key   string `json:"key,omitempty"`
}
