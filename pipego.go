package pipego

import (
	"context"
	"time"
)

// StepFunc is a function signature,
// It is used to padronize function calls, making it possible to create a set of generic behaviors.
// It's created on the pipeline initialization, and runs during the `Run` function.
// A StepFunc is a job, that might never run, or might run until it succeeds.
type StepFunc func(ctx context.Context) (err error)

type PipelineReport struct {
	Duration time.Duration
	logTree  *traceTree
}

func (p PipelineReport) Logs(level ErrLevel) (out []error) {
	return p.logTree.Errors(level)
}

func (p PipelineReport) LogTree(level ErrLevel) string {
	return p.logTree.BuildLogTree(level, 0)
}

// Run receives a context, and runs all pipeline functions.
// It runs until the first non-nil error or completion.
func Run(ctx context.Context, steps ...StepFunc) (report PipelineReport, err error) {
	t1 := time.Now()
	ctx, report.logTree = initializeCtx(ctx)
	Trace(ctx, "starting Run method with %d steps", len(steps))
	err = runSteps(ctx, steps...)
	report.Duration = time.Since(t1)
	Trace(ctx, "Run method finished in %s", report.Duration)
	return
}

func runSteps(ctx context.Context, steps ...StepFunc) (err error) {
	for i, step := range steps {
		Trace(ctx, "running step[%d]", i)
		if err = ctx.Err(); err != nil {
			Trace(ctx, "ctx is cancelled: %s. finishing execution", err)
			return
		}
		if err = step(ctx); err != nil {
			Trace(ctx, "step[%d] errored: %s. finishing execution", i, err)
			return
		}
	}
	return
}
