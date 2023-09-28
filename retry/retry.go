package retry

import (
	"math"
	"time"

	pp "github.com/sonalys/pipego"
)

const Inf = -1

type Retrier interface {
	Retry(retryNumber int) time.Duration
}

type constantRetry struct {
	time.Duration
}

func (r constantRetry) Retry(n int) time.Duration {
	return r.Duration
}

// Constant is a constant retry implementation.
// It always return the same delay.
// Example: 2s: 2s, 2s, 2s, 2s...
func Constant(n int, delay time.Duration, steps ...pp.StepFunc) pp.StepFunc {
	return newRetry(n, constantRetry{delay}, steps...)
}

type linearRetry struct {
	time.Duration
}

func (r linearRetry) Retry(n int) time.Duration {
	return r.Duration * time.Duration(n)
}

// Linear is a linear retry implementation.
// It returns a linear series for the delay calculation.
// Example: 1s: 1s, 2s, 3s, 4s, ...
func Linear(n int, delay time.Duration, steps ...pp.StepFunc) pp.StepFunc {
	return newRetry(n, linearRetry{delay}, steps...)
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

// Exp is a exponential retry implementation.
// Given an initialDelay, it does (initialDelay * n) ^ exp.
// Example: n ^ 2 + 1s = 1s, 3s, 9s...
func Exp(n int, initialDelay, maxDelay time.Duration, exp float64, steps ...pp.StepFunc) pp.StepFunc {
	return newRetry(n, expRetry{initialDelay, maxDelay, exp}, steps...)
}

// Retry implements a pipeline step for retrying all children steps inside.
// If retries = -1, it will retry until it succeeds.
func newRetry(retries int, r Retrier, steps ...pp.StepFunc) pp.StepFunc {
	return func(ctx pp.Context) (err error) {
		for i, step := range steps {
			ctx = pp.AutomaticSection(ctx, step, i)
			for n := 0; n < retries || retries == -1; n++ {
				if err = step(ctx); err == nil {
					break
				}
				time.Sleep(r.Retry(n))
			}
			if err != nil {
				return err
			}
		}
		return err
	}
}
