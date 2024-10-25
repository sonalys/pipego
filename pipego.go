package pp

import (
	"context"
)

type (
	// Step is a function signature,
	// It is used to padronize function calls, making it possible to create a set of generic behaviors.
	// It's created on the pipeline initialization, and runs during the `Run` function.
	// A Step is a job, that might never run, or might run until it succeeds.
	Step func(ctx context.Context) (err error)

	Steps []Step
)

// Run receives a context, and runs all pipeline functions.
// It runs until the first non-nil error or completion.
func Run(ctx context.Context, steps ...Step) error {
	return runSteps(ctx, steps...)
}

func runSteps(ctx context.Context, steps ...Step) error {
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

func (s Steps) Parallel(n uint16) Step {
	return Parallel(n, s...)
}
