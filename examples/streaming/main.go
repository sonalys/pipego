package main

import (
	"context"
	"math/rand"
	"os"
	"time"

	pp "github.com/sonalys/pipego"
	"github.com/sonalys/pipego/retry"
	"golang.org/x/exp/slog"
)

// API is a generic API implementation.
type API struct{}

// fetchData implements a generic data fetcher signature.
func (a API) fetchData(ctx pp.Context, id string) <-chan int {
	ch := make(chan int)
	ctx, cancel := ctx.WithTimeout(10 * time.Second)
	// Note that if this wasn't async, both fetch and chanDivide would need to be insire parallel stage.
	go func() {
		defer cancel()
		defer close(ch)
		for ctx.Err() == nil {
			ch <- rand.Intn(10)
			time.Sleep(time.Second)
		}
	}()
	return ch
}

type PipelineDependencies struct {
	API interface {
		fetchData(ctx pp.Context, id string) <-chan int
	}
}

type Pipeline struct {
	dep PipelineDependencies
	// We need to use pointers with ChanDivide func because at initialization, the field is not set yet.
	values <-chan int
}

func newPipeline(dep PipelineDependencies) Pipeline {
	return Pipeline{
		dep:    dep,
		values: make(<-chan int),
	}
}

func (s *Pipeline) fetchValues(id string) pp.StepFunc {
	return func(ctx pp.Context) (err error) {
		s.values = s.dep.API.fetchData(ctx, id)
		return
	}
}

func getLogger(ctx pp.Context, sectionName string) slog.Logger {
	return *slog.New(slog.NewJSONHandler(ctx.GetWriter(), nil)).With(slog.Attr{
		Key:   "section",
		Value: slog.AnyValue(sectionName),
	})
}

func main() {
	ctx := context.Background()
	api := API{}
	pipeline := newPipeline(PipelineDependencies{
		API: api,
	})
	r, err := pp.New(
		// Setup a simple example of a streaming response.
		retry.Constant(retry.Inf, time.Second,
			pipeline.fetchValues("objectID"),
		),
		pp.ChanDivide(&pipeline.values,
			func(ctx pp.Context, i int) (err error) {
				logger := getLogger(ctx, "worker 1")
				logger.Info("got value %d", i)
				return
			},
			// Slower worker that will take longer to execute values.
			func(ctx pp.Context, i int) (err error) {
				logger := getLogger(ctx, "worker 2")
				logger.Info("got value %d", i)
				time.Sleep(2 * time.Second)
				return
			},
		),
	).
		// WithSections allows us to use our pp.Context.Section function, segmentating logs by sections
		// Pipego is capable of using reflection to automatically segmentate functions by name.
		WithOptions(
			pp.WithAutomaticSections(),
			pp.WithSections(),
		).
		Run(ctx)
	if err != nil {
		println("could not execute pipeline: ", err.Error())
	}
	r.LogTree(os.Stdout)
	// [root]
	//       [main.main.Constant.newRetry.func6] step=0
	//                       [main.main.(*Pipeline).fetchValues.func3] step=0
	//       [github.com/sonalys/pipego.ChanDivide[...].func1] step=1
	//                       [main.main.func2] step=1
	//                                       [worker 2]
	//                                               {"level":"info","message":"got value 4"}
	//                                               {"level":"info","message":"got value 6"}
	//                                               {"level":"info","message":"got value 5"}
	//                                               {"level":"info","message":"got value 5"}
	//                       [main.main.func1] step=0
	//                                       [worker 1]
	//                                               {"level":"info","message":"got value 2"}
	//                                               {"level":"info","message":"got value 1"}
	//                                               {"level":"info","message":"got value 7"}
	//                                               {"level":"info","message":"got value 3"}
	//                                               {"level":"info","message":"got value 5"}
	//                                               {"level":"info","message":"got value 4"}
}
