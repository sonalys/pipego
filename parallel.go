package pp

import (
	"errors"
	"sync"
	"time"
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
		t1 := time.Now()
		ctx.SetSection("parallel")
		ctx.Trace("starting parallelism = %d with %d steps", n, len(steps))
		ctx, cancel := ctx.WithCancel()
		defer cancel()
		// Semaphore for controlling parallelism level.
		sem := make(chan struct{}, n)
		wg := sync.WaitGroup{}
		wg.Add(len(steps))

		errChan := make(chan error, 1)
		for i, step := range steps {
			ctx.Trace("step[%d] is queued", i)
			sem <- struct{}{}
			ctx.Trace("step[%d] is running", i)
			go func(i int, step StepFunc) {
				defer func() {
					<-sem
					wg.Done()
					ctx.Trace("step[%d] is finished", i)
				}()
				if err := ctx.Err(); err != nil {
					ctx.Trace("ctx is cancelled: %s. finishing execution", err)
					return
				}
				if err := step(ctx); err != nil {
					ctx.Trace("step[%d] failed: %s. finishing execution", i, err)
					errChan <- err
					cancel()
				}
			}(i, step)
		}
		ctx.Trace("waiting tasks to finish")
		// We wait for all steps to either succeed, or gracefully shutdown if context is cancelled.
		wg.Wait()
		ctx.Trace("closing errChan")
		close(errChan)
		ctx.Trace("Parallel method finished in %s", time.Since(t1))
		return <-errChan
	}
}
