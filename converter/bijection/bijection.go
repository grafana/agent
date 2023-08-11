package bijection

import (
	"fmt"
	"reflect"
)

type Bijection[A interface{}, B interface{}] interface {
	ConvertAToB(a *A, b *B) error
	ConvertBToA(b *B, a *A) error
}

type StructBijection[A interface{}, B interface{}] struct {
	Mappings map[PropertyPair]interface{}
}

type FnBijection[A interface{}, B interface{}] struct {
	AtoB func(*A, *B) error
	BtoA func(*B, *A) error
}

var Identiy = FnBijection[any, any]{
	AtoB: func(a *any, b *any) error {
		*b = *a
		return nil
	},
	BtoA: func(b *any, a *any) error {
		*a = *b
		return nil
	},
}

func (f FnBijection[A, B]) ConvertAToB(a *A, b *B) error {
	return f.AtoB(a, b)
}

func (f FnBijection[A, B]) ConvertBToA(b *B, a *A) error {
	return f.BtoA(b, a)
}

type PropertyPair struct {
	A string
	B string
}

func NewBijection[A interface{}, B interface{}](a A, b B) StructBijection[A, B] {
	return StructBijection[A, B]{}
}

func (bj *StructBijection[A, B]) Bind(
	fromProp string,
	toProp string,
	fwd func(interface{}) (interface{}, error),
	rev func(interface{}) (interface{}, error),
) {
	var (
		a A
		b B
	)

	typeA := reflect.TypeOf(a)
	if _, ok := typeA.FieldByName(fromProp); !ok {
		panic(fmt.Sprintf("field %s not found in type %s", fromProp, typeA.Name()))
	}

	typeB := reflect.TypeOf(b)
	if _, ok := typeB.FieldByName(toProp); !ok {
		panic(fmt.Sprintf("field %s not found in type %s", toProp, typeB.Name()))
	}

	bj.Mappings[PropertyPair{fromProp, toProp}] = FnBijection[any, any]{
		AtoB: fwd,
	}
}

func (bj *StructBijection[A, B]) ConvertAToB(from *A, to *B) error {
	for props, fn := range bj.Mappings {
		fromProp := reflect.ValueOf(from).Elem().FieldByName(conv.A)
		toProp := reflect.ValueOf(to).Elem().FieldByName(conv.BProp)
		converted, err := conv.AB(fromProp.Interface())
		if err != nil {
			return err
		}
		toProp.Set(reflect.ValueOf(converted))
	}
	return nil
}

func (bj *StructBijection[A, B]) ConvertBToA(from *B, to *A) error {
	for _, conv := range bj.props {
		fromProp := reflect.ValueOf(from).Elem().FieldByName(conv.BProp)
		toProp := reflect.ValueOf(to).Elem().FieldByName(conv.AProp)
		converted, err := conv.BA(fromProp.Interface())
		if err != nil {
			return err
		}
		toProp.Set(reflect.ValueOf(converted))
	}
	return nil
}

func identity(a interface{}) (b interface{}, err error) {
	return a, nil
}
