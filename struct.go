package pp

import (
	"context"
	"errors"
)

var NilStructErr = errors.New("struct pointer is nil")

type StructWrapper[T any] func(context.Context, *T) error

// Struct receives a pointer to a struct, which is used as reference for a modifying function `f`.
// It then returns a pipeline function calling `f`.
// This function must be used with caution, because running it in parallel might cause data races.
// Avoid calling a pipeline that modifies the same field twice in parallel.
// If you only want to modify 1 field, use `Field`
func Struct[T any](s *T, wrapped StructWrapper[T]) StepFunc {
	return func(ctx context.Context) error {
		if s == nil {
			return NilStructErr
		}
		return wrapped(ctx, s)
	}
}
