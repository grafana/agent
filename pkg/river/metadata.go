package river

import (
	"reflect"
	"strconv"
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
	Name          string `json:"string"`
	IsSingleton   bool   `json:"is_singleton"`
	ArgumentField string `json:"argument_field"`
	ExportField   string `json:"export_field"`
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

func (d DataType) Equals(dt DataType) bool {
	if len(d.Fields) != len(dt.Fields) {
		return false
	}
	for _, f := range d.Fields {
		found := false
		for _, foundField := range dt.Fields {
			if f.DataType != foundField.DataType {
				return false
			}
			if f.ArrayType != foundField.ArrayType {
				return false
			}
			if f.IsAttribute != foundField.IsAttribute {
				return false
			}
			if f.IsBlock != foundField.IsBlock {
				return false
			}
			if f.IsMap != foundField.IsMap {
				return false
			}
			if f.MapKeyType != foundField.MapKeyType {
				return false
			}
			if f.MapValueType != foundField.MapValueType {
				return false
			}
			found = true

		}
		if !found {
			return false
		}
	}
	return true
}

type MetadataDict struct {
	Types []DataType
}

func (md *MetadataDict) GenerateComponent(name string, isSingleton bool, arguments interface{}, exports interface{}) (Component, error) {
	c := Component{
		Name:        name,
		IsSingleton: isSingleton,
	}
	argname := ""
	var err error
	if arguments != nil {
		argname, err = md.generateField("arguments", reflect.TypeOf(arguments))
		if err != nil {
			return Component{}, err
		}
	}
	expname := ""
	if exports != nil {
		expname, err = md.generateField("exports", reflect.TypeOf(exports))
		if err != nil {
			return Component{}, err
		}
	}

	c.ArgumentField = argname
	c.ExportField = expname
	return c, nil
}

func (md *MetadataDict) find(name string) (bool, DataType) {
	for _, x := range md.Types {
		if x.Name == name {
			return true, x
		}
	}
	return false, DataType{}
}

func (md *MetadataDict) generateField(preferredName string, p reflect.Type) (string, error) {
	dt := DataType{}
	t := reflect_utils.NonPointer(p)
	fields := rivertags.Get(t)
	mFields := make([]TagField, 0)
	var err error
	for _, field := range fields {
		fName := strings.Join(field.Name, ".")
		metaField := TagField{
			Name:        fName,
			IsBlock:     field.IsBlock(),
			IsAttribute: field.IsAttr(),
			IsOptional:  field.IsOptional(),
		}
		reflectField := t.FieldByIndex(field.Index)
		datatype := getType(reflectField.Type)
		metaField.DataType = datatype
		if metaField.DataType == "array" {
			metaField, err = md.handleArray(fName, metaField, reflectField.Type)
			if err != nil {
				return "", err
			}

		} else if metaField.DataType == "map" {
			metaField, err = md.handleMap(metaField, reflectField.Type)
			if err != nil {
				return "", err
			}

		} else if metaField.DataType == "object" {
			metaField, err = md.handleObject(fName, metaField, reflectField.Type)
			if err != nil {
				return "", err
			}
		}
		mFields = append(mFields, metaField)
	}
	dt.Fields = mFields
	dt.Name = getName(preferredName, dt, 0, md.Types)

	isUnique := true
	for _, x := range md.Types {
		if x.Equals(dt) {
			isUnique = false
			break
		}
	}
	if isUnique {
		md.Types = append(md.Types, dt)
	}

	return dt.Name, nil
}

func (md *MetadataDict) handleArray(riverName string, tg TagField, t reflect.Type) (TagField, error) {
	tg.IsArray = true
	elem := value.RiverType(t.Elem())

	tg.ArrayType = elem.String()
	if tg.ArrayType == "object" {
		elem := t.Elem()
		k := elem.Kind()
		objType := getType(t)
		name := ""
		var err error
		if k == reflect.Map {
			tg, err = md.handleMap(tg, t)
			if err != nil {
				return tg, err
			}
		} else if objType == "object" {
			tg, err = md.handleObject(riverName, tg, t)
			if err != nil {
				return tg, err
			}
		}
		tg.DataType = name
	}
	return tg, nil
}

func (md *MetadataDict) handleObject(riverName string, tg TagField, t reflect.Type) (TagField, error) {
	name, err := md.generateField(riverName, t)
	if err != nil {
		return tg, err
	}
	tg.DataType = name
	return tg, nil
}

func (md *MetadataDict) handleMap(tg TagField, t reflect.Type) (TagField, error) {
	tg.IsMap = true
	elemVal := value.RiverType(t.Elem())
	tg.MapValueType = elemVal.String()

	keyVal := value.RiverType(t.Key())
	tg.MapKeyType = keyVal.String()
	return tg, nil
}

func getName(preferredName string, dt DataType, iteration int, dataTypes []DataType) string {
	for _, x := range dataTypes {
		if x.Equals(dt) {
			return x.Name
		}
	}
	for _, x := range dataTypes {
		if x.Name == preferredName {
			iteration = iteration + 1
			return getName(x.Name+strconv.Itoa(iteration), dt, iteration, dataTypes)
		}
	}
	return preferredName
}

func getType(t reflect.Type) string {
	riverType := value.RiverType(t)

	k := t.Kind()
	if k == reflect.Map {
		return "map"
	}
	return riverType.String()
}
