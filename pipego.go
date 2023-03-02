package pipego

import (
	"context"
)

// StepFunc is a function signature,
// It is used to padronize function calls, making it possible to create a set of generic behaviors.
// It's created on the pipeline initialization, and runs during the `Run` function.
// A StepFunc is a job, that might never run, or might run until it succeeds.
type StepFunc func(ctx context.Context) (err error)

// Run receives a context, and runs all pipeline functions.
// It runs until the first non-nil error or completion.
func Run(ctx context.Context, steps ...StepFunc) (err error) {
	for _, step := range steps {
		err = step(ctx)
		// Exits if there is error or context is cancelled.
		if err != nil || ctx.Err() != nil {
			return
		}
	}
	return
}
