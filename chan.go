package pp

import "sync"

// ChanWorker defines a function signature to process values returned from a channel.
type ChanWorker[T any] func(Context, T) error

// ChanDivide divides the input of a channel between all the given workers,
// they process load as they are free to do so.
// We only accept *<-chan T because during the initialization of the pipeline the channel
// field will still be unset.
// Please provide an initialized channel pointer, using nil pointers will result in panic.
// ChanDivide must be used inside a `parallel` section,
// unless the channel providing values is in another go-routine.
// ChanDivide and the provided chan in the same go-routine will dead-lock.
func ChanDivide[T any](ch *<-chan T, workers ...ChanWorker[T]) StepFunc {
	if ch == nil {
		panic("cannot use nil chan pointer")
	}
	// We define a waitGroup to wait for all worker's routines to end.
	var wg sync.WaitGroup
	wg.Add(len(workers))
	// We also define an errChan to get the first error to happen and return it.
	errChan := make(chan error, len(workers))
	return func(ctx Context) (err error) {
		ctx, cancel := ctx.WithCancel()
		defer cancel()
		for i := range workers {
			// Spawns 1 routine for each worker, making them consume from job channel.
			go func(i int) {
				defer wg.Done()
				for {
					select {
					// Case for worker waiting for a job.
					case v, ok := <-*ch:
						// Job channel is closed, all waiting workers should end.
						if !ok {
							return
						}
						stepCtx := AutomaticSection(ctx, workers[i], i)
						// Execute job and cancel other jobs in case of error.
						if err := workers[i](stepCtx, v); err != nil {
							errChan <- err
							cancel()
							return
						}
					// Context cancellation, all jobs must end.
					case <-ctx.Done():
						return
					}
				}
			}(i)
		}
		wg.Wait()
		close(errChan)
		return <-errChan
	}
}
