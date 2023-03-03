package pipego_test

import (
	"context"
	"testing"

	"github.com/sonalys/pipego"
	"github.com/stretchr/testify/require"
)

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

func Test_Slice(t *testing.T) {
	ctx := context.Background()
	t.Run("nil field", func(t *testing.T) {
		err := pipego.Slice(nil, func(ctx context.Context) ([]int, error) {
			return nil, nil
		})(ctx)
		require.Equal(t, pipego.NilFieldError, err)
	})
	t.Run("slice", func(t *testing.T) {
		var got []int
		f := func(ctx context.Context) ([]int, error) {
			return []int{1, 2, 3}, nil
		}
		err := pipego.Slice(&got, f)(ctx)
		require.NoError(t, err)
		require.Equal(t, []int{1, 2, 3}, got)
	})
}

func Test_Map(t *testing.T) {
	ctx := context.Background()
	t.Run("nil map", func(t *testing.T) {
		err := pipego.Map(nil, func(ctx context.Context) (map[int]int, error) {
			return nil, nil
		})(ctx)
		require.Equal(t, pipego.NilFieldError, err)
	})
	t.Run("map", func(t *testing.T) {
		var got map[int]int
		f := func(ctx context.Context) (map[int]int, error) {
			return map[int]int{1: 1}, nil
		}
		err := pipego.Map(&got, f)(ctx)
		require.NoError(t, err)
		require.Equal(t, map[int]int{1: 1}, got)
	})
}
