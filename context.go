package pp

import (
	"bytes"
	"context"
	"io"
	"log"
	"time"

	"github.com/sonalys/pipego/internal"
)

type (
	// ppContext represents our internal context handler,
	// Capable of doing structured logging, sectioning and cancellations and timeouts.
	ppContext struct {
		context.Context
	}

	CancelFunc      context.CancelFunc
	CancelCausefunc context.CancelCauseFunc

	ContextData struct {
		logs    *[]LogNodeV2
		current int
	}
)

var contextKey = key(-1)

// NewContext creates a new pp.Context from context.Background().
func NewContext() Context {
	ctx := context.Background()
	return FromContext(ctx)
}

// NewContext creates a new pp.Context from context.Background().
func FromContext(ctx context.Context) Context {
	logs := []LogNodeV2{
		{
			Parent: -1,
			Buffer: bytes.NewBufferString("[root]"),
		},
	}
	return &ppContext{
		Context: context.WithValue(ctx, contextKey, ContextData{
			logs:    &logs,
			current: 0,
		}),
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
	cd, ok := ctx.Value(contextKey).(ContextData)
	if !ok {
		return log.Writer()
	}
	return cd.Current().Buffer
}

// GetSection returns the section name this context is in.
func (ctx ppContext) GetSection() string {
	cd, ok := ctx.Value(contextKey).(ContextData)
	if !ok {
		return ""
	}
	cur := cd.Current()
	for {
		if cur.Section != "" {
			return cur.Section
		}
		cur, ok = cd.Parent()
		if !ok {
			return ""
		}
	}
}

func (ctx ppContext) SetSection(name string, msgAndArgs ...any) Context {
	cd, ok := ctx.Value(contextKey).(ContextData)
	if !ok {
		return ctx
	}
	if withSections, _ := ctx.Value(internal.SectionKey).(bool); !withSections {
		return ctx
	}
	lock.Lock()
	defer lock.Unlock()

	sectionIndex := cd.FindSection(name)
	lenLogs := len(*cd.logs)
	var cur int
	// If a section header already exists, we don't need to create a new one.
	if sectionIndex == -1 {
		*cd.logs = append(*cd.logs,
			// Section header with indentation level X.
			LogNodeV2{
				Parent:   cd.current,
				Section:  name,
				Buffer:   bytes.NewBufferString(NewSectionFormatter(name, msgAndArgs)),
				Children: []int{lenLogs + 1},
			},
			LogNodeV2{
				Parent:  lenLogs,
				Section: name,
				Buffer:  bytes.NewBuffer([]byte{}),
			},
		)
		// Update parent ref to children.
		(*cd.logs)[cd.current].Children = append((*cd.logs)[cd.current].Children, lenLogs)
		cur = lenLogs + 1
	} else {
		// Existing section headers will only have 1 child.
		cur = (*cd.logs)[sectionIndex].Children[0]
	}
	return ppContext{
		Context: context.WithValue(ctx.Context, contextKey, ContextData{
			logs:    cd.logs,
			current: cur,
		}),
	}
}
