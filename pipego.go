package pipego

import (
	"context"
	"errors"
	"sync"
)

type StepFunc func(ctx context.Context) (err error)

type FetchFunc[T any] func(context.Context) (*T, error)

// Parallel runs all the given steps in parallel,
// It cancels context for the first non-nil error and returns.
// It runs 'n' go-routines at a time.
func Parallel(n uint16, steps ...StepFunc) StepFunc {
	return func(ctx context.Context) error {
		ctx, cancel := context.WithCancel(ctx)
		defer cancel()
		// Semaphore for controlling parallelism level.
		sem := make(chan struct{}, n)
		wg := sync.WaitGroup{}
		wg.Add(len(steps))

		errChan := make(chan error, 1)
		for _, step := range steps {
			sem <- struct{}{}
			go func(step StepFunc) {
				defer func() {
					<-sem
					wg.Done()
				}()
				if err := step(ctx); err != nil {
					errChan <- err
				}
			}(step)
		}
		wg.Wait()
		close(errChan)
		return <-errChan
	}
}

// Run receives a context, and runs all pipeline functions.
// It runs until the first non-nil error or completion.
func Run(ctx context.Context, steps ...StepFunc) (err error) {
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()
	for _, step := range steps {
		err = step(ctx)
		if err != nil {
			return
		}
	}
	return
}

// Field receives a pointer `*field` and returns a pipeline function that assigns
// any non-nil value to the pointer, if no error is returned.
// Field is supposed to be used for changing only 1 field at a time.
// For changing more than one field at once, use `Struct`.
func Field[T any](field *T, fetch FetchFunc[T]) StepFunc {
	return func(ctx context.Context) error {
		if field == nil {
			return errors.New("field pointer is nil")
		}
		ptr, err := fetch(ctx)
		if ptr != nil {
			*field = *ptr
		}
		return err
	}
}

// Struct receives a pointer to a struct, which is used as reference for a modifying function `f`.
// It then returns a pipeline function calling `f`.
// This function must be used with caution, because running it in parallel might cause data races.
// Avoid calling a pipeline that modifies the same field twice in parallel.
// If you only want to modify 1 field, use `Field`
func Struct[T any](s *T, f func(context.Context, *T) error) StepFunc {
	return func(ctx context.Context) error {
		return f(ctx, s)
	}
}
