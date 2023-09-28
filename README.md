# Pipego

Pipego is a robust and type safe pipelining framework, made to improve go's error handling, while also allowing you to load balance, add fail safety, better parallelism and modularization of your \
code.

## Features

This library has support for:

- **Parallelism**: fetch data in parallel, with a cancellable context like in errorGroup implementation.
- **Retriability**: choose from constant, linear and exponential backoffs for retrying any step.
- **Load balance**: you can easily split slices and channels over go-routines using different algorithms.
- **Plug and play api**: you can implement any middleware you want on top of pipego's API.
- **Sections**: your code can automatically segmentate it's logging by reflecting the function names, and grouping sections under the same parent.

## Sections

Your code is automatically traced for each step and section, with not much effort.

This segment generates the following logs:
```go
func getLogger(ctx pp.Context, sectionName string) zerolog.Logger {
	ctx = ctx.Section(sectionName)
	return zerolog.New(ctx.GetWriter())
}

func main() {
	ctx := context.Background()
	r, err := pp.New(
		// Setup a simple example of a streaming response.
		retry.Constant(retry.Inf, time.Second,
			pipeline.fetchValues("objectID"),
		),
		pp.ChanDivide(&pipeline.values,
			func(ctx pp.Context, i int) (err error) {
				logger := getLogger(ctx, "worker 1")
				logger.Info().Msgf("got value %d", i)
				return
			},
			// Slower worker that will take longer to execute values.
			func(ctx pp.Context, i int) (err error) {
				logger := getLogger(ctx, "worker 2")
				logger.Info().Msgf("got value %d", i)
				time.Sleep(2 * time.Second)
				return
			},
		),
	).
		WithOptions(
			pp.WithAutomaticSections(),
			pp.WithSections(),
		).
		Run(ctx)
}
```
```go
[root]
        [main.main.Constant.newRetry.func6] step=0
                        [main.main.(*Pipeline).fetchValues.func3] step=0
        [github.com/sonalys/pipego.ChanDivide[...].func1] step=1
                        [main.main.func2] step=1
                                        [worker 2]
                                                {"level":"info","message":"got value 4"}
                                                {"level":"info","message":"got value 6"}
                                                {"level":"info","message":"got value 5"}
                                                {"level":"info","message":"got value 5"}
                        [main.main.func1] step=0
                                        [worker 1]
                                                {"level":"info","message":"got value 2"}
                                                {"level":"info","message":"got value 1"}
                                                {"level":"info","message":"got value 7"}
                                                {"level":"info","message":"got value 3"}
                                                {"level":"info","message":"got value 5"}
                                                {"level":"info","message":"got value 4"}
```

## Functions

### Parallel

With parallel you can run any given steps at `n` parallelism.

### Retry

You can define different retry behaviors for the given steps.

### Timeout

You define a total timeout all the steps inside should take, otherwise cancel them.

### WrapErr

You define a function that will cast any error returned by given steps to a specific error, example: integration error.

### DivideSliceInSize

Divides any given slice in groups of size `n` which can be processed parallel for example.

### DivideSliceInGroups

Does the same thing as DivideSliceInSize, but divide the slice into `n` groups instead.

### ChanDivide

Creates a pool of workers, which takes values from the provided channel `ch` as soon as the worker is available.

## Examples

All examples are under the [examples folder](./examples/)

- [Simple](./examples/simple/main.go) [Slice, Parallel, Errors]
- [Tracing](./examples/tracing/main.go) [Slice, Parallel, Errors]
- [Streaming](./examples/streaming/main.go) [Field, Warnings]
- [Aggregation](./examples/aggregation/main.go) [Slice, Parallel, Errors]

### Simple example

```go
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
	r, err := pp.Run(ctx,
		retry.Constant(3, time.Second,
			p.fetchInput("id"),
		),
		pp.Parallel(2,
			p.sumInput,
			p.sqrInput,
		),
	)
	if err != nil {
		println(err.Error())
	}
	fmt.Printf("Execution took %s.\n%#v\n", r.Duration, p)
	// Execution took 82.54µs.
	// main.pipeline{API:main.api{}, input:4, sum:8, square:16}
}
```

## Logging

To get Pipego working with your favorite logger, you just need to instantiate a new logger providing the
io.Writer you get from `ctx.GetWriter()`.

With this configuration, you can segment your logs by sections, easily visualize parallelization and multiple log
instances.

```go
package main

import (
	"context"
	"errors"
	"os"
	"time"

	"github.com/rs/zerolog"
	pp "github.com/sonalys/pipego"
	"github.com/sonalys/pipego/retry"
)

func getLogger(ctx pp.Context) zerolog.Logger {
	return zerolog.New(ctx.GetWriter()).
		With().
		Str("section", ctx.GetSection()). // You can even put the section inside your logger, for filtering later.
		Logger()
}

func funcA(ctx pp.Context) (err error) {
	log := getLogger(ctx)
	log.Info().Msg("testing info")
	return
}

func funcB(ctx pp.Context) (err error) {
	ctx = ctx.SetSection("test section")
	log := getLogger(ctx)
	log.Error().Msg("from inside section")
	return
}

func funcC(ctx pp.Context) (err error) {
	log := getLogger(ctx)
	log.Error().Msg("from inside retry")
	return errors.New("error")
}

func main() {
	ctx := context.Background()
	resp, _ := pp.New(
		pp.Parallel(2,
			funcA,
			funcB,
		),
		retry.Constant(3, time.Second,
			funcC,
		),
	).
		// WithSections allows us to use our pp.Context.Section function, segmentating logs by sections
		// Pipego is capable of using reflection to automatically segmentate functions by name.
		WithOptions(
			pp.WithAutomaticSections(),
			pp.WithSections(),
		).
		Run(ctx)
	// Note that the section log can also be customized by modifying pp.NewSectionFormatter.
	resp.LogTree(os.Stdout)
	// [root]
	//	[main.main.Parallel.func1] step=0
	//		[main.funcB] step=1
	//			[test section]
	//				{"level":"error","section":"test section","message":"from inside section"}
	//		[main.funcA] step=0
	//			{"level":"info","section":"main.funcA","message":"testing info"}
	//	[main.main.Constant.newRetry.func4] step=1
	//		[main.funcC] step=0
	//			{"level":"error","section":"main.funcC","message":"from inside retry"}
	//			{"level":"error","section":"main.funcC","message":"from inside retry"}
	//			{"level":"error","section":"main.funcC","message":"from inside retry"}
}
```

## Contributions

With Pipego, I aim to offer a sufficient framework for facilitating any popular pipeline flows.
If you have any ideas, feel free to open an issue
and discuss it's implementation with me.

Writing more unit tests and fixing any possible bugs would also be nice from any developer.

## Disclaimer

This library is not stable yet, and it's not production ready.
