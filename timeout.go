package pp

import (
	"context"
	"sync"
	"time"
)

// Timeout limits all children steps to execute in the given duration,
// the timer starts when the first step is run.
// All steps shares the same timeout.
func Timeout(d time.Duration, steps ...Step) (out Steps) {
	out = make(Steps, 0, len(steps))
	getTimer := sync.OnceValue(func() *time.Timer { return time.NewTimer(d) })
	for _, step := range steps {
		enclosedStep := func(ctx context.Context) (err error) {
			// Sets a cancellable context bounded to a unique timer, started when the first step is run.
			ctx, cancel := context.WithCancelCause(ctx)
			defer cancel(context.DeadlineExceeded)

			resultCh := make(chan error, 1)
			defer close(resultCh)

			go func() {
				resultCh <- step(ctx)
			}()

			select {
			case <-ctx.Done():
				return ctx.Err()
			case <-getTimer().C:
				cancel(context.DeadlineExceeded)
				return context.DeadlineExceeded
			case err := <-resultCh:
				return err
			}
		}
		out = append(out, enclosedStep)
	}
	return
}
