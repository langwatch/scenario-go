package ptr

func Ptr[T any](v T) *T {
	return &v
}

func ValueOrZero[T any](v *T) T {
	if v == nil {
		return *new(T)
	}
	return *v
}
