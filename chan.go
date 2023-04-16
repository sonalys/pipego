package pp

import "sync"

// ChanWorker defines a function signature to process values returned from a channel.
type ChanWorker[T any] func(Context, T) error

// ChanDivide divides the input of a channel between all the given workers, they process load as they are free to do so.
// We only accept *<-chan T because during the initialization of the pipeline the channel field will still be unset,
// Since functions normally return the channel they will use, and we are not the only ones providing the channel to them.
func ChanDivide[T any](ch *<-chan T, workers ...ChanWorker[T]) StepFunc {
	errChan := make(chan error, 1)
	var wg sync.WaitGroup
	wg.Add(len(workers))
	return func(ctx Context) (err error) {
		ctx, cancel := ctx.WithCancel()
		defer cancel()
		for i := range workers {
			go func(i int) {
				defer wg.Done()
				for {
					select {
					case v, ok := <-*ch:
						if !ok {
							return
						}
						if err := workers[i](ctx, v); err != nil {
							errChan <- err
						}
					case <-ctx.Done():
						return
					}
				}
			}(i)
		}
		wg.Wait()
		return
	}
}
