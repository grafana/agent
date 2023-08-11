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
	aToBMap map[PropertyNames]unsafeConversionFn
	bToAMap map[PropertyNames]unsafeConversionFn
}

type unsafeConversionFn func(interface{}, interface{}) error

type FnBijection[A interface{}, B interface{}] struct {
	AtoB func(*A, *B) error
	BtoA func(*B, *A) error
}

// BindField needs to be a function, because methods cannot have extra generic parameters
// other than the ones defined on the structure.
func BindField[A any, B any, FieldA any, FieldB any](si *StructBijection[A, B], names PropertyNames, fBijection Bijection[FieldA, FieldB]) {
	si.ensureInitialised()
	si.aToBMap[names] = func(a interface{}, b interface{}) error {
		typedA, ok := a.(*FieldA)
		if !ok {
			return fmt.Errorf("type of %+v is %[1]T, which does not match the A type of bijection %+v", a, fBijection)
		}
		typedB, ok := b.(*FieldB)
		if !ok {
			return fmt.Errorf("type of %+v is %[1]T, which does not match the B type of bijection %+v", b, fBijection)
		}
		return fBijection.ConvertAToB(typedA, typedB)
	}

	si.bToAMap[names] = func(b interface{}, a interface{}) error {
		typedA, ok := a.(*FieldA)
		if !ok {
			return fmt.Errorf("type of %+v is %[1]T, which does not match the A type of bijection %+v", a, fBijection)
		}
		typedB, ok := b.(*FieldB)
		if !ok {
			return fmt.Errorf("type of %+v is %[1]T, which does not match the B type of bijection %+v", b, fBijection)
		}
		return fBijection.ConvertBToA(typedB, typedA)
	}
}

func (si *StructBijection[A, B]) ensureInitialised() {
	if si.aToBMap == nil {
		si.aToBMap = make(map[PropertyNames]unsafeConversionFn)
	}
	if si.bToAMap == nil {
		si.bToAMap = make(map[PropertyNames]unsafeConversionFn)
	}
}

func Copy[A any]() Bijection[A, A] {
	return FnBijection[A, A]{
		AtoB: func(a *A, b *A) error {
			*b = *a
			return nil
		},
		BtoA: func(b *A, a *A) error {
			*a = *b
			return nil
		},
	}
}

func (f FnBijection[A, B]) ConvertAToB(a *A, b *B) error {
	return f.AtoB(a, b)
}

func (f FnBijection[A, B]) ConvertBToA(b *B, a *A) error {
	return f.BtoA(b, a)
}

type PropertyNames struct {
	A string
	B string
}

func NewBijection[A interface{}, B interface{}](a A, b B) StructBijection[A, B] {
	return StructBijection[A, B]{}
}

func (si *StructBijection[A, B]) ConvertAToB(from *A, to *B) error {
	for names, fn := range si.aToBMap {
		fromProp := reflect.ValueOf(from).Elem().FieldByName(names.A)
		toProp := reflect.ValueOf(to).Elem().FieldByName(names.B)
		err := fn(fromProp.Addr().Interface(), toProp.Addr().Interface())
		if err != nil {
			return err
		}
	}
	return nil
}

func (si *StructBijection[A, B]) ConvertBToA(from *B, to *A) error {
	for names, fn := range si.bToAMap {
		fromProp := reflect.ValueOf(from).Elem().FieldByName(names.B)
		toProp := reflect.ValueOf(to).Elem().FieldByName(names.A)
		err := fn(fromProp.Addr().Interface(), toProp.Addr().Interface())
		if err != nil {
			return err
		}
	}
	return nil
}

func identity(a interface{}) (b interface{}, err error) {
	return a, nil
}
