package pp

import (
	"context"
	"fmt"
	"log"
	"strings"
	"sync"
	"time"

	"github.com/google/uuid"
)

var (
	key           = struct{}{}
	DefaultLogger = log.Default()
	LogLevel      = Warn
)

type (
	LogLevelType  int
	pipelineError struct {
		lv  LogLevelType
		err string
		t   time.Time
	}
	traceTree struct {
		id        string
		name      string
		logs      []pipelineError
		children  []traceTree
		lock      sync.Mutex
		createdAt time.Time
	}
)

const (
	Trace LogLevelType = 0
	Debug LogLevelType = 1
	Info  LogLevelType = 2
	Warn  LogLevelType = 3
	Error LogLevelType = 4
)

func (t *traceTree) AddLog(e pipelineError) {
	t.lock.Lock()
	defer t.lock.Unlock()
	t.logs = append(t.logs, e)
}

func (t *traceTree) AddChild(ctx *ppContext, id, name string) {
	t.lock.Lock()
	defer t.lock.Unlock()
	t.children = append(t.children, traceTree{
		id:        id,
		name:      name,
		lock:      sync.Mutex{},
		createdAt: time.Now(),
	})
	ctx.Context = context.WithValue(ctx.Context, key, &t.children[len(t.children)-1])
}

func (t *traceTree) Errors(lv LogLevelType) []error {
	errs := make([]error, 0, len(t.children)+len(t.logs))
	for i := range t.logs {
		if t.logs[i].lv >= lv {
			errs = append(errs, t.logs[i])
		}
	}
	for i := range t.children {
		errs = append(errs, t.children[i].Errors(lv)...)
	}
	return errs
}

func (t *traceTree) getPipelineErrors() []pipelineError {
	out := make([]pipelineError, len(t.logs))
	copy(out, t.logs)
	for i := range t.children {
		out = append(out, t.children[i].getPipelineErrors()...)
	}
	return out
}

func (t *traceTree) BuildLogTree(level LogLevelType, ident int) string {
	var b strings.Builder
	if ident > 0 {
		b.WriteString(strings.Repeat("\t", ident-1))
	}
	b.WriteString(fmt.Sprintf("[%s] %s\n", t.id, t.name))
	writeLine := func(msg string, args ...any) {
		b.WriteString(strings.Repeat("\t", ident))
		b.WriteString(fmt.Sprintf(msg, args...))
		b.WriteRune('\n')
	}
	for i := range t.logs {
		writeLine(t.logs[i].Error())
	}
	for i := range t.children {
		b.WriteString(t.children[i].BuildLogTree(level, ident+1))
	}
	return b.String()
}

func (e pipelineError) Error() string {
	return fmt.Sprintf("%s: %s", e.t.Format(time.RFC3339), e.err)
}

func initializeCtx(old context.Context) (*ppContext, *traceTree) {
	ctx := ppContext{old}
	tree, ok := ctx.Value(key).(*traceTree)
	// In case we are running nested Run commands, we can still rebuild chain of logs from the true root.
	if !ok {
		tree = &traceTree{
			id:        uuid.NewString(),
			name:      "root",
			lock:      sync.Mutex{},
			createdAt: time.Now(),
		}
		ctx = ppContext{context.WithValue(ctx, key, tree)}
	}
	ctx.SetSection("run")
	return &ctx, tree
}

func getTraceTree(ctx context.Context) *traceTree {
	tree, ok := ctx.Value(key).(*traceTree)
	if !ok {
		panic("logTree not in context")
	}
	return tree
}

// SetSection changes context values to reference old node id as parent, and new node id as current.
func (ctx *ppContext) SetSection(groupName string, nodeID ...string) {
	tree := getTraceTree(ctx)
	var id string
	if len(nodeID) > 0 {
		id = nodeID[0]
	} else {
		id = uuid.NewString()
	}
	tree.AddChild(ctx, id, groupName)
	return
}

func (ctx ppContext) logMessage(lv LogLevelType, message string, args ...any) {
	if lv < LogLevel {
		return
	}
	e := pipelineError{
		lv:  lv,
		t:   time.Now(),
		err: fmt.Sprintf(message, args...),
	}
	getTraceTree(ctx).AddLog(e)
	if DefaultLogger != nil && lv >= LogLevel {
		DefaultLogger.Print(e.err)
	}
}

func (ctx ppContext) Trace(message string, args ...any) {
	ctx.logMessage(Trace, message, args...)
}

func (ctx ppContext) Debug(message string, args ...any) {
	ctx.logMessage(Debug, message, args...)
}

func (ctx ppContext) Info(message string, args ...any) {
	ctx.logMessage(Info, message, args...)
}

func (ctx ppContext) Warn(message string, args ...any) {
	ctx.logMessage(Warn, message, args...)
}
