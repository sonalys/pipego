package pp_test

import (
	"context"
	"testing"

	pp "github.com/sonalys/pipego"
	"github.com/stretchr/testify/require"
)

func Test_Field(t *testing.T) {
	ctx := context.Background()
	t.Run("nil field", func(t *testing.T) {
		err := pp.Field(nil, func(ctx context.Context) (*int, error) {
			return nil, nil
		})(ctx)
		require.Equal(t, pp.NilFieldError, err)
	})
	t.Run("mutating single field", func(t *testing.T) {
		var got string
		f := func(ctx context.Context) (*string, error) {
			return pointer("response"), nil
		}
		err := pp.Field(&got, f)(ctx)
		require.NoError(t, err)
		require.Equal(t, "response", got)
	})
	t.Run("mutating struct", func(t *testing.T) {
		type field struct {
			a, b int
		}
		var got field
		err := pp.Field(&got, func(ctx context.Context) (*field, error) {
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
		err := pp.Slice(nil, func(ctx context.Context) ([]int, error) {
			return nil, nil
		})(ctx)
		require.Equal(t, pp.NilFieldError, err)
	})
	t.Run("slice", func(t *testing.T) {
		var got []int
		f := func(ctx context.Context) ([]int, error) {
			return []int{1, 2, 3}, nil
		}
		err := pp.Slice(&got, f)(ctx)
		require.NoError(t, err)
		require.Equal(t, []int{1, 2, 3}, got)
	})
	t.Run("embed slice", func(t *testing.T) {
		var got struct {
			data []int
		}
		f := func(ctx context.Context) ([]int, error) {
			return []int{1, 2, 3}, nil
		}
		err := pp.Slice(&got.data, f)(ctx)
		require.NoError(t, err)
		require.Equal(t, []int{1, 2, 3}, got.data)
	})
}

func Test_Map(t *testing.T) {
	ctx := context.Background()
	t.Run("nil map", func(t *testing.T) {
		err := pp.Map(nil, func(ctx context.Context) (map[int]int, error) {
			return nil, nil
		})(ctx)
		require.Equal(t, pp.NilFieldError, err)
	})
	t.Run("map", func(t *testing.T) {
		var got map[int]int
		f := func(ctx context.Context) (map[int]int, error) {
			return map[int]int{1: 1}, nil
		}
		err := pp.Map(&got, f)(ctx)
		require.NoError(t, err)
		require.Equal(t, map[int]int{1: 1}, got)
	})
}
