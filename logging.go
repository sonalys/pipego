package pp

import (
	"bufio"
	"fmt"
	"io"
	"strings"
	"sync"

	"github.com/sonalys/pipego/internal"
)

type key int

var (
	// We use a single lock to control mutability for all logs.
	lock = sync.Mutex{}
)

type (
	LogNodeV2 struct {
		Buffer   io.ReadWriter
		Parent   int
		Section  string
		Children []int
	}
)

func (cd ContextData) FindSection(name string) int {
	for _, id := range cd.Current().Children {
		if (*cd.logs)[id].Section == name {
			return id
		}
	}
	return -1
}

func (cd ContextData) Current() LogNodeV2 {
	return (*cd.logs)[cd.current]
}

func (cd ContextData) Parent() (LogNodeV2, bool) {
	parentID := (*cd.logs)[cd.current].Parent
	if parentID == -1 {
		return LogNodeV2{}, false
	}
	return (*cd.logs)[parentID], true
}

func AutomaticSection(ctx Context, step any, i int) Context {
	if ok, _ := ctx.Value(internal.AutomaticSectionKey).(bool); ok {
		ctx = ctx.SetSection(internal.GetFunctionName(step), "step=%d", i)
	}
	return ctx
}

// NewSectionFormatter is an injectable section formatter thant can be replaced by your custom log formatter.
var NewSectionFormatter = func(sectionName string, maskAndArgs ...any) string {
	if len(maskAndArgs) > 1 {
		return fmt.Sprintf("[%s] %s\n", sectionName, fmt.Sprintf(maskAndArgs[0].(string), maskAndArgs[1:]...))
	}
	return fmt.Sprintf("[%s]\n", sectionName)
}

func (e ContextData) Tree(w io.Writer) {
	lock.Lock()
	defer lock.Unlock()
	e.tree(w, 0, 0)
}

func (cd ContextData) tree(w io.Writer, cur, indent int) {
	r := bufio.NewReader((*cd.logs)[cur].Buffer)
	// indent every line of the buffer.
	for {
		line, _, err := r.ReadLine()
		if err != nil {
			break
		}
		w.Write([]byte(strings.Repeat("  ", indent)))
		_, err = w.Write(line)
		if err != nil {
			break
		}
		w.Write([]byte("\n"))
	}
	indent++
	for _, id := range (*cd.logs)[cur].Children {
		cd.tree(w, id, indent)
	}
}
