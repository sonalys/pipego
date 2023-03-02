# Pipego

Pipego is a pipelining framework, made to improve go's error handling, while also allowing you \
to remove responsibilities and better modularize your code.

## Features

This library has support for:
* Parallelism: fetch data in parallel, with a cancellable context like in errorGroup implementation.
* Retriability: choose from constant, linear and exponential backoffs for retrying any step.
* Plug and play api: you can implement any middleware you want on top of pipego's API.

## Examples

```go
type DatabaseObject struct{}

// Usual function we have on our APIs
fetchAPI := func(ctx context.Context, id string) (*DatabaseObject, error) {
	return &DatabaseObject{}, nil
}

func main() {
	type pipelineData struct {
		a1 *DatabaseObject
		a2 *DatabaseObject
		a3 *DatabaseObject
	}

	data := pipelineData{}
	ctx := context.Background()

	var err error
	// Without pipego, you would have to do something like this:
	data.a1, err = fetchAPI(ctx, "a1")
	if err != nil {
		return
	}
	data.a2, err = fetchAPI(ctx, "a2")
	if err != nil {
		return
	}
	data.a3, err = fetchAPI(ctx, "a3")
	if err != nil {
		return
	}
	// With pipego, you can handle everything in one place:
	err = pipego.Run(ctx,
		pipego.Field(data.a1, adaptedFetch("a1")),
		pipego.Field(data.a2, adaptedFetch("a2")),
		pipego.Field(data.a3, adaptedFetch("a3")),
	)
	// Single line error check, instead of multiple lines.
	if err != nil {
		return
	}
	// And more...
	err = pipego.Run(ctx,
		// Retries the children steps 3 times, linearly.
		pipego.Retry(3, pipego.LinearRetry(time.Second),
			// Set 2 go-routines for fetching data in parallel.
			pipego.Parallel(2,
				pipego.Field(data.a1, adaptedFetch("a1")),
				pipego.Field(data.a2, adaptedFetch("a2")),
				pipego.Field(data.a3, adaptedFetch("a3")),
			),
		),
		// Here you can go by both approaches, either using a compact wrapped version
		pipego.Field(data.a1, adaptedFetch("a1")),
		// or inlining your own wrapper for fetching the data.
		func(ctx context.Context) (err error) {
			data.a1, err = fetchAPI(ctx, "a1")
			return
		},
	)
	if err != nil {
		return
	}
}
```