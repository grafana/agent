package encoding

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/grafana/agent/pkg/river/internal/rivertags"
	"github.com/grafana/agent/pkg/river/internal/value"
)

// attributeField represents a json representation of a river attribute.
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
	if val.Reflect().IsZero() {
		return nil
	}
	af.Type = attr
	af.Name = strings.Join(f.Name, ".")

	rv, err := convertRiverValue(val)
	if err != nil {
		return err
	}
	if !rv.hasValue() {
		return fmt.Errorf("unable to find value for %T in convertAttribute", val.Interface())
	}
	af.valField = rv
	return nil
}
