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

// NewContext creates a new pp.Context from context.Background().
func NewContext() Context {
	return &ppContext{context.Background()}
}

// WithCancel is a wrapper for context.WithCancel, to facilitate with type convertion.
func (ctx ppContext) WithCancel() (Context, context.CancelFunc) {
	new, cancel := context.WithCancel(ctx.Context)
	return &ppContext{new}, cancel
}

// WithTimeout is a wrapper for context.WithCancel, to facilitate with type convertion.
func (ctx ppContext) WithTimeout(d time.Duration) (Context, context.CancelFunc) {
	new, cancel := context.WithTimeout(ctx.Context, d)
	return &ppContext{new}, cancel
}
