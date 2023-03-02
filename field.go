package pipego

import (
	"context"
	"errors"
)

var NilFieldError = errors.New("field pointer is nil")

// FetchFunc determines a pattern for the `Struct` function to automatically wrap common
// data fetching functions into the pipego format.
type FetchFunc[T any] func(context.Context) (*T, error)

// Field receives a pointer `*field` and returns a pipeline function that assigns
// any non-nil value to the pointer, if no error is returned.
// Field is supposed to be used for changing only 1 field at a time.
// For changing more than one field at once, use `Struct`.
func Field[T any](field *T, fetch FetchFunc[T]) StepFunc {
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
