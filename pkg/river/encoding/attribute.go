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
	Field      `json:",omitempty"`
	valueField *ValueField
	arrayField *ArrayField
	mapField   *MapField
}

func newAttribute(val value.Value, f rivertags.Field) (*AttributeField, error) {
	af := &AttributeField{}
	return af, af.convertAttribute(val, f)
}

func (af *AttributeField) isValid() bool {
	if af == nil {
		return false
	}
	if af.mapField != nil {
		return af.mapField.hasValue()
	} else if af.arrayField != nil {
		return af.arrayField.hasValue()
	} else if af.valueField != nil {
		return af.valueField.hasValue()
	}
	return false
}

// MarshalJSON implements json marshaller.
func (af *AttributeField) MarshalJSON() ([]byte, error) {
	if af.valueField != nil {
		af.Field.Value = af.valueField
	} else if af.mapField != nil {
		af.Field.Value = af.mapField
	} else if af.arrayField != nil {
		af.Field.Value = af.arrayField
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

	vf, arrF, mf, err := convertRiverValue(val)
	if err != nil {
		return err
	}

	if vf.hasValue() {
		af.valueField = vf
	} else if arrF.hasValue() {
		af.arrayField = arrF
	} else if mf.hasValue() {
		af.mapField = mf
	} else {
		return fmt.Errorf("unable to find value for %T in convertAttribute", val.Interface())
	}
	return nil
}
