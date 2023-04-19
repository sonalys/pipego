package main

import (
	"context"
	"errors"
	"fmt"
	"math/rand"
	"time"

	pp "github.com/sonalys/pipego"
	"github.com/sonalys/pipego/retry"
)

// API is a generic API implementation.
type API struct{}

// fetchData implements a generic data fetcher signature.
func (a API) fetchData(_ pp.Context, id string) ([]int, error) {
	// Here we are simply implementing a failure mechanism to test our retriability.
	switch n := rand.Intn(10); {
	case n < 3:
		return []int{1, 2, 3, 4, 5}, nil
	default:
		return nil, errors.New("unexpected error")
	}
}

type PipelineDependencies struct {
	API interface {
		fetchData(_ pp.Context, id string) ([]int, error)
	}
}

type Pipeline struct {
	dep PipelineDependencies

	values []int

	Sum   int
	AVG   int
	Count int
}

func newPipeline(dep PipelineDependencies) Pipeline {
	return Pipeline{dep: dep}
}

func (s *Pipeline) fetchValues(id string) pp.StepFunc {
	return func(ctx pp.Context) (err error) {
		s.values, err = s.dep.API.fetchData(ctx, id)
		return
	}
}

func (s *Pipeline) calcSum(_ pp.Context) (err error) {
	for _, v := range s.values {
		s.Sum += v
	}
	return
}

func (s *Pipeline) calcCount(_ pp.Context) (err error) {
	s.Count = len(s.values)
	return
}

func (s *Pipeline) calcAverage(_ pp.Context) (err error) {
	// simple example of aggregation error.
	if s.Count == 0 {
		return errors.New("cannot calculate average for empty slice")
	}
	s.AVG = s.Sum / s.Count
	return
}

func main() {
	ctx := context.Background()
	api := API{}
	pipeline := newPipeline(PipelineDependencies{
		API: api,
	})
	r, err := pp.Run(ctx,
		retry.Constant(retry.Inf, time.Second,
			pipeline.fetchValues("objectID"),
		),
		pp.Parallel(2,
			pipeline.calcSum,
			pipeline.calcCount,
		),
		pipeline.calcAverage,
	)
	if err != nil {
		println("could not execute pipeline: ", err.Error())
	}
	fmt.Printf("Execution took %s.\n%+v\n", r.Duration, pipeline)
	// 2023/04/19 09:24:46 fetched 5 objects
	// Execution took 157.831µs.
	// {dep:{API:{}} values:[1 2 3 4 5] Sum:15 AVG:3 Count:5}
}
