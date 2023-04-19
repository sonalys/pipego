# Pipego

Pipego is a robust and type safe pipelining framework, made to improve go's error handling, while also allowing you to load balance, add fail safety, better parallelism and modularization of your \
code.

## Features

This library has support for:

- **Parallelism**: fetch data in parallel, with a cancellable context like in errorGroup implementation.
- **Retriability**: choose from constant, linear and exponential backoffs for retrying any step.
- **Load balance**: you can easily split slices and channels over go-routines using different algorithms.
- **Plug and play api**: you can implement any middleware you want on top of pipego's API.
- **Sections**: Divide your code execution into sections, you will be able to retrieve structured logs
  by logLevel and section.

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
		ctx.Debug("fetching data for id: %s", id)
		p.input, err = p.API.fetch(ctx, id)
		ctx.Debug("response is %d with err: %v", p.input, err)
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
	pp.LogLevel = pp.Error
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

You can use pipego.Context to log any errors and divide your code into sections.

You can set the log level and output with:

```go
func main() {
 pp.DefaultLogger = nil
 pp.LogLevel = pp.Error
}
```

```go
package main

import (
	"context"
	"errors"
	"fmt"
	"time"

	pp "github.com/sonalys/pipego"
	"github.com/sonalys/pipego/retry"
)

func main() {
	ctx := context.Background()
	// Do not print any logs in stdOut.
	pp.DefaultLogger = nil
	report, err := pp.Run(ctx,
		pp.Parallel(2,
			func(ctx pp.Context) (err error) {
				ctx.Warn("parallel 1 warn")
				return
			},
			func(ctx pp.Context) (err error) {
				ctx.Info("parallel 2 info")
				return
			},
			retry.Constant(3, time.Second,
				func(ctx pp.Context) (err error) {
					return errors.New("error")
				},
			),
		),
	)
	if err != nil {
		println(err.Error())
	}
	fmt.Printf("finished in %s with %d warnings.\n\n", report.Duration, len(report.Logs(pp.ErrLevelWarn)))

	println("reconstructing log tree:")
	println(report.LogTree(pp.ErrLevelTrace))
	// 	error
	// finished in 3.001202993s with 4 warnings.
	//
	// reconstructing log tree:
	// [root] new context initialized
	//       2023-04-19T16:08:26+02:00: starting Run method with 1 steps
	//       2023-04-19T16:08:26+02:00: running step[0]
	//       [parallel] n = 2 steps = 3
	//               2023-04-19T16:08:26+02:00: waiting tasks to finish
	//               [step-2]
	//                       2023-04-19T16:08:26+02:00: queued
	//                       2023-04-19T16:08:26+02:00: running
	//                       [retry] n = 3 r = retry.constantRetry
	//                               2023-04-19T16:08:26+02:00: retry failed #1: error
	//                               2023-04-19T16:08:27+02:00: retry failed #2: error
	//                               2023-04-19T16:08:28+02:00: retry failed #3: error
	//               [step-0]
	//                       2023-04-19T16:08:26+02:00: queued
	//                       2023-04-19T16:08:26+02:00: running
	//                       2023-04-19T16:08:26+02:00: parallel 0 warn
	//                       2023-04-19T16:08:26+02:00: finished
	//               [step-1]
	//                       2023-04-19T16:08:26+02:00: queued
	//                       2023-04-19T16:08:26+02:00: running
	//                       2023-04-19T16:08:26+02:00: parallel 1 info
	//                       2023-04-19T16:08:26+02:00: finished
	//               2023-04-19T16:08:29+02:00: closing errChan
	//               2023-04-19T16:08:29+02:00: parallel method finished in 3.00027916s
	//       2023-04-19T16:08:29+02:00: Run method finished in 3.00030725s
}

```
