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
	Name        string  `json:"name"`
	IsBlock     bool    `json:"is_block"`
	IsAttribute bool    `json:"is_attribute"`
	IsArray     bool    `json:"is_array"`
	IsOptional  bool    `json:"is_optional"`
	DataType    string  `json:"datatype"`
	Fields      []Field `json:"fields"`
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
	value
	reflect_utils.WalkStructElements(t, func(field reflect.StructField) bool {

	})

	return c, nil
}
