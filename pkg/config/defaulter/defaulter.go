package defaulter

import (
	"reflect"
	"strings"

	"gopkg.in/yaml.v3"
)

func GetFieldsNotDefined(node *yaml.Node, v interface{}) []string {
	undefined := make([]string, 0)
	val := reflect.ValueOf(v)
	indirect := reflect.Indirect(val)
	rv := indirect.Type()
	for i := 0; i < rv.NumField(); i++ {
		f := rv.Field(i)
		yamlTag, found := f.Tag.Lookup("yaml")
		if !found {
			continue
		}
		fieldName := strings.Split(yamlTag, ",")[0]
		if doesNodeExist(fieldName, node.Content) {
			continue
		}
		undefined = append(undefined, yamlTag)
	}
	return undefined
}

func ApplyDefaultsDefined(node *yaml.Node, v interface{}) {
	val := reflect.ValueOf(v)
	indirect := reflect.Indirect(val)
	rv := indirect.Type()
	for i := 0; i < rv.NumField(); i++ {
		f := rv.Field(i)
		yamlTag, found := f.Tag.Lookup("yaml")
		if !found {
			continue
		}

		fieldName := strings.Split(yamlTag, ",")[0]
		if doesNodeExist(fieldName, node.Content) {
			continue
		}

		defaultVal, defFound := f.Tag.Lookup("default")
		if !defFound {
			continue
		}
		switch indirect.FieldByName(f.Name).Type() {
		case reflect.TypeOf(""):
			indirect.FieldByName(f.Name).Set(reflect.ValueOf(defaultVal))
		}
	}
}

func doesNodeExist(name string, nodes []*yaml.Node) bool {
	if len(nodes) == 0 {
		return false
	}
	if len(nodes) == 1 {
		return false
	}
	for i := 0; i < len(nodes); i = i + 2 {
		if nodes[i].Kind == yaml.ScalarNode && nodes[i].Value == name {
			return true
		}
	}
	return false
}
