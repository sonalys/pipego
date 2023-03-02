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
					cancel()
				}
			}(step)
		}
		// We wait for all steps to either succeed, or gracefully shutdown if context is cancelled.
		wg.Wait()
		close(errChan)
		return <-errChan
	}
}
