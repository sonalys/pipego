package pp_test

func pointer[T any](v T) *T {
	return &v
}
