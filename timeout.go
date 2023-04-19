package pp

import (
	"errors"
	"sync"
	"time"

	"github.com/google/uuid"
)

var TimeoutErr = errors.New("timeout")

// Timeout limits all children steps to execute in the given duration,
// the timer starts when the first step is run.
// All steps shares the same timeout.
func Timeout(d time.Duration, steps ...StepFunc) (out []StepFunc) {
	out = make([]StepFunc, 0, len(steps))
	var once sync.Once
	var timer *time.Timer
	timeoutID := uuid.NewString()
	for _, step := range steps {
		out = append(out, func(ctx Context) (err error) {
			ctx.SetSection("timeout", timeoutID)
			// Sets a cancellable context bounded to a unique timer, started when the first step is run.
			ctx, cancel := ctx.WithCancelCause()
			once.Do(func() {
				timer = time.NewTimer(d)
			})
			out := make(chan error, 1)
			defer close(out)
			go func() {
				out <- step(ctx)
			}()
			// Select gets the first channel to return a result,
			// either context cancellation, timeout or response from step.
			select {
			case <-ctx.Done():
				return ctx.Err()
			case <-timer.C:
				cancel(TimeoutErr)
				return TimeoutErr
			case err := <-out:
				return err
			}
		})
	}
	return
}
