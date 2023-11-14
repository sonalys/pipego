package pp

import (
	"sync"
)

// Parallel runs all the given steps in parallel,
// It cancels context for the first non-nil error and returns.
// It runs 'n' go-routines at a time.
func Parallel(n uint16, steps ...StepFunc) StepFunc {
	return func(ctx Context) (err error) {
		if n <= 0 {
			n = uint16(len(steps))
		}
		ctx, cancel := ctx.WithCancel()
		defer cancel()
		// Semaphore for controlling parallelism level.
		sem := make(chan struct{}, n)
		wg := sync.WaitGroup{}
		wg.Add(len(steps))

		errChan := make(chan error, len(steps))
		for i, step := range steps {
			go func(i int, step StepFunc) {
				ctx = AutomaticSection(ctx, step, i)
				sem <- struct{}{}
				defer func() {
					<-sem
					wg.Done()
				}()
				if err := ctx.Err(); err != nil {
					return
				}
				if err := step(ctx); err != nil {
					errChan <- err
					cancel()
				}
			}(i, step)
		}
		// We wait for all steps to either succeed, or gracefully shutdown if context is cancelled.
		wg.Wait()
		close(errChan)
		return <-errChan
	}
}
