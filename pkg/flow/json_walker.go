package flow

import (
	"fmt"
	"reflect"
	"strings"
	"time"

	"github.com/grafana/agent/component/common/relabel"

	"github.com/grafana/regexp"

	"github.com/grafana/agent/pkg/flow/rivertypes"

	"github.com/grafana/agent/pkg/flow/internal/controller"
	"github.com/grafana/agent/pkg/river/rivertags"
)

type Walker struct {
}

func NewWalker() *Walker {
	return &Walker{}
}

func (w *Walker) ConvertBlock(c *controller.ComponentNode) *Field {
	nf := &Field{}
	nf.Type = "block"
	nf.Key = strings.Join(c.ID(), ".")
	fields := make([]*Field, 0)
	args := w.ConvertArguments(c)
	if args != nil {
		fields = append(fields, args)
	}
	exports := w.ConvertExports(c)
	if exports != nil {
		fields = append(fields)
	}
	nf.Value = fields
	return nf

}

func (w *Walker) ConvertArguments(c *controller.ComponentNode) *Field {
	return convertField(c.Arguments(), &rivertags.Field{
		Name:  []string{"arguments"},
		Index: nil,
		Flags: 0,
	})
}

func (w *Walker) ConvertExports(c *controller.ComponentNode) *Field {
	return convertField(c.Exports(), &rivertags.Field{
		Name:  []string{"exports"},
		Index: nil,
		Flags: 0,
	})
}

// ConvertToField converts to a generic field for json
func (w *Walker) ConvertToField(in interface{}, name string) *Field {
	return convertField(in, &rivertags.Field{
		Name:  []string{name},
		Index: nil,
		Flags: 0,
	})
}

func convertField(in interface{}, f *rivertags.Field) *Field {
	nf := &Field{}
	if f != nil && len(f.Name) > 0 {
		nf.Key = f.Name[len(f.Name)-1]
	}
	if nf.Key == "follow_redirects" {
		println("")
	}

	nt := reflect.TypeOf(in)
	vIn := reflect.ValueOf(in)
	if in != nil {
		for nt.Kind() == reflect.Pointer && !vIn.IsZero() {
			vIn = vIn.Elem()
			nt = vIn.Type()
		}
		in = vIn.Interface()
	}

	if in == nil || (reflect.TypeOf(in).Kind() == reflect.Pointer && reflect.ValueOf(in).IsNil()) || reflect.ValueOf(in).IsZero() {
		/*nf.Value = &Field{
			Type: "null",
		}
		return nf*/
		return nil
	}
	switch in.(type) {
	case int, int8, int16, int32, int64, uint8, uint16, uint32, uint64, float32, float64:
		nf.Value = convertNumber(in)
		return nf
	case string:
		nf.Value = &Field{
			Type:  "string",
			Value: in,
		}
		return nf
	case bool:
		nf.Value = &Field{
			Type:  "string",
			Value: in,
		}
		return nf
	case rivertypes.Secret:
		nf.Value = &Field{
			Type:  "string",
			Value: "(secret)",
		}
		return nf
	case rivertypes.OptionalSecret:
		nf.Value = &Field{
			Type:  "string",
			Value: "(secret)",
		}
		maybeSecret := in.(rivertypes.OptionalSecret)
		if !maybeSecret.IsSecret {
			nf.Value.(*Field).Value = maybeSecret.Value
		}
		return nf
	case time.Duration:
		nf.Value = &Field{
			Type:  "string",
			Value: in.(time.Duration).String(),
		}
		return nf
	case regexp.Regexp:
		nf.Value = &Field{
			Type:  "string",
			Value: in.(*regexp.Regexp).String(),
		}
		return nf
	case relabel.Regexp:
		nf.Value = &Field{
			Type:  "string",
			Value: in.(relabel.Regexp).String(),
		}
		return nf
	}

	switch nt.Kind() {
	// This handles aliases for other types.
	case reflect.String:
		conv := reflect.ValueOf(in)
		nf.Value = &Field{
			Type:  "string",
			Value: conv.String(),
		}
		return nf
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		conv := reflect.ValueOf(in)
		nf.Type = "number"
		nf.Value = conv.Int()
		return nf
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		conv := reflect.ValueOf(in)
		nf.Type = "number"
		nf.Value = conv.Uint()
		return nf
	case reflect.Float32, reflect.Float64:
		conv := reflect.ValueOf(in)
		nf.Type = "number"
		nf.Value = conv.Float()
		return nf
	case reflect.Struct:
		if f != nil && f.IsBlock() {
			nf.Type = "block"
		} else {
			nf.Type = "object"
		}

		fields := make([]*Field, 0)
		riverFields := rivertags.Get(reflect.TypeOf(in))
		for _, rf := range riverFields {
			fieldValue := vIn.FieldByIndex(rf.Index)
			found := convertField(fieldValue.Interface(), &rf)
			if found != nil {
				fields = append(fields, found)
			}
		}
		nf.Value = fields
		return nf
	case reflect.Array, reflect.Slice:
		nf.Type = "array"
		fields := make([]*Field, 0)
		for i := 0; i < vIn.Len(); i++ {
			arrEle := vIn.Index(i).Interface()
			found := convertField(arrEle, nil)
			if found != nil {
				fields = append(fields, found)
			}
		}
		nf.Value = fields
		return nf
	case reflect.Map:
		nf.Type = "map"
		fields := make([]*Field, 0)
		iter := vIn.MapRange()
		for iter.Next() {
			mf := &Field{}
			mf.Key = iter.Key().String()
			mf.Value = convertField(iter.Value().Interface(), nil)
			if mf.Value != nil {
				fields = append(fields, mf)
			}
		}
		nf.Value = fields
		return nf
	default:
		panic(fmt.Sprintf("unknown type %T for kind %d", in, vIn.Kind()))
	}

	panic("could not convert object for json_walking")
}

func convertNumber(in interface{}) *Field {

	f := &Field{}
	f.Type = "number"
	switch in.(type) {
	case int:
		f.Value = in
	case int8:
		f.Value = in
	case int16:
		f.Value = in
	case int32:
		f.Value = in
	case int64:
		f.Value = in
	case uint:
		f.Value = in
	case uint8:
		f.Value = in
	case uint16:
		f.Value = in
	case uint32:
		f.Value = in
	case uint64:
		f.Value = in
	case float32:
		f.Value = in
	case float64:
		f.Value = in
	default:
		panic("invalid input")
	}
	return f
}

type Field struct {
	Key   string      `json:"key,omitempty"`
	Type  string      `json:"type,omitempty"`
	Value interface{} `json:"value,omitempty"`
}
