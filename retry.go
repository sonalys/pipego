package pipego

import (
	"context"
	"math"
	"time"
)

type Retrier interface {
	Retry(retryNumber int) time.Duration
}

type constantRetry struct {
	time.Duration
}

func (r constantRetry) Retry(n int) time.Duration {
	return r.Duration
}

// ConstantRetry is a constant retry implementation.
// It always return the same delay.
// Example: 2s: 2s, 2s, 2s, 2s...
func ConstantRetry(delay time.Duration) Retrier {
	return constantRetry{delay}
}

type linearRetry struct {
	time.Duration
}

func (r linearRetry) Retry(n int) time.Duration {
	return r.Duration * time.Duration(n)
}

// LinearRetry is a linear retry implementation.
// It returns a linear series for the delay calculation.
// Example: 1s: 1s, 2s, 3s, 4s, ...
func LinearRetry(delay time.Duration) Retrier {
	return linearRetry{delay}
}

type expRetry struct {
	initial time.Duration
	max     time.Duration
	exp     float64
}

func (r expRetry) Retry(n int) time.Duration {
	delay := time.Duration(math.Pow(float64(n), r.exp))*time.Second + r.initial
	if delay > r.max {
		return r.max
	}
	return delay
}

// ExpRetry is a exponential retry implementation.
// Given an initialDelay, it does (initialDelay * n) ^ exp.
// Example: n ^ 2 + 1s = 1s, 3s, 9s...
func ExpRetry(initialDelay, maxDelay time.Duration, exp float64) Retrier {
	return expRetry{initialDelay, maxDelay, exp}
}

// Retry implements a pipeline step for retrying all children steps inside.
// If retries = -1, it will retry until it succeeds.
func Retry(retries int, r Retrier, steps ...StepFunc) StepFunc {
	return func(ctx context.Context) (err error) {
		for _, step := range steps {
			for n := 0; n < retries || retries == -1; n++ {
				if err = step(ctx); err == nil {
					break
				}
				time.Sleep(r.Retry(n))
			}
		}
		return err
	}
}
