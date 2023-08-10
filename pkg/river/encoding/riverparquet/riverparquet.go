package riverparquet

import (
	"fmt"
	"reflect"
	"strings"

	"github.com/grafana/agent/pkg/river/encoding/riverjson"
	"github.com/grafana/agent/pkg/river/internal/reflectutil"
	"github.com/grafana/agent/pkg/river/internal/rivertags"
	"github.com/grafana/agent/pkg/river/internal/value"
)

var goRiverDefaulter = reflect.TypeOf((*value.Defaulter)(nil)).Elem()

func GetComponentDetail(val interface{}) []Row {
	rv := reflect.ValueOf(val)

	var idCounter uint = 1
	return getComponentDetailInt(rv, 0, &idCounter)
}

func getComponentDetailInt(rv reflect.Value, parentId uint, idCounter *uint) []Row {
	for rv.Kind() == reflect.Pointer {
		if rv.IsNil() {
			return []Row{}
		}
		rv = rv.Elem()
	}

	if rv.Kind() == reflect.Invalid {
		return []Row{}
	} else if rv.Kind() != reflect.Struct {
		panic(fmt.Sprintf("river/encoding/riverjson: can only encode struct values to bodies, got %s", rv.Kind()))
	}

	fields := rivertags.Get(rv.Type())
	defaults := reflect.New(rv.Type()).Elem()
	if defaults.CanAddr() && defaults.Addr().Type().Implements(goRiverDefaulter) {
		defaults.Addr().Interface().(value.Defaulter).SetToDefault()
	}

	componentDetails := make([]Row, 0, len(fields))

	for _, field := range fields {
		fieldVal := reflectutil.Get(rv, field)
		fieldValDefault := reflectutil.Get(defaults, field)

		var isEqual = fieldVal.Comparable() && fieldVal.Equal(fieldValDefault)
		var isZero = fieldValDefault.IsZero() && fieldVal.IsZero()

		if field.IsOptional() && (isEqual || isZero) {
			continue
		}

		componentDetail := encodeFieldAsComponentDetail(nil, parentId, idCounter, field, fieldVal)
		componentDetails = append(componentDetails, componentDetail...)
	}

	return componentDetails
}

