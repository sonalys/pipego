package main

import (
	"context"
	"fmt"
	"math/rand"
	"time"

	pp "github.com/sonalys/pipego"
	"github.com/sonalys/pipego/retry"
)

// API is a generic API implementation.
type API struct{}

// fetchData implements a generic data fetcher signature.
func (a API) fetchData(ctx pp.Context, id string) <-chan int {
	ch := make(chan int)
	ctx, cancel := ctx.WithTimeout(10 * time.Second)
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
	values *<-chan int
}

func newPipeline(dep PipelineDependencies) Pipeline {
	return Pipeline{
		dep:    dep,
		values: new(<-chan int),
	}
}

func (s *Pipeline) fetchValues(id string) pp.StepFunc {
	return func(ctx pp.Context) (err error) {
		*s.values = s.dep.API.fetchData(ctx, id)
		return
	}
}

func main() {
	pp.LogLevel = pp.Info
	ctx := context.Background()
	api := API{}
	pipeline := newPipeline(PipelineDependencies{
		API: api,
	})
	r, err := pp.Run(ctx,
		// Setup a simple example of a streaming response.
		retry.Constant(retry.Inf, time.Second,
			pipeline.fetchValues("objectID"),
		),
		pp.ChanDivide(pipeline.values,
			func(ctx pp.Context, i int) (err error) {
				ctx.Info("got value on worker 1: %d", i)
				return
			},
			// Slower worker that will take longer to execute values.
			func(ctx pp.Context, i int) (err error) {
				ctx.Info("got value on worker 2: %d", i)
				time.Sleep(2 * time.Second)
				return
			},
		),
	)
	if err != nil {
		println("could not execute pipeline: ", err.Error())
	}
	// println(report.LogTree(pp.ErrLevelTrace))
	fmt.Printf("Execution took %s.\n%+v\n", r.Duration, pipeline)
}
