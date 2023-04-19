package pp

import (
	"context"
	"fmt"
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
	return &ppContext{context.WithValue(ctx, key, &logNode{
		lv:      Info,
		section: "root",
		msg:     "new context initialized",
		t:       time.Now(),
	})}
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

func (ctx ppContext) Section(name string, msgAndArgs ...any) Context {
	lock.Lock()
	defer lock.Unlock()
	entry := ctx.Value(key).(*logNode)
	var msg string
	if len(msgAndArgs) > 0 {
		msg = fmt.Sprintf(msgAndArgs[0].(string), msgAndArgs[1:]...)
	}
	entry.children = append(entry.children, logNode{
		lv:      Info,
		section: name,
		msg:     msg,
	})
	newCtx := context.WithValue(ctx.Context, key, &entry.children[len(entry.children)-1])
	ctx.Context = newCtx
	return ctx
}
