package bijection

import (
	"fmt"
	"reflect"
)

type Bijection[A any, B any] struct {
	TypeA reflect.Type
	TypeB reflect.Type
	props []PropertyBijection
}

type PropertyBijection struct {
	AProp string
	BProp string
	AB    func(any) (any, error)
	BA    func(any) (any, error)
}

func (b *Bijection[A, B]) Bind(
	fromProp string,
	toProp string,
	fwd func(any) (any, error),
	rev func(any) (any, error),
) {
	if _, ok := b.TypeA.FieldByName(fromProp); !ok {
		panic(fmt.Sprintf("field %s not found in type %s", fromProp, b.TypeA.Name()))
	}
	if _, ok := b.TypeB.FieldByName(toProp); !ok {
		panic(fmt.Sprintf("field %s not found in type %s", toProp, b.TypeB.Name()))
	}
	b.props = append(b.props, PropertyBijection{
		AProp: fromProp,
		BProp: toProp,
		AB:    fwd,
		BA:    rev,
	})
}

func (b *Bijection[A, B]) ConvertAToB(from *A, to *B) error {
	for _, conv := range b.props {
		fromProp := reflect.ValueOf(from).Elem().FieldByName(conv.AProp)
		toProp := reflect.ValueOf(to).Elem().FieldByName(conv.BProp)
		converted, err := conv.AB(fromProp.Interface())
		if err != nil {
			return err
		}
		toProp.Set(reflect.ValueOf(converted))
	}
	return nil
}

func (b *Bijection[A, B]) ConvertBToA(from *B, to *A) error {
	for _, conv := range b.props {
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

func NewBijection[A any, B any](a A, b B) Bijection[A, B] {
	return Bijection[A, B]{
		TypeA: reflect.TypeOf(a),
		TypeB: reflect.TypeOf(b),
	}
}

func identity(a interface{}) (b interface{}, err error) {
	return a, nil
}
