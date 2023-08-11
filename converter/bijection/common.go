package bijection

import "reflect"

func Inverted[A any, B any](bj Bijection[A, B]) Bijection[B, A] {
	return FnBijection[B, A]{
		AtoB: bj.ConvertBToA,
		BtoA: bj.ConvertAToB,
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

func Cast[A any, B any]() Bijection[A, B] {
	return FnBijection[A, B]{
		AtoB: func(a *A, b *B) error {
			va := reflect.ValueOf(*a)
			*b = va.Convert(reflect.TypeOf(*b)).Interface().(B)
			return nil
		},
		BtoA: func(b *B, a *A) error {
			vb := reflect.ValueOf(*b)
			*a = vb.Convert(reflect.TypeOf(*a)).Interface().(A)
			return nil
		},
	}
}
