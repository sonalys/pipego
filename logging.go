package pp

import (
	"fmt"
	"log"
	"strings"
	"sync"
	"time"
)

var (
	key = struct{}{}
	// DefaultLogger specifies the logger which will be used to output pipeline logs.
	DefaultLogger = log.Default()
	// LogLevel defines which level will be output to the logger.
	LogLevel = Warn
	// We use a single lock to control mutability for all logs.
	lock = sync.Mutex{}
)

type (
	LogLevelType int
	logNode      struct {
		lv       LogLevelType
		section  string
		msg      string
		children []logNode
		t        time.Time
	}
)

const (
	Trace LogLevelType = iota
	Debug
	Info
	Warn
	Error
)

func (e logNode) Logs(lv LogLevelType) []string {
	resp := make([]string, 0, len(e.children)+1)
	if e.lv >= lv {
		resp = append(resp, e.Error(0))
	}
	for i := range e.children {
		resp = append(resp, e.children[i].Logs(lv)...)
	}
	return resp
}

func (e logNode) Tree(lv LogLevelType) []string {
	return e.tree(lv, 0)
}

func (e logNode) tree(lv LogLevelType, indent int) []string {
	resp := make([]string, 0, len(e.children)+1)
	if e.lv >= lv {
		resp = append(resp, e.Error(indent))
	}
	for i := range e.children {
		resp = append(resp, e.children[i].tree(lv, indent+1)...)
	}
	return resp
}

// Error implements error interface.
func (e logNode) Error(ident int) string {
	if e.section != "" {
		return fmt.Sprintf("%s[%s] %s", strings.Repeat("\t", ident), e.section, e.msg)
	}
	return fmt.Sprintf("%s%s: %s", strings.Repeat("\t", ident), e.t.Format(time.RFC3339), e.msg)
}

// logMessage is responsible for adding a log to the current log node in context.
func (ctx ppContext) logMessage(lv LogLevelType, message string, args ...any) {
	e := logNode{
		lv:  lv,
		t:   time.Now(),
		msg: fmt.Sprintf(message, args...),
	}
	if DefaultLogger != nil && lv >= LogLevel {
		DefaultLogger.Print(e.Error(0))
	}
	lock.Lock()
	defer lock.Unlock()
	entry := ctx.Value(key).(*logNode)
	entry.children = append(entry.children, e)
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

func (ctx ppContext) Error(message string, args ...any) {
	ctx.logMessage(Error, message, args...)
}

func (ctx ppContext) Warn(message string, args ...any) {
	ctx.logMessage(Warn, message, args...)
}
