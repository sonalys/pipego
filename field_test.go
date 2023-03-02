package pipego_test

import (
	"context"
	"testing"

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

func Test_Field(t *testing.T) {
	ctx := context.Background()
	t.Run("nil field", func(t *testing.T) {
		err := pipego.Field(nil, func(ctx context.Context) (*int, error) {
			return nil, nil
		})(ctx)
		require.Equal(t, pipego.NilFieldError, err)
	})
	t.Run("mutating single field", func(t *testing.T) {
		var got string
		f := func(ctx context.Context) (*string, error) {
			return pointer("response"), nil
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
