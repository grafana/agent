package bijection

type FnBijection[A interface{}, B interface{}] struct {
	AtoB func(*A, *B) error
	BtoA func(*B, *A) error
}

func (f FnBijection[A, B]) ConvertAToB(a *A, b *B) error {
	return f.AtoB(a, b)
}

func (f FnBijection[A, B]) ConvertBToA(b *B, a *A) error {
	return f.BtoA(b, a)
}
