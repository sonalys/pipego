package pp

import (
	"context"
)

type (
	// StepFunc is a function signature,
	// It is used to padronize function calls, making it possible to create a set of generic behaviors.
	// It's created on the pipeline initialization, and runs during the `Run` function.
	// A StepFunc is a job, that might never run, or might run until it succeeds.
	StepFunc func(ctx context.Context) (err error)

	Steps []StepFunc
)

// Run receives a context, and runs all pipeline functions.
// It runs until the first non-nil error or completion.
func Run(ctx context.Context, steps ...StepFunc) error {
	return runSteps(ctx, steps...)
}

func runSteps(ctx context.Context, steps ...StepFunc) error {
	var err error
	for _, step := range steps {
		if err = ctx.Err(); err != nil {
			return err
		}
		if err = step(ctx); err != nil {
			return err
		}
	}
	return err
}

func (s Steps) Group() func(context.Context) error {
	return func(ctx context.Context) (err error) {
		return runSteps(ctx, s...)
	}
}