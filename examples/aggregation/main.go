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
	// Here we are simply implementing a deterministic failure mechanism to test our retriability.
	rnd := rand.New(rand.NewSource(2))
	// 33% chance of returning a slice of integers, or failing.
	switch rnd.Intn(3) {
	case 0, 1:
		return nil, errors.New("unexpected error")
	default:
		return []int{1, 2, 3, 4, 5}, nil
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

func (s *Pipeline) fetchValues(ctx pp.Context) (err error) {
	s.values, err = s.dep.API.fetchData(ctx, "id")
	return
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
		retry.Retry(retry.Inf, retry.Constant(time.Second),
			pipeline.fetchValues,
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
	// println(report.LogTree(pp.ErrLevelTrace))
	fmt.Printf("Execution took %s.\n%+v\n", r.Duration, pipeline)
}
