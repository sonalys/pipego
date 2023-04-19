package pp

import (
	"context"
	"fmt"
	"io"
	"time"
)

// ppContext represents our internal context handler,
// Capable of doing structured logging, sectioning and cancellations and timeouts.
type ppContext struct {
	context.Context
}

type CancelFunc context.CancelFunc
type CancelCausefunc context.CancelCauseFunc

// NewContext creates a new pp.Context from context.Background().
func NewContext() Context {
	ctx := context.Background()
	return FromContext(ctx)
}

// NewContext creates a new pp.Context from context.Background().
func FromContext(ctx context.Context) Context {
	node := LogNode{
		Section: []byte("root"),
		Message: []byte(NewSectionFormatter("root", "new context initialized")),
	}
	return &ppContext{context.WithValue(ctx, contextKey, &node)}
}

// WithCancel is a wrapper for context.WithCancel, to facilitate with type convertion.
func (ctx ppContext) WithCancel() (Context, CancelFunc) {
	new, cancel := context.WithCancel(ctx.Context)
	return &ppContext{new}, CancelFunc(cancel)
}

// WithTimeout is a wrapper for context.WithCancel, to facilitate with type convertion.
func (ctx ppContext) WithTimeout(d time.Duration) (Context, CancelFunc) {
	new, cancel := context.WithTimeout(ctx.Context, d)
	return &ppContext{new}, CancelFunc(cancel)
}

// WithCancelCause is a wrapper for context.WithCancelCause, to facilitate with type convertion.
func (ctx ppContext) WithCancelCause() (Context, CancelCausefunc) {
	new, cancel := context.WithCancelCause(ctx.Context)
	return &ppContext{new}, CancelCausefunc(cancel)
}

func (ctx ppContext) GetWriter() io.Writer {
	node := getLogNode(ctx)
	return node
}

func getLogNode(ctx Context) *LogNode {
	return ctx.Value(contextKey).(*LogNode)
}

func (ctx ppContext) Section(name string, msgAndArgs ...any) Context {
	lock.Lock()
	defer lock.Unlock()
	entry := getLogNode(ctx)
	var msg string
	if len(msgAndArgs) > 0 {
		msg = NewSectionFormatter(name, fmt.Sprintf(msgAndArgs[0].(string), msgAndArgs[1:]...))
	}
	entry.Children = append(entry.Children, LogNode{
		Section: []byte(name),
		Message: []byte(msg),
	})
	newCtx := context.WithValue(ctx.Context, contextKey, &entry.Children[len(entry.Children)-1])
	return ppContext{newCtx}
}
