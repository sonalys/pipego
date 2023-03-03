package main

import (
	"context"
	"errors"
	"fmt"

	"github.com/sonalys/pipego"
)

type Object struct {
	values []int
}

type ResultObject struct {
	sum   int
	avg   int
	count int
}

func aggSum(result *ResultObject, td Object) pipego.StepFunc {
	return func(ctx context.Context) (err error) {
		for _, v := range td.values {
			result.sum += v
		}
		return nil
	}
}

func aggCount(result *ResultObject, td Object) pipego.StepFunc {
	return func(ctx context.Context) (err error) {
		result.count = len(td.values)
		return nil
	}
}

func aggAvg(result *ResultObject, td Object) pipego.StepFunc {
	return func(ctx context.Context) (err error) {
		// simple example of aggregation error.
		if result.count == 0 {
			return errors.New("cannot calculate average for empty slice")
		}
		result.avg = result.sum / result.count
		return nil
	}
}

func main() {
	ctx := context.Background()
	testData := Object{
		values: []int{1, 2, 3, 4, 5},
	}
	var result ResultObject
	// Simple example where we calculate sum and count in parallel,
	// then we calculate average, re-utilizing previous steps result.
	report, err := pipego.Run(ctx,
		pipego.Parallel(2,
			aggSum(&result, testData),
			aggCount(&result, testData),
		),
		aggAvg(&result, testData),
	)
	if err != nil {
		println(err.Error())
		return
	}
	fmt.Printf("execution took %s.\n%+v\n", report.Duration, result)
}
