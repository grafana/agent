package river

import (
	"github.com/grafana/agent/pkg/river/internal/rivertags"
	"github.com/grafana/agent/pkg/river/internal/value"
	reflect_utils "github.com/muir/reflectutils"
	"reflect"
	"strings"
)

package parser

import (
"reflect"

"github.com/grafana/agent/pkg/river/rivertags"
reflect_utils "github.com/muir/reflectutils"
)

type Metadata struct {
	ProductName string `json:"product_name"`
	Version     string `json:"version"`
	URL         string `json:"url"`
}

type Component struct {
	Name        string  `json:"string"`
	IsSingleton bool    `json:"is_singleton"`
	Fields      []Field `json:"fields"`
}

type Field struct {
	Name         string  `json:"name"`
	IsBlock      bool    `json:"is_block"`
	IsAttribute  bool    `json:"is_attribute"`
	IsArray      bool    `json:"is_array"`
	IsMap bool `json:"is_map"`
	IsOptional   bool    `json:"is_optional"`
	DataType     string  `json:"datatype"`
	ArrayType    string  `json:"array_type,omitempty"`
	MapKeyType   string  `json:"key_type,omitempty"`
	MapValueType string  `json:"value_type,omitempty"`
	Fields       []Field `json:"fields,omitempty"`
}

func GenerateComponent(name string, isSingleton bool, v interface{}) (Component, error) {
	c := Component{
		Name:        name,
		IsSingleton: isSingleton,
		Fields:      make([]Field, 0),
	}
	t := reflect.TypeOf(v)
	t = reflect_utils.NonPointer(t)
	fields := rivertags.Get(t)
	val := value.Encode(v)
	for _, field := range fields {
		metaField := Field{
			Name:        strings.Join(field.Name,"."),
			IsBlock:     field.IsBlock(),
			IsAttribute: field.IsAttr(),
			IsOptional:  field.IsOptional(),
		}
		fieldVal := val.Index(field.Index[0])
		metaField.DataType = fieldVal.Describe()
		if metaField.DataType == "array" {
			metaField.IsArray = true
			elemVal := value.RiverType(fieldVal.Reflect().Type().Elem())
			metaField.ArrayType = elemVal.String()
		}
		if metaField.DataType == "map" {
			metaField.IsMap = true
			elemVal := value.RiverType(fieldVal.Reflect().Type().Elem())
			metaField.MapValueType = elemVal.String()

			keyVal := value.RiverType(fieldVal.Reflect().Type().Key())
			metaField.MapKeyType = keyVal.String()
		}
		c.Fields = append(c.Fields,metaField)
	}

	return c, nil
}

