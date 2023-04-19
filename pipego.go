package pp

import (
	"context"
	"strings"
	"time"
)

type (
	// Context represents our internal context handler,
	// Capable of doing structured logging, sectioning and cancellations and timeouts.
	Context interface {
		context.Context
		Trace(message string, args ...any)
		Debug(message string, args ...any)
		Info(message string, args ...any)
		Error(message string, args ...any)
		Warn(message string, args ...any)
		WithTimeout(d time.Duration) (Context, CancelFunc)
		WithCancel() (Context, CancelFunc)
		WithCancelCause() (Context, CancelCausefunc)

		// SetSection is used to section your code into sections, that you can name and trace them back after the execution.
		// groupID is used to differentiate sections with same name, grouping sections under the same parent for example.
		Section(name string, msgAndArgs ...any) Context
	}

	// StepFunc is a function signature,
	// It is used to padronize function calls, making it possible to create a set of generic behaviors.
	// It's created on the pipeline initialization, and runs during the `Run` function.
	// A StepFunc is a job, that might never run, or might run until it succeeds.
	StepFunc func(ctx Context) (err error)

	// Response holds information about the pipeline execution, such as section statistics and structured log tree.
	Response struct {
		Duration time.Duration
		logTree  *logNode
	}
)

func (p Response) Logs(lv LogLevelType) (out []string) {
	return p.logTree.Logs(lv)
}

func (p Response) LogTree(lv LogLevelType) string {
	return strings.Join(p.logTree.Tree(lv), "\n")
}

// Run receives a context, and runs all pipeline functions.
// It runs until the first non-nil error or completion.
func Run(old context.Context, steps ...StepFunc) (r Response, err error) {
	t1 := time.Now()
	ctx := FromContext(old)
	r.logTree = ctx.Value(key).(*logNode)
	ctx.Trace("starting Run method with %d steps", len(steps))
	err = runSteps(ctx, steps...)
	r.Duration = time.Since(t1)
	ctx.Trace("Run method finished in %s", r.Duration)
	return r, err
}

func runSteps(ctx Context, steps ...StepFunc) error {
	var err error
	for i, step := range steps {
		ctx.Trace("running step[%d]", i)
		if err = ctx.Err(); err != nil {
			ctx.Trace("ctx is cancelled: %s. finishing execution", err)
			return err
		}
		if err = step(ctx); err != nil {
			return err
		}
	}
	return err
}
