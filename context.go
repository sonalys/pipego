package pp

import (
	"context"
	"fmt"
	"io"
	"log"
	"time"

	"github.com/sonalys/pipego/internal"
)

// ppContext represents our internal context handler,
// Capable of doing structured logging, sectioning and cancellations and timeouts.
type ppContext struct {
	context.Context
}

type CancelFunc context.CancelFunc
type CancelCausefunc context.CancelCauseFunc

// NewContext creates a new pp.Context from context.Background().
func NewContext(withSections bool) Context {
	ctx := context.Background()
	return FromContext(ctx, withSections)
}

// NewContext creates a new pp.Context from context.Background().
func FromContext(ctx context.Context, withSections bool) Context {
	node := LogNode{
		Section: []byte("root"),
		Message: []byte(NewSectionFormatter("root", "new context initialized")),
	}
	return &ppContext{
		Context: context.WithValue(ctx, contextKey, &node),
	}
}

// WithCancel is a wrapper for context.WithCancel, to facilitate with type convertion.
func (ctx ppContext) WithCancel() (Context, CancelFunc) {
	new, cancel := context.WithCancel(ctx.Context)
	return &ppContext{
		Context: new,
	}, CancelFunc(cancel)
}

// WithTimeout is a wrapper for context.WithCancel, to facilitate with type convertion.
func (ctx ppContext) WithTimeout(d time.Duration) (Context, CancelFunc) {
	new, cancel := context.WithTimeout(ctx.Context, d)
	return &ppContext{
		Context: new,
	}, CancelFunc(cancel)
}

// WithCancelCause is a wrapper for context.WithCancelCause, to facilitate with type convertion.
func (ctx ppContext) WithCancelCause() (Context, CancelCausefunc) {
	new, cancel := context.WithCancelCause(ctx.Context)
	return &ppContext{
		Context: new,
	}, CancelCausefunc(cancel)
}

func (ctx ppContext) GetWriter() io.Writer {
	node := getLogNode(ctx)
	if node == nil {
		return log.Writer()
	}
	return node
}

func getLogNode(ctx context.Context) *LogNode {
	return ctx.Value(contextKey).(*LogNode)
}

func AutomaticSection(ctx Context, step any, i int) {
	if ok, _ := ctx.Value(internal.AutomaticSectionKey).(bool); ok {
		ctx = ctx.Section(internal.GetFunctionName(step), "step=%d", i)
	}
}

func (ctx ppContext) Section(name string, msgAndArgs ...any) Context {
	if withSections, _ := ctx.Value(internal.SectionKey).(bool); withSections {
		return ctx
	}
	lock.Lock()
	defer lock.Unlock()
	entry := getLogNode(ctx)
	if entry == nil {
		return ctx
	}
	var msg string
	if len(msgAndArgs) > 0 {
		msg = NewSectionFormatter(name, fmt.Sprintf(msgAndArgs[0].(string), msgAndArgs[1:]...))
	} else {
		msg = fmt.Sprintf("[%s]\n", name)
	}
	entry.Children = append(entry.Children, LogNode{
		Section: []byte(name),
		Message: []byte(msg),
	})
	newCtx := context.WithValue(ctx.Context, contextKey, &entry.Children[len(entry.Children)-1])
	return ppContext{
		Context: newCtx,
	}
}
