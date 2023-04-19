package pp

import (
	"errors"
	"fmt"
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
		ctx = ctx.Section("parallel", "n = %d steps = %d", n, len(steps))
		ctx, cancel := ctx.WithCancel()
		defer cancel()
		// Semaphore for controlling parallelism level.
		sem := make(chan struct{}, n)
		wg := sync.WaitGroup{}
		wg.Add(len(steps))

		errChan := make(chan error, 1)
		for i, step := range steps {
			go func(i int, step StepFunc) {
				ctx := ctx.Section(fmt.Sprintf("step-%d", i))
				ctx.Trace("queued")
				sem <- struct{}{}
				ctx.Trace("running")
				defer func() {
					<-sem
					wg.Done()
					ctx.Trace("finished")
				}()
				if err := step(ctx); err != nil {
					ctx.Trace("failed with error: %s", err)
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
		ctx.Trace("parallel method finished in %s", time.Since(t1))
		return <-errChan
	}
}
