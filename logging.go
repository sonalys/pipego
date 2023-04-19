package pp

import (
	"fmt"
	"io"
	"strings"
	"sync"
	"time"
)

var (
	contextKey = struct{}{}
	// We use a single lock to control mutability for all logs.
	lock = sync.Mutex{}
)

type (
	LogNode struct {
		Section   []byte
		Message   []byte
		Children  []LogNode
		Timestamp time.Time
	}
)

// NewSectionFormatter is an injectable section formatter thant can be replaced by your custom log formatter.
var NewSectionFormatter = func(sectionName string, description string) string {
	return fmt.Sprintf("[%s] %s\n", sectionName, description)
}

func (l *LogNode) Write(p []byte) (n int, err error) {
	lock.Lock()
	defer lock.Unlock()
	msg := make([]byte, len(p))
	copy(msg, p)
	l.Children = append(l.Children, LogNode{
		Section:   l.Section,
		Message:   msg,
		Timestamp: time.Now(),
	})
	return len(p), nil
}

func (e LogNode) Logs(w io.Writer) {
	w.Write(e.Message)
	for i := range e.Children {
		e.Children[i].Logs(w)
	}
	return
}

func (e LogNode) String() string {
	return string(e.Message)
}

func (e LogNode) Tree(w io.Writer) {
	lock.Lock()
	defer lock.Unlock()
	e.tree(w, 0)
}

func (e LogNode) tree(w io.Writer, indent int) {
	w.Write([]byte(strings.Repeat("\t", indent)))
	w.Write(e.Message)
	indent++
	for i := range e.Children {
		e.Children[i].tree(w, indent)
	}
}
