package main

import (
	"context"
	"fmt"
	"time"

	pp "github.com/sonalys/pipego"
	"github.com/sonalys/pipego/retry"
)

type api struct{}

func (a api) fetch(ctx context.Context, id string) (int, error) {
	return 4, nil
}

type pipeline struct {
	API interface {
		fetch(ctx context.Context, id string) (int, error)
	}

	input  int
	sum    int
	square int
}

func (p *pipeline) fetchInput(id string) pp.StepFunc {
	return func(ctx pp.Context) (err error) {
		p.input, err = p.API.fetch(ctx, id)
		return
	}
}

func (p *pipeline) sumInput(ctx pp.Context) (err error) {
	p.sum = p.input + p.input
	return
}

func (p *pipeline) sqrInput(ctx pp.Context) (err error) {
	p.square = p.input * p.input
	return
}

func main() {
	ctx := context.Background()
	p := pipeline{
		API: api{},
	}
	r, err := pp.New(
		retry.Constant(3, time.Second,
			p.fetchInput("id"),
		),
		pp.Parallel(2,
			p.sumInput,
			p.sqrInput,
		),
	).Run(ctx)
	if err != nil {
		println(err.Error())
	}
	fmt.Printf("Execution took %s.\n%#v\n", r.Duration, p)
	// Execution took 82.54µs.
	// main.pipeline{API:main.api{}, input:4, sum:8, square:16}
}
