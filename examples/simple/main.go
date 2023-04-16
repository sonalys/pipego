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
	result int
}

func (p *pipeline) fetchInput(id string) pp.StepFunc {
	return func(ctx pp.Context) (err error) {
		ctx.Debug("fetching data for id: %s", id)
		p.input, err = p.API.fetch(ctx, id)
		ctx.Debug("response is %d with err: %v", p.input, err)
		return
	}
}

func (p *pipeline) process(ctx pp.Context) (err error) {
	p.result = p.input * p.input
	ctx.Debug("result is %d", p.result)
	return
}

func main() {
	ctx := context.Background()
	pp.LogLevel = pp.Debug
	p := pipeline{
		API: api{},
	}
	r, err := pp.Run(ctx,
		retry.Retry(3, retry.Constant(time.Second),
			p.fetchInput("id"),
		),
		p.process,
	)
	if err != nil {
		println(err.Error())
	}
	fmt.Printf("Execution took %s.\n%+v\n", r.Duration, p.result)
}
