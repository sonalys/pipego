package pp

import (
	"context"
	"io"
	"time"
)

type (
	// Context represents our internal context handler,
	// Capable of doing structured logging, sectioning and cancellations and timeouts.
	Context interface {
		context.Context

		WithTimeout(d time.Duration) (Context, CancelFunc)
		WithCancel() (Context, CancelFunc)
		WithCancelCause() (Context, CancelCausefunc)

		// SetSection is used to section your code into sections, that you can name and trace them back after the execution.
		// groupID is used to differentiate sections with same name, grouping sections under the same parent for example.
		Section(name string, msgAndArgs ...any) Context
		// GetWriter is used to get current section io.Writer, this way you can plug and play
		// with any golang's logger library by pointing it torwards this io.Writer on every step you want.
		GetWriter() io.Writer
	}

	// StepFunc is a function signature,
	// It is used to padronize function calls, making it possible to create a set of generic behaviors.
	// It's created on the pipeline initialization, and runs during the `Run` function.
	// A StepFunc is a job, that might never run, or might run until it succeeds.
	StepFunc func(ctx Context) (err error)

	// Response holds information about the pipeline execution, such as section statistics and structured log tree.
	Response struct {
		Duration time.Duration
		LogNode  *LogNode
	}
)

// Run receives a context, and runs all pipeline functions.
// It runs until the first non-nil error or completion.
func Run(old context.Context, steps ...StepFunc) (r Response, err error) {
	t1 := time.Now()
	ctx := FromContext(old)
	r.LogNode = getLogNode(ctx)
	err = runSteps(ctx, steps...)
	r.Duration = time.Since(t1)
	return r, err
}

func runSteps(ctx Context, steps ...StepFunc) error {
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
