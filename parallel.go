package pp

import (
	"errors"
	"fmt"
	"sync"
)

var ZeroParallelismErr = errors.New("parallelism is set to 0")

// Parallel runs all the given steps in parallel,
// It cancels context for the first non-nil error and returns.
// It runs 'n' go-routines at a time.
func Parallel(n uint16, steps ...StepFunc) StepFunc {
	return func(ctx Context) (err error) {
		if n == 0 {
			return ZeroParallelismErr
		}
		ctx = ctx.Section("parallel", "n = %d steps = %d", n, len(steps))
		ctx, cancel := ctx.WithCancel()
		defer cancel()
		// Semaphore for controlling parallelism level.
		sem := make(chan struct{}, n)
		wg := sync.WaitGroup{}
		wg.Add(len(steps))

		errChan := make(chan error, len(steps))
		for i, step := range steps {
			go func(i int, step StepFunc) {
				ctx := ctx.Section(fmt.Sprintf("step-%d", i))
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
