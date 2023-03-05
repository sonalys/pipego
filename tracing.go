package pipego

import (
	"context"
	"fmt"
	"log"
	"strings"
	"sync"
	"time"

	"github.com/google/uuid"
)

var key = struct{}{}

type ErrLevel int

const (
	ErrLevelTrace ErrLevel = 0
	ErrLevelDebug ErrLevel = 1
	ErrLevelInfo  ErrLevel = 2
	ErrLevelWarn  ErrLevel = 3
	ErrLevelError ErrLevel = 4
)

type traceTree struct {
	id        string
	name      string
	logs      []pipelineError
	children  []traceTree
	lock      sync.Mutex
	createdAt time.Time
}

func (t *traceTree) AddLog(e pipelineError) {
	t.lock.Lock()
	defer t.lock.Unlock()
	t.logs = append(t.logs, e)
}

func (t *traceTree) AddChild(ctx context.Context, id, name string) context.Context {
	t.lock.Lock()
	defer t.lock.Unlock()
	t.children = append(t.children, traceTree{
		id:        id,
		name:      name,
		lock:      sync.Mutex{},
		createdAt: time.Now(),
	})
	return context.WithValue(ctx, key, &t.children[len(t.children)-1])
}

func (t *traceTree) Errors(lv ErrLevel) []error {
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

func (t *traceTree) BuildLogTree(level ErrLevel, ident int) string {
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

type pipelineError struct {
	lv  ErrLevel
	err string
	t   time.Time
}

func (e pipelineError) Error() string {
	return fmt.Sprintf("%s: %s", e.t.Format(time.RFC3339), e.err)
}

func initializeCtx(ctx context.Context) (context.Context, *traceTree) {
	tree, ok := ctx.Value(key).(*traceTree)
	// In case we are running nested Run commands, we can still rebuild chain of logs from the true root.
	if ok {
		ctx = ConfigureCtx(ctx, "run")
	} else {
		tree = &traceTree{
			id:        uuid.NewString(),
			name:      "root",
			lock:      sync.Mutex{},
			createdAt: time.Now(),
		}
		ctx = context.WithValue(ctx, key, tree)
	}
	return ctx, tree
}

func getTraceTree(ctx context.Context) *traceTree {
	tree, ok := ctx.Value(key).(*traceTree)
	if !ok {
		panic("logTree not in context")
	}
	return tree
}

// ConfigureCtx changes context values to reference old node id as parent, and new node id as current.
func ConfigureCtx(ctx context.Context, groupName string, nodeID ...string) context.Context {
	tree := getTraceTree(ctx)
	var id string
	if len(nodeID) > 0 {
		id = nodeID[0]
	} else {
		id = uuid.NewString()
	}
	return tree.AddChild(ctx, id, groupName)
}

var DefaultLogger = log.Default()
var DefaultLoglevel = ErrLevelWarn

func logMessage(ctx context.Context, lv ErrLevel, message string, args ...any) {
	e := pipelineError{
		lv:  lv,
		t:   time.Now(),
		err: fmt.Sprintf(message, args...),
	}
	getTraceTree(ctx).AddLog(e)
	if DefaultLogger != nil && lv >= DefaultLoglevel {
		DefaultLogger.Print(e.err)
	}
}

func Trace(ctx context.Context, message string, args ...any) {
	logMessage(ctx, ErrLevelTrace, message, args...)
}

func Debug(ctx context.Context, message string, args ...any) {
	logMessage(ctx, ErrLevelDebug, message, args...)
}

func Info(ctx context.Context, message string, args ...any) {
	logMessage(ctx, ErrLevelInfo, message, args...)
}

func Warn(ctx context.Context, message string, args ...any) {
	logMessage(ctx, ErrLevelWarn, message, args...)
}