func encodeFieldAsComponentDetail(prefix []string, parentId uint, idCounter *uint, field rivertags.Field, fieldValue reflect.Value) []Row {
	fieldName := strings.Join(field.Name, ".")

	for fieldValue.Kind() == reflect.Pointer {
		if fieldValue.IsNil() {
			break
		}
		fieldValue = fieldValue.Elem()
	}

	switch {
	case field.IsAttr():
		jsonVal, err := riverjson.MarshalValue(fieldValue.Interface())
		if err != nil {
			panic("river/encoding/riverjson: failed to create a JSON value: " + err.Error())
		}

		curId := *idCounter
		*idCounter += 1
		return []Row{
			{
				ID:         curId,
				ParentID:   parentId,
				Name:       fieldName,
				Label:      "",
				RiverType:  "attr",
				RiverValue: jsonVal,
			},
		}

	case field.IsBlock():
		fullName := mergeSlices(prefix, field.Name)

		switch {
		case fieldValue.Kind() == reflect.Map:
			// Iterate over the map and add each element as an attribute into it.

			if fieldValue.Type().Key().Kind() != reflect.String {
				panic("river/encoding/riverjson: unsupported map type for block; expected map[string]T, got " + fieldValue.Type().String())
			}

			blockID := *idCounter
			*idCounter += 1

			componentDetails := []Row{{
				ID:        blockID,
				ParentID:  parentId,
				Name:      strings.Join(fullName, "."),
				Label:     "",
				RiverType: "block",
			}}

			iter := fieldValue.MapRange()
			for iter.Next() {
				mapKey, mapValue := iter.Key(), iter.Value()

				jsonVal, err := riverjson.MarshalValue(mapValue.Interface())
				if err != nil {
					panic("river/encoding/riverjson: failed to create a JSON value: " + err.Error())
				}

				curId := *idCounter
				*idCounter += 1
				cd := Row{
					ID:         curId,
					ParentID:   blockID,
					Name:       mapKey.String(),
					Label:      "",
					RiverType:  "attr",
					RiverValue: jsonVal,
				}

				componentDetails = append(componentDetails, cd)
			}

			return componentDetails

		case fieldValue.Kind() == reflect.Slice, fieldValue.Kind() == reflect.Array:
			componentDetails := []Row{}

			for i := 0; i < fieldValue.Len(); i++ {
				elem := fieldValue.Index(i)

				// Recursively call encodeField for each element in the slice/array.
				// The recursive call will hit the case below and add a new block for
				// each field encountered.
				componentDetails = append(componentDetails, encodeFieldAsComponentDetail(prefix, parentId, idCounter, field, elem)...)
			}

			return componentDetails

		case fieldValue.Kind() == reflect.Struct:
			if fieldValue.IsZero() {
				curId := *idCounter
				*idCounter += 1

				// It shouldn't be possible to have a required block which is unset,
				// but we'll encode something anyway.
				return []Row{{
					ID:        curId,
					ParentID:  parentId,
					Name:      strings.Join(fullName, "."),
					Label:     "",
					RiverType: "block",
				}}
			}

			blockID := *idCounter
			*idCounter += 1

			componentDetails := []Row{{
				ID:        blockID,
				ParentID:  parentId,
				Name:      strings.Join(fullName, "."),
				Label:     getBlockLabel(fieldValue),
				RiverType: "block",
			}}

			componentDetails = append(componentDetails, getComponentDetailInt(fieldValue, blockID, idCounter)...)
			return componentDetails

		case fieldValue.Kind() == reflect.Interface:
			// Special case: try to get the underlying value as a block instead.
			if fieldValue.IsNil() {
				return []Row{}
			}
			return encodeFieldAsComponentDetail(prefix, parentId, idCounter, field, fieldValue.Elem())

		default:
			panic(fmt.Sprintf("river/encoding/riverjson: unrecognized block kind %s", fieldValue.Kind()))
		}

	case field.IsEnum():
		// Blocks within an enum have a prefix set.
		newPrefix := mergeSlices(prefix, field.Name)

		switch {
		case fieldValue.Kind() == reflect.Slice, fieldValue.Kind() == reflect.Array:
			details := []Row{}
			for i := 0; i < fieldValue.Len(); i++ {
				details = append(details, encodeEnumElementToDetails(newPrefix, fieldValue.Index(i), parentId, idCounter)...)
			}
			return details

		default:
			panic(fmt.Sprintf("river/encoding/riverjson: unrecognized enum kind %s", fieldValue.Kind()))
		}
	}

	return nil
}

func encodeEnumElementToDetails(prefix []string, enumElement reflect.Value, parentId uint, idCounter *uint) []Row {
	for enumElement.Kind() == reflect.Pointer {
		if enumElement.IsNil() {
			return nil
		}
		enumElement = enumElement.Elem()
	}

	fields := rivertags.Get(enumElement.Type())

	details := []Row{}

	// Find the first non-zero field and encode it.
	for _, field := range fields {
		fieldVal := reflectutil.Get(enumElement, field)
		if !fieldVal.IsValid() || fieldVal.IsZero() {
			continue
		}

		details = append(details, encodeFieldAsComponentDetail(prefix, parentId, idCounter, field, fieldVal)...)
		break
	}

	return details
}

func mergeSlices[V any](a, b []V) []V {
	if len(a) == 0 {
		return b
	} else if len(b) == 0 {
		return a
	}

	res := make([]V, 0, len(a)+len(b))
	res = append(res, a...)
	res = append(res, b...)
	return res
}

// getBlockLabel returns the label for a given block.
func getBlockLabel(rv reflect.Value) string {
	tags := rivertags.Get(rv.Type())
	for _, tag := range tags {
		if tag.Flags&rivertags.FlagLabel != 0 {
			return reflectutil.Get(rv, tag).String()
		}
	}

	return ""
}
