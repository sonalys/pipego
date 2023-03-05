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

type Object struct {
	values []int
}

type ResultObject struct {
	sum   int
	avg   int
	count int
}

func sum(result *ResultObject, td Object) pp.StepFunc {
	return func(ctx context.Context) (err error) {
		for _, v := range td.values {
			result.sum += v
		}
		return nil
	}
}

func count(result *ResultObject, td Object) pp.StepFunc {
	return func(ctx context.Context) (err error) {
		result.count = len(td.values)
		return nil
	}
}

func average(result *ResultObject, td Object) pp.StepFunc {
	return func(ctx context.Context) (err error) {
		// simple example of aggregation error.
		if result.count == 0 {
			return errors.New("cannot calculate average for empty slice")
		}
		result.avg = result.sum / result.count
		return nil
	}
}

type API struct{}

func (a *API) fetchData(id string) pp.FetchSlice[int] {
	rnd := rand.New(rand.NewSource(1))
	return func(ctx context.Context) ([]int, error) {
		switch rnd.Intn(3) {
		case 0, 1:
			return nil, errors.New("unexpected error")
		default:
			return []int{1, 2, 3, 4, 5}, nil
		}
	}
}

func main() {
	ctx := context.Background()
	api := &API{}
	var data Object
	var result ResultObject
	// Simple example where we calculate sum and count in parallel,
	// then we calculate average, re-utilizing previous steps result.
	report, err := pp.Run(ctx,
		retry.Retry(5, retry.ConstantRetry(time.Second),
			pp.Slice(&data.values, api.fetchData("dataID"))),
		pp.Parallel(2,
			sum(&result, data),
			count(&result, data),
		),
		average(&result, data),
	)
	if err != nil {
		println("could not execute pipeline: ", err.Error())
	}
	fmt.Printf("Execution took %s.\n%+v\n", report.Duration, result)
	println(report.LogTree(pp.ErrLevelTrace))
}
