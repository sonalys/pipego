package pipego

import (
	"context"
	"errors"
	"sync"
)

var ZeroParallelismErr = errors.New("parallelism is set to 0")

// Parallel runs all the given steps in parallel,
// It cancels context for the first non-nil error and returns.
// It runs 'n' go-routines at a time.
func Parallel(n uint16, steps ...StepFunc) StepFunc {
	return func(ctx context.Context) error {
		if n == 0 {
			return ZeroParallelismErr
		}
		ctx = configureCtx(ctx, "parallel")
		Trace(ctx, "starting parallelism = %d with %d steps", n, len(steps))
		ctx, cancel := context.WithCancel(ctx)
		defer cancel()
		// Semaphore for controlling parallelism level.
		sem := make(chan struct{}, n)
		wg := sync.WaitGroup{}
		wg.Add(len(steps))

		errChan := make(chan error, 1)
		for i, step := range steps {
			Trace(ctx, "provided step %d is queued", i)
			sem <- struct{}{}
			Trace(ctx, "provided step %d is running", i)
			go func(i int, step StepFunc) {
				defer func() {
					<-sem
					wg.Done()
					Trace(ctx, "provided step %d is finished", i)
				}()
				if err := ctx.Err(); err != nil {
					Trace(ctx, "ctx is cancelled: %s. finishing step %d execution", err, i)
					return
				}
				if err := step(ctx); err != nil {
					Trace(ctx, "step[%d] errored: %s. finishing execution", i, err)
					errChan <- err
					cancel()
				}
			}(i, step)
		}
		Trace(ctx, "waiting tasks to finish")
		// We wait for all steps to either succeed, or gracefully shutdown if context is cancelled.
		wg.Wait()
		Trace(ctx, "closing errChan")
		close(errChan)
		Trace(ctx, "parallelism finished")
		return <-errChan
	}
}
