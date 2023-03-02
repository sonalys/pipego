package pipego_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/sonalys/pipego"
	"github.com/sonalys/pipego/retry"
	"github.com/stretchr/testify/require"
)

func Test_Run(t *testing.T) {
	ctx := context.Background()
	t.Run("no steps", func(t *testing.T) {
		err := pipego.Run(ctx)
		require.NoError(t, err)
	})
	t.Run("with steps", func(t *testing.T) {
		run := false
		err := pipego.Run(ctx, func(ctx context.Context) (err error) {
			run = true
			return
		})
		require.NoError(t, err)
		require.True(t, run)
	})
	t.Run("keep step order", func(t *testing.T) {
		var i int
		err := pipego.Run(ctx,
			func(ctx context.Context) (err error) {
				require.Equal(t, 0, i)
				i++
				return
			},
			func(ctx context.Context) (err error) {
				require.Equal(t, 1, i)
				i++
				return
			},
		)
		require.NoError(t, err)
		require.Equal(t, 2, i)
	})
	t.Run("stop on error", func(t *testing.T) {
		err := pipego.Run(ctx,
			func(ctx context.Context) (err error) {
				return pipego.NilFieldError
			},
			func(ctx context.Context) (err error) {
				require.Fail(t, "should not run")
				return
			},
		)
		require.Equal(t, pipego.NilFieldError, err)
	})
}

func Test_Example1(t *testing.T) {
	// Object model from database.
	type DatabaseObject struct{}
	// Usual function we have on our APIs
	fetchAPI := func(ctx context.Context, id string) (*DatabaseObject, error) {
		return &DatabaseObject{}, nil
	}
	// Adapted function for fitting inside pipego.Field wrapper.
	adaptedFetch := func(id string) pipego.FetchFunc[DatabaseObject] {
		return func(ctx context.Context) (*DatabaseObject, error) {
			return fetchAPI(ctx, id)
		}
	}
	type pipelineData struct {
		a1 *DatabaseObject
		a2 *DatabaseObject
		a3 *DatabaseObject
	}
	data := pipelineData{}
	ctx := context.Background()

	var err error
	// Old methodology vs new one:
	// Old one:
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
	// New one:
	err = pipego.Run(ctx,
		pipego.Field(data.a1, adaptedFetch("a1")),
		pipego.Field(data.a2, adaptedFetch("a2")),
		pipego.Field(data.a3, adaptedFetch("a3")),
	)
	if err != nil {
		require.Fail(t, err.Error())
		return
	}
	// You can also easily add logic on top of it, like parallelism and retries.
	err = pipego.Run(ctx,
		retry.Retry(3, retry.LinearRetry(time.Second),
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
	// Single line error check, instead of multiple lines.
	if err != nil {
		require.Fail(t, err.Error())
		return
	}
}

func Test_Aggregation(t *testing.T) {
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
	aggSum := func(td data) pipego.StepFunc {
		return func(ctx context.Context) (err error) {
			for _, v := range td.values {
				result.sum += v
			}
			return nil
		}
	}
	aggCount := func(td data) pipego.StepFunc {
		return func(ctx context.Context) (err error) {
			result.count = len(td.values)
			return nil
		}
	}
	aggAvg := func(ctx context.Context) (err error) {
		// simple example of aggregation error.
		if result.count == 0 {
			return errors.New("cannot calculate average for empty slice")
		}
		result.avg = result.sum / result.count
		return nil
	}
	// Simple example where we calculate sum and count in parallel,
	// then we calculate average, re-utilizing previous steps result.
	err := pipego.Run(ctx,
		pipego.Parallel(2,
			aggSum(testData),
			aggCount(testData),
		),
		aggAvg,
	)
	require.NoError(t, err)
	require.EqualValues(t, 3, result.avg)
}
