package pp

import (
	"context"
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
	ctx, _ = initializeCtx(ctx)
	return &ppContext{ctx}
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
