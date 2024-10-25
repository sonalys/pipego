package pp

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestDivideSliceInSize(t *testing.T) {
	mockFactory := func() (*int, func(i int) Step) {
		var count int
		return &count, func(_ int) Step {
			count++
			return func(ctx context.Context) (err error) {
				return nil
			}
		}
	}
	t.Run("empty", func(t *testing.T) {
		count, empty := mockFactory()
		steps := DivideSliceInSize([]int{}, 1, empty)
		require.Empty(t, steps)
		require.Equal(t, 0, *count)
	})
	t.Run("exact size", func(t *testing.T) {
		count, empty := mockFactory()
		steps := DivideSliceInSize([]int{1, 2, 3, 4}, 2, empty)
		require.Len(t, steps, 2)
		require.Equal(t, 4, *count)
	})
	t.Run("non matching size", func(t *testing.T) {
		count, empty := mockFactory()
		steps := DivideSliceInSize([]int{1, 2, 3, 4}, 3, empty)
		require.Len(t, steps, 2)
		require.Equal(t, 4, *count)
	})
	t.Run("size bigger than slice", func(t *testing.T) {
		count, empty := mockFactory()
		steps := DivideSliceInSize([]int{1, 2, 3, 4}, 5, empty)
		require.Len(t, steps, 1)
		require.Equal(t, 4, *count)
	})
}

func TestDivideSliceInGroups(t *testing.T) {
	mockFactory := func() (*int, func(i int) Step) {
		var count int
		return &count, func(_ int) Step {
			count++
			return func(ctx context.Context) (err error) {
				return nil
			}
		}
	}
	t.Run("empty", func(t *testing.T) {
		count, empty := mockFactory()
		steps := DivideSliceInGroups([]int{}, 1, empty)
		require.Empty(t, steps)
		require.Equal(t, 0, *count)
	})
	t.Run("exact size", func(t *testing.T) {
		count, empty := mockFactory()
		steps := DivideSliceInGroups([]int{1, 2, 3, 4}, 2, empty)
		require.Len(t, steps, 2)
		require.Equal(t, 4, *count)
	})
	t.Run("non matching size", func(t *testing.T) {
		count, empty := mockFactory()
		steps := DivideSliceInGroups([]int{1, 2, 3, 4}, 3, empty)
		require.Len(t, steps, 3)
		require.Equal(t, 4, *count)
	})
	t.Run("size bigger than slice", func(t *testing.T) {
		count, empty := mockFactory()
		steps := DivideSliceInGroups([]int{1, 2, 3, 4}, 5, empty)
		require.Len(t, steps, 4)
		require.Equal(t, 4, *count)
	})
}
