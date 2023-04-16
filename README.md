# Pipego

Pipego is a robust and type safe pipelining framework, made to improve go's error handling, while also allowing you to load balance, add fail safety, better parallelism and modularization of your \
code.

## Features

This library has support for:

- **Parallelism**: fetch data in parallel, with a cancellable context like in errorGroup implementation.
- **Retriability**: choose from constant, linear and exponential backoffs for retrying any step.
- **Load balance**: you can easily split slices and channels over go-routines using different algorithms.
- **Plug and play api**: you can implement any middleware you want on top of pipego's API.

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
	pp.LogLevel = pp.ErrLevelDebug
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
			retry.Retry(3, retry.Constant(time.Second),
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
}

```
