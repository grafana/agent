package bijection2

type Bijection[A interface{}, B interface{}] interface {
	ConvertAToB(a *A, b *B) error
	ConvertBToA(b *B, a *A) error
}
