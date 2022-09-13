package encoding

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/grafana/agent/pkg/river/internal/rivertags"
	"github.com/grafana/agent/pkg/river/internal/value"
)

// AttributeField represents a json representation of a river attribute.
type AttributeField struct {
	Field `json:",omitempty"`
	field RiverValue
}

func newAttribute(val value.Value, f rivertags.Field) (*AttributeField, error) {
	af := &AttributeField{}
	return af, af.convertAttribute(val, f)
}

func (af *AttributeField) isValid() bool {
	if af == nil || af.field == nil {
		return false
	}
	return af.field.hasValue()
}

// MarshalJSON implements json marshaller.
func (af *AttributeField) MarshalJSON() ([]byte, error) {
	if af.field.hasValue() {
		af.Field.Value = af.field
	} else {
		return nil, fmt.Errorf("attribute field did not have any valid values")
	}

	type temp AttributeField
	return json.Marshal((*temp)(af))
}

func (af *AttributeField) convertAttribute(val value.Value, f rivertags.Field) error {
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
	af.field = rv
	return nil
}
