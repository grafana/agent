package config

import (
	"encoding/json"
	"reflect"

	"github.com/fatih/structs"
)

// jsonnetMarshal marshals a value for passing to Jsonnet.
//
// Structs are marshaled to the JSON representation of the Go value, ignoring
// all json struct tags. Fields from structs must be accesed as they would from
// Go, with the exception of embedded fields which can only be accessed through
// the embedded type name.
func jsonnetMarshal(val interface{}) ([]byte, error) {
	return json.Marshal(jsonnetValue(val))
}

func jsonnetValue(in interface{}) interface{} {
	inValue := reflect.ValueOf(in)

	switch inValue.Kind() {
	case reflect.Ptr:
		if inValue.IsNil() {
			return nil
		}
		return jsonnetValue(inValue.Elem().Interface())
	case reflect.Struct:
		return structs.Map(in)
	case reflect.Array, reflect.Slice:
		elem := make([]interface{}, inValue.Len())
		for i := 0; i < inValue.Len(); i++ {
			elem[i] = jsonnetValue(inValue.Index(i).Interface())
		}
		return elem
	default:
		return in
	}
}
