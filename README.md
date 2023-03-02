# Pipego

Pipego is a robust and type safe pipelining framework, made to improve go's error handling, while also allowing you to load balance, add fail safety, better parallelism and modularization of your \
code.

## Features

This library has support for:
* **Parallelism**: fetch data in parallel, with a cancellable context like in errorGroup implementation.
* **Retriability**: choose from constant, linear and exponential backoffs for retrying any step.
* **Load balance**: you can easily split slices and channels over go-routines using different algorithms.
* **Plug and play api**: you can implement any middleware you want on top of pipego's API.

## Examples

### Data Fetching
```go
type DatabaseObject struct{}

// Usual function we have on our APIs
fetchAPI := func(ctx context.Context, id string) (*DatabaseObject, error) {
	return &DatabaseObject{}, nil
}

// With pipego, you can handle everything in one place:
// You need to declare once a wrapper for the fetch function
wrappedFetch := func(id string) FetchFunc[DatabaseObject] {
	return func(ctx context.Context) (*DatabaseObject, error) {
		return fetchAPI(ctx, id)
	}
}

func main() {
	var data struct {
		a1 *DatabaseObject
		a2 *DatabaseObject
		a3 *DatabaseObject
	}
	ctx := context.Background()
	err := pipego.Run(ctx,
		pipego.Field(data.a1, wrappedFetch("a1")),
		pipego.Field(data.a2, wrappedFetch("a2")),
		pipego.Field(data.a3, wrappedFetch("a3")),
	)
	// Single line error check, instead of multiple lines.
	if err != nil {
		return
	}
}
```

### Parallelism and retriability
```go
...
err = pipego.Run(ctx,
	// Retries the children steps 3 times, linearly.
	pipego.Retry(3, pipego.LinearRetry(time.Second),
		// Set 2 go-routines for fetching data in parallel.
		pipego.Parallel(2,
			pipego.Field(data.a1, wrappedFetch("a1")),
			pipego.Field(data.a2, wrappedFetch("a2")),
			pipego.Field(data.a3, wrappedFetch("a3")),
		),
	),
	// Here you can go by both approaches, either using a compact wrapped version
	pipego.Field(data.a1, wrappedFetch("a1")),
	// or inlining your own wrapper for fetching the data.
	func(ctx context.Context) (err error) {
		data.a1, err = fetchAPI(ctx, "a1")
		return
	},
)
if err != nil {
	return
}
```

### Aggregation

```go
func main() {
	type data struct {
		values []int
	}
	var testData data
	testData.values = []int{1, 2, 3, 4, 5}

	ctx := context.Background()
	var result struct {
		sum   int
		avg   int
		count int
	}
	aggSum := func(td *data) pipego.StepFunc {
		return func(ctx context.Context) (err error) {
			for _, v := range td.values {
				result.sum += v
			}
			return
		}
	}
	aggCount := func(td *data) pipego.StepFunc {
		return func(ctx context.Context) (err error) {
			result.count = len(td.values)
			return
		}
	}
	aggAvg := func(ctx context.Context) (err error) {
			// simple example of aggregation error.
			if result.count == 0 {
				return errors.New("cannot calculate average for empty slice")
			}
			result.avg = result.sum / result.count
			return
		}
	// Example where we calculate sum and count in parallel,
	// then we calculate average, re-utilizing previous steps result.
	err := pipego.Run(ctx,
		pipego.Parallel(2,
			aggSum(&testData),
			aggCount(&testData),
		),
		aggAvg,
	)
	if err != nil {
		return
	}
	println(result.avg) // 3
}
```

### Load Balancing

```go
func main() {
	var data struct {
		values []int
	}
	data.values = []int{1, 2, 3, 4, 5, 6, 7, 8, 9}
	var sum int64
	sumAgg := func(slice []int) pipego.StepFunc {
		return func(ctx context.Context) error {
			var localSum int
			for _, v := range slice {
				localSum += v
			}
			atomic.AddInt64(&sum, int64(localSum))
			return nil
		}
	}
	ctx := context.Background()
	err := pipego.Run(ctx,
		// Divide the slice into groups of 3 elements and process in parallel using 3 go-routines.
		pipego.Parallel(3,
			divide.DivideSize(data.values, 3, sumAgg)...,
		),
	)
	if err != nil {
		return
	}
	println(sum) // 45
}
```