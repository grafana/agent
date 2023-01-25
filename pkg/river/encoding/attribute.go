package encoding

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/grafana/agent/pkg/river/internal/rivertags"
	"github.com/grafana/agent/pkg/river/internal/value"
)

// attributeField represents a JSON representation of a river attribute.
type attributeField struct {
	field    `json:",omitempty"`
	valField riverField
}

func newAttribute(val value.Value, f rivertags.Field) (*attributeField, error) {
	af := &attributeField{}
	return af, af.convertAttribute(val, f)
}

func (af *attributeField) hasValue() bool {
	if af == nil || af.valField == nil {
		return false
	}
	return af.valField.hasValue()
}

// MarshalJSON implements json marshaller.
func (af *attributeField) MarshalJSON() ([]byte, error) {
	if af.valField == nil {
		return nil, fmt.Errorf("the value of the attribute field is nil")
	}

	if af.valField.hasValue() {
		af.field.Value = af.valField
	} else {
		return nil, fmt.Errorf("attribute field did not have any valid values")
	}

	type temp attributeField
	return json.Marshal((*temp)(af))
}

func (af *attributeField) convertAttribute(val value.Value, f rivertags.Field) error {
	if !f.IsAttr() {
		return fmt.Errorf("convertAttribute requires a field that is an attribute got %T", val.Interface())
	}
	if !val.Reflect().IsValid() {
		return nil
	}
	af.Type = attrType
	af.Name = strings.Join(f.Name, ".")

	rv, err := convertRiverValue(val)
	if err != nil {
		return err
	}
	if !rv.hasValue() {
		return nil
	}
	af.valField = rv
	return nil
}
