package pp

import (
	"errors"
	"sync"
	"time"

	"github.com/google/uuid"
)

var TimeoutErr = errors.New("timeout")

// cancellableStep is an utility function to help wait steps until step is executed or context is done.
func cancellableStep(ctx Context, step StepFunc) error {
	out := make(chan error, 1)
	go func() {
		out <- step(ctx)
	}()
	select {
	case <-ctx.Done():
		return nil
	case err := <-out:
		return err
	}
}

// Timeout limits all children steps to execute in the given duration,
// the timer starts when the first step is run.
// All steps shares the same timeout.
func Timeout(d time.Duration, steps ...StepFunc) (out []StepFunc) {
	out = make([]StepFunc, 0, len(steps))
	var startAt time.Time
	var startOnce *sync.Once
	setStart := func() {
		startAt = time.Now()
	}
	timeoutID := uuid.NewString()
	for _, step := range steps {
		out = append(out, func(ctx Context) (err error) {
			ctx.SetSection("timeout", timeoutID)
			startOnce.Do(setStart)
			remainingTime := d - time.Now().Sub(startAt)
			if remainingTime < 0 {
				return TimeoutErr
			}
			ctx, cancel := ctx.WithTimeout(remainingTime)
			defer cancel()
			return cancellableStep(ctx, step)
		})
	}
	return
}
