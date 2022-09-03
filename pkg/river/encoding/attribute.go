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
	Field       `json:",omitempty"`
	valueField  *ValueField
	arrayField  *ArrayField
	mapField    *MapField
	structField *StructField
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
		return af.mapField.isValid()
	} else if af.arrayField != nil {
		return af.arrayField.isValid()
	} else if af.structField != nil {
		return af.structField.isValid()
	} else if af.valueField != nil {
		return af.valueField.isValid()
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
	} else if af.structField != nil {
		af.Field.Value = af.structField
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

	vf, arrF, mf, sf, err := convertRiverValue(val)
	if err != nil {
		return err
	}

	if vf.isValid() {
		af.valueField = vf
	} else if arrF.isValid() {
		af.arrayField = arrF
	} else if mf.isValid() {
		af.mapField = mf
	} else if sf.isValid() {
		af.structField = sf
	} else {
		return fmt.Errorf("unable to find value for %T in convertAttribute", val.Interface())
	}
	return nil
}
