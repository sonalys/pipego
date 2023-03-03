package pipego

import (
	"context"
	"fmt"
	"time"
)

// StepFunc is a function signature,
// It is used to padronize function calls, making it possible to create a set of generic behaviors.
// It's created on the pipeline initialization, and runs during the `Run` function.
// A StepFunc is a job, that might never run, or might run until it succeeds.
type StepFunc func(ctx context.Context) (err error)

type PipelineReport struct {
	Duration time.Duration
	Warnings []error
}

var warnKey struct{}

func Warn(ctx context.Context, message string, args ...any) {
	ch, ok := ctx.Value(warnKey).(chan error)
	if !ok {
		panic("warnings not in context")
	}
	ch <- fmt.Errorf(message, args...)
}

// Run receives a context, and runs all pipeline functions.
// It runs until the first non-nil error or completion.
func Run(ctx context.Context, steps ...StepFunc) (report PipelineReport, err error) {
	t1 := time.Now()
	var warnCh = make(chan error, 0)
	go func() {
		for v := range warnCh {
			report.Warnings = append(report.Warnings, v)
		}
	}()
	ctx = context.WithValue(ctx, warnKey, warnCh)
	for _, step := range steps {
		err = step(ctx)
		// Exits if there is error or context is cancelled.
		if err != nil || ctx.Err() != nil {
			return
		}
	}
	close(warnCh)
	report.Duration = time.Since(t1)
	return
}
