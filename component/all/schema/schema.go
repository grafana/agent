package schema

import (
	"fmt"
	"github.com/grafana/agent/component/all/schema/rivertags"
	"gopkg.in/yaml.v2"
	"reflect"
	"strings"
)

type Component struct {
	Name      string  `yaml:"name"`
	Arguments []Field `yaml:"arguments,omitempty"`
	Exports   []Field `yaml:"exports,omitempty"`
}

type Field struct {
	Type       string   `yaml:"type"`
	Name       string   `yaml:"name"`
	RiverFlags []string `yaml:"flags,omitempty"`
	Nested     []Field  `yaml:"nested,omitempty"`
}

func RiverToYAML(v interface{}) (string, error) {
	fields, err := getFields(v, "")
	if err != nil {
		return "", err
	}

	yamlData, err := yaml.Marshal(fields)
	if err != nil {
		return "", err
	}

	return string(yamlData), nil
}

func getFields(v interface{}, parentName string) ([]Field, error) {
	typ := reducePointer(reflect.TypeOf(v))
	if typ.Kind() != reflect.Struct {
		return nil, fmt.Errorf("expected struct, got %q when processing %q", typ, parentName)
	}

	var fields []Field
	tags := rivertags.Get(typ)
	for _, tag := range tags {
		field := typ.FieldByIndex(tag.Index)
		riverFieldShortName := strings.Join(tag.Name, ".")
		riverFieldFullName := parentName + "." + riverFieldShortName
		docsFieldType := getRiverDocsFieldType(field.Type)
		f := Field{
			Type:       docsFieldType,
			Name:       riverFieldShortName,
			RiverFlags: riverFlags(tag),
		}
		if hasRiverTags(field.Type) {
			arguments, err := getFields(reflect.New(field.Type).Interface(), riverFieldFullName)
			if err != nil {
				return nil, err
			}
			f.Nested = arguments
		}
		fields = append(fields, f)
	}

	return fields, nil
}

func reducePointer(typ reflect.Type) reflect.Type {
	for typ.Kind() == reflect.Ptr {
		typ = typ.Elem()
	}
	return typ
}

func hasRiverTags(t reflect.Type) bool {
	t = reducePointer(t)
	if t.Kind() != reflect.Struct {
		return false
	}

	hasRiverTag := false
	for i := 0; i < t.NumField(); i++ {
		f := t.Field(i)
		if _, ok := f.Tag.Lookup("river"); ok {
			hasRiverTag = true
			break
		}
	}

	if !hasRiverTag {
		return false
	}

	tags := rivertags.Get(t)
	return len(tags) > 0
}

func getRiverDocsFieldType(fieldType reflect.Type) string {
	fieldType = reducePointer(fieldType)
	if fieldType.Kind() == reflect.Slice {
		insideType := getRiverDocsFieldType(fieldType.Elem())
		return "list(" + insideType + ")"
	} else if fieldType.Kind() == reflect.Map {
		insideType := getRiverDocsFieldType(fieldType.Elem())
		return "map(" + insideType + ")"
	} else if fieldType.String() == "time.Duration" {
		return "duration"
	} else {
		return fieldType.Name()
	}
}

// TODO(thampiotr): put this in upstream
func riverFlags(fieldTags rivertags.Field) []string {
	f := fieldTags.Flags
	attrs := make([]string, 0, 5)

	if f&rivertags.FlagAttr != 0 {
		attrs = append(attrs, "attr")
	}
	if f&rivertags.FlagBlock != 0 {
		attrs = append(attrs, "block")
	}
	if f&rivertags.FlagEnum != 0 {
		attrs = append(attrs, "enum")
	}
	if f&rivertags.FlagOptional != 0 {
		attrs = append(attrs, "optional")
	}
	if f&rivertags.FlagLabel != 0 {
		attrs = append(attrs, "label")
	}
	if f&rivertags.FlagSquash != 0 {
		attrs = append(attrs, "squash")
	}

	return attrs
}
