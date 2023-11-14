package main

import (
	"context"
	"errors"
	"math/rand"
	"time"

	pp "github.com/sonalys/pipego"
	"github.com/sonalys/pipego/retry"
)

// API is a generic API implementation.
type API struct{}

// fetchData implements a generic data fetcher signature.
func (a API) fetchData(_ context.Context, id string) ([]int, error) {
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
		fetchData(_ context.Context, id string) ([]int, error)
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
	return func(ctx context.Context) (err error) {
		s.values, err = s.dep.API.fetchData(ctx, id)
		return
	}
}

func (s *Pipeline) calcSum(_ context.Context) (err error) {
	for _, v := range s.values {
		s.Sum += v
	}
	return
}

func (s *Pipeline) calcCount(_ context.Context) (err error) {
	s.Count = len(s.values)
	return
}

func (s *Pipeline) calcAverage(_ context.Context) (err error) {
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
	err := pp.Run(ctx,
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
	// {dep:{API:{}} values:[1 2 3 4 5] Sum:15 AVG:3 Count:5}
}
