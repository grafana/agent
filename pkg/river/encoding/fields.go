package encoding

// field represents a value in river.
type field struct {
	Name  string      `json:"name,omitempty"`
	Type  string      `json:"type,omitempty"`
	Value interface{} `json:"value,omitempty"`
}

// valueField represents a value in river.
type valueField struct {
	Type  string      `json:"type,omitempty"`
	Value interface{} `json:"value,omitempty"`
}

func (vf *valueField) hasValue() bool {
	if vf == nil {
		return false
	}
	return vf.Value != nil
}

// keyField represents a map backed field.
type keyField struct {
	field `json:",omitempty"`
	Key   string `json:"key,omitempty"`
}
