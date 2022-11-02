package river

import (
	"reflect"
	"strings"

	"github.com/grafana/agent/pkg/river/internal/rivertags"
	"github.com/grafana/agent/pkg/river/internal/value"
	reflect_utils "github.com/muir/reflectutils"
)

type Metadata struct {
	ProductName string `json:"product_name"`
	Version     string `json:"version"`
	URL         string `json:"url"`
}

type Component struct {
	Name          string   `json:"string"`
	IsSingleton   bool     `json:"is_singleton"`
	ArgumentField TagField `json:"argument_field"`
	ExportField   TagField `json:"export_field"`
}

type TagField struct {
	Name         string `json:"name"`
	IsBlock      bool   `json:"is_block"`
	IsAttribute  bool   `json:"is_attribute"`
	IsArray      bool   `json:"is_array"`
	IsMap        bool   `json:"is_map"`
	IsOptional   bool   `json:"is_optional"`
	DataType     string `json:"datatype"`
	ArrayType    string `json:"array_type,omitempty"`
	MapKeyType   string `json:"map_key_type,omitempty"`
	MapValueType string `json:"map_value_type,omitempty"`
}

type MapType struct {
	Key   string `json:"key"`
	Value string `json:"value"`
}

type DataType struct {
	Name   string     `json:"name"`
	Fields []TagField `json:"fields"`
}

func GenerateComponent(name string, isSingleton bool, arguments interface{}, exports interface{}) (Component, error) {
	/*c := Component{
		Name:        name,
		IsSingleton: isSingleton,
	}
	_, err := generateField("arguments", arguments)
	if err != nil {
		return Component{}, err
	}
	_, err = generateField("exports", exports)
	if err != nil {
		return Component{}, err
	}
	//	c.ArgumentField = arg
	//	c.ExportField = exp*/
	return Component{}, nil
}

func generateField(tag *TagField, v interface{}) error {

	t := reflect.TypeOf(v)
	t = reflect_utils.NonPointer(t)
	val := value.Encode(v)
	refl := val.Reflect()
	fields := rivertags.Get(t)
	for _, field := range fields {
		metaField := &TagField{
			Name:        strings.Join(field.Name, "."),
			IsBlock:     field.IsBlock(),
			IsAttribute: field.IsAttr(),
			IsOptional:  field.IsOptional(),
			Fields:      make([]*TagField, 0),
		}
		reflectField := refl.FieldByIndex(field.Index)
		fieldVal := value.Encode(reflectField.Interface())
		datatype := getType(fieldVal)
		metaField.DataType = datatype
		if metaField.DataType == "array" {
			metaField.IsArray = true
			elemVal := value.RiverType(fieldVal.Reflect().Type().Elem())
			metaField.ArrayType = elemVal.String()
		} else if metaField.DataType == "map" {
			metaField.IsMap = true
			elemVal := value.RiverType(fieldVal.Reflect().Type().Elem())
			metaField.MapValueType = elemVal.String()

			keyVal := value.RiverType(fieldVal.Reflect().Type().Key())
			metaField.MapKeyType = keyVal.String()
		} else if metaField.DataType == "object" {
			err := generateField(metaField, reflectField.Interface())
			if err != nil {
				return err
			}
		}
		tag.Fields = append(tag.Fields, metaField)
	}
	return nil
}

func getType(val value.Value) string {
	if val.Reflect().Kind() == reflect.Map {
		return "map"
	}
	return val.Describe()
}
