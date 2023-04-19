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
	key = struct{}{}
	// DefaultLogger specifies the logger which will be used to output pipeline logs.
	DefaultLogger = log.Default()
	// LogLevel defines which level will be output to the logger.
	LogLevel = Warn
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
	Trace LogLevelType = iota
	Debug
	Info
	Warn
	Error
)

// addLog adds a log line to the current tree node.
func (t *traceTree) addLog(e pipelineError) {
	t.lock.Lock()
	defer t.lock.Unlock()
	t.logs = append(t.logs, e)
}

// Errors walks through the log tree returning all logs that match the filter.
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

// BuildLogTree walks through the log tree printing all logs that match the criteria with indentation.
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

// Error implements error interface.
func (e pipelineError) Error() string {
	return fmt.Sprintf("%s: %s", e.t.Format(time.RFC3339), e.err)
}

// initializeCtx tries to infer if the context was already initialized or not, setting
// traceTree and section accordingly.
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

// getTraceTree returns the log trace tree from context, it panics if not present.
func getTraceTree(ctx context.Context) *traceTree {
	tree, ok := ctx.Value(key).(*traceTree)
	if !ok {
		panic("logTree not in context")
	}
	return tree
}

// addNode adds a section to the traceTree.
func (t *traceTree) addNode(ctx *ppContext, id, name string) {
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

// SetSection changes context values to reference old node id as parent, and new node id as current.
func (ctx *ppContext) SetSection(groupName string, nodeID ...string) {
	tree := getTraceTree(ctx)
	var id string
	if len(nodeID) > 0 {
		id = nodeID[0]
	} else {
		id = uuid.NewString()
	}
	tree.addNode(ctx, id, groupName)
	return
}

// logMessage is responsible for adding a log to the current log node in context.
func (ctx ppContext) logMessage(lv LogLevelType, message string, args ...any) {
	e := pipelineError{
		lv:  lv,
		t:   time.Now(),
		err: fmt.Sprintf(message, args...),
	}
	getTraceTree(ctx).addLog(e)
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
