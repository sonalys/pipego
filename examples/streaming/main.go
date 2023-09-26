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
				return
			},
			// Slower worker that will take longer to execute values.
			func(_ pp.Context, _ int) (err error) {
				time.Sleep(2 * time.Second)
				return
			},
		),
	).Run(ctx)
	if err != nil {
		println("could not execute pipeline: ", err.Error())
	}
	fmt.Printf("Execution took %s.\n%+v\n", r.Duration, pipeline)
	// go run examples/streaming/main.go                                                                            1 ↵
	// 2023/04/19 09:32:02 got value on worker 2: 5
	// 2023/04/19 09:32:03 got value on worker 1: 5
	// 2023/04/19 09:32:04 got value on worker 1: 3
	// 2023/04/19 09:32:05 got value on worker 2: 1
	// 2023/04/19 09:32:06 got value on worker 1: 3
	// 2023/04/19 09:32:07 got value on worker 1: 6
	// 2023/04/19 09:32:08 got value on worker 2: 4
	// 2023/04/19 09:32:09 got value on worker 1: 9
	// 2023/04/19 09:32:10 got value on worker 1: 8
	// 2023/04/19 09:32:11 got value on worker 1: 2
	// 2023/04/19 09:32:12 got value on worker 2: 2
	// Execution took 12.00069829s.
}
