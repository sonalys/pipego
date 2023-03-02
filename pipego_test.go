package pipego_test

import (
	"context"
	"testing"
	"time"

	"github.com/sonalys/pipego"
	"github.com/stretchr/testify/require"
)

type pipelineData struct {
	a int
	b int
	c int
}

type pipelineResponse struct {
	sum   int
	avg   int
	count int
}

func aggSum(d *pipelineData) func(context.Context) (*int, error) {
	return func(ctx context.Context) (*int, error) {
		sum := d.a + d.b + d.c
		return &sum, nil
	}
}

func aggAvg(d *pipelineData) func(context.Context) (*int, error) {
	return func(ctx context.Context) (*int, error) {
		avg := (d.a + d.b + d.c) / 3
		return &avg, nil
	}
}

func aggCount(d *pipelineData) func(context.Context) (*int, error) {
	return func(ctx context.Context) (*int, error) {
		count := 3
		return &count, nil
	}
}

func fetchInt(v int) func(context.Context) (*int, error) {
	return func(ctx context.Context) (*int, error) {
		return &v, nil
	}
}

func Pointer[T any](v T) *T {
	return &v
}

func Test_Field(t *testing.T) {
	ctx := context.Background()

	t.Run("mutating single field", func(t *testing.T) {
		var got string
		f := func(ctx context.Context) (*string, error) {
			return Pointer("response"), nil
		}
		err := pipego.Field(&got, f)(ctx)
		require.NoError(t, err)
		require.Equal(t, "response", got)
	})

	t.Run("mutating struct", func(t *testing.T) {
		type field struct {
			a, b int
		}
		var got field
		err := pipego.Field(&got, func(ctx context.Context) (*field, error) {
			return &field{
				a: 1,
			}, nil
		})(ctx)
		require.NoError(t, err)
		require.Equal(t, field{
			a: 1,
		}, got)
	})
}

func Test_Struct(t *testing.T) {
	ctx := context.Background()

	t.Run("mutating one field", func(t *testing.T) {
		type str struct {
			a, b int
		}
		var got str
		err := pipego.Struct(&got, func(ctx context.Context, t *str) error {
			t.a = 1
			return nil
		})(ctx)
		require.NoError(t, err)
		require.Equal(t, str{
			a: 1,
		}, got)
	})
	t.Run("mutating two fields", func(t *testing.T) {
		type str struct {
			a, b int
		}
		var got str
		err := pipego.Struct(&got, func(ctx context.Context, t *str) error {
			t.a = 1
			return nil
		})(ctx)
		require.NoError(t, err)
		err = pipego.Struct(&got, func(ctx context.Context, t *str) error {
			t.b = 2
			return nil
		})(ctx)
		require.NoError(t, err)
		require.Equal(t, str{
			a: 1,
			b: 2,
		}, got)
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
		pipego.Retry(3, pipego.LinearRetry(time.Second),
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
