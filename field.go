package pipego

import (
	"context"
	"errors"
)

// NilFieldError is returned when a nil pointer is passed to a wrapper function.
var NilFieldError = errors.New("field pointer is nil")

// Wrapper definitions for common data structures in apis.
type (
	FetchField[T any]             func(context.Context) (*T, error)
	FetchSlice[T any]             func(context.Context) ([]T, error)
	FetchMap[K comparable, V any] func(context.Context) (map[K]V, error)
)

// Field receives a pointer `*field` and returns a pipeline function that assigns
// any non-nil value to the pointer, if no error is returned.
// Field is supposed to be used for changing only 1 field at a time.
// For changing more than one field at once, use `Struct`.
func Field[T any](field *T, fetch FetchField[T]) StepFunc {
	return func(ctx context.Context) error {
		if field == nil {
			return NilFieldError
		}
		ptr, err := fetch(ctx)
		if ptr != nil {
			*field = *ptr
		}
		return err
	}
}

// Slice receives a pointer to a slice and a wrapper, and returns a StepFunc.
// This function should be used for slices, since it's a very common signature
// of APIs to return ([]T, error).
func Slice[T any](field *[]T, fetch FetchSlice[T]) StepFunc {
	return func(ctx context.Context) (err error) {
		if field == nil {
			return NilFieldError
		}
		*field, err = fetch(ctx)
		return
	}
}

// Map receives a pointer to a map and a wrapper, and returns a StepFunc.
// This function should be used for maps, since it's a very common signature
// of APIs to return (map[K]V, error).
func Map[K comparable, V any](field *map[K]V, fetch FetchMap[K, V]) StepFunc {
	return func(ctx context.Context) (err error) {
		if field == nil {
			return NilFieldError
		}
		*field, err = fetch(ctx)
		return
	}
}
