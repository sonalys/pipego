package pp

import (
	"context"
	"io"
	"time"

	"github.com/sonalys/pipego/internal"
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
		SetSection(name string, msgAndArgs ...any) Context
		GetSection() string
		GetPath() string
		// GetWriter is used to get current section io.Writer, this way you can plug and play
		// with any golang's logger library by pointing it torwards this io.Writer on every step you want.
		GetWriter() io.Writer
	}

	// StepFunc is a function signature,
	// It is used to padronize function calls, making it possible to create a set of generic behaviors.
	// It's created on the pipeline initialization, and runs during the `Run` function.
	// A StepFunc is a job, that might never run, or might run until it succeeds.
	StepFunc func(ctx Context) (err error)

	Pipeline struct {
		withSections          bool
		withAutomaticSections bool
		steps                 []StepFunc
	}

	// Response holds information about the pipeline execution, such as section statistics and structured log tree.
	Response struct {
		// Duration for the pipeline execution to finish with or without errors.
		Duration time.Duration
		// If the pipeline is initialized with WithLogging flag, this field will allow you to fetch sectioned logging per
		// step, being able to know in which function each log came from.
		contextData ContextData
	}
)

func (r Response) LogTree(w io.Writer) {
	r.contextData.Tree(w)
}

func New(steps ...StepFunc) Pipeline {
	return Pipeline{
		withSections: false,
		steps:        steps,
	}
}

func WithSections() pipelineOption {
	return func(p Pipeline) Pipeline {
		p.withSections = true
		return p
	}
}

func WithAutomaticSections() pipelineOption {
	return func(p Pipeline) Pipeline {
		p.withAutomaticSections = true
		return p
	}
}

type pipelineOption func(p Pipeline) Pipeline

func (p Pipeline) WithOptions(opts ...pipelineOption) Pipeline {
	for _, f := range opts {
		p = f(p)
	}
	return p
}

// Run receives a context, and runs all pipeline functions.
// It runs until the first non-nil error or completion.
func (p Pipeline) Run(old context.Context) (r Response, err error) {
	t1 := time.Now()
	old = context.WithValue(old, internal.AutomaticSectionKey, p.withAutomaticSections)
	old = context.WithValue(old, internal.SectionKey, p.withSections)
	ctx := FromContext(old)
	if p.withSections {
		r.contextData = ctx.Value(contextKey).(ContextData)
	}
	err = runSteps(ctx, p.steps...)
	r.Duration = time.Since(t1)
	return r, err
}

func runSteps(ctx Context, steps ...StepFunc) error {
	var err error
	for i, step := range steps {
		if err = ctx.Err(); err != nil {
			return err
		}
		stepCtx := AutomaticSection(ctx, step, i)
		if err = step(stepCtx); err != nil {
			return err
		}
	}
	return err
}
