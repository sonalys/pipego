package divide

import (
	"context"
	"sync/atomic"
	"testing"

	"github.com/sonalys/pipego"
	"github.com/stretchr/testify/require"
)

func TestDivideSize(t *testing.T) {
	getGroups := func() (*[][]int, func(slice []int) pipego.StepFunc) {
		group := [][]int{}
		return &group, func(slice []int) pipego.StepFunc {
			group = append(group, slice)
			return nil
		}
	}
	t.Run("by size", func(t *testing.T) {
		t.Run("empty", func(t *testing.T) {
			_, factory := getGroups()
			got := DivideSize([]int{}, 5, factory)
			require.Empty(t, got)
		})
		t.Run("slice is smaller", func(t *testing.T) {
			groups, factory := getGroups()
			_ = DivideSize([]int{1, 2, 3, 4}, 5, factory)
			require.Equal(t, [][]int{
				{1, 2, 3, 4},
			}, *groups)
		})
		t.Run("slice is bigger", func(t *testing.T) {
			_, factory := getGroups()
			got := DivideSize([]int{1, 2, 3, 4}, 2, factory)
			require.Len(t, got, 2)
		})
		t.Run("odd length", func(t *testing.T) {
			groups, factory := getGroups()
			_ = DivideSize([]int{1, 2, 3, 4}, 3, factory)
			require.Equal(t, [][]int{
				{1, 2, 3},
				{4},
			}, *groups)
		})
		t.Run("odd number 4 / 3", func(t *testing.T) {
			groups, factory := getGroups()
			DivideSize([]int{1, 2, 3, 4}, 3, factory)
			require.Equal(t, [][]int{
				{1, 2, 3}, {4},
			}, *groups)
		})
		t.Run("odd number 5 / 3", func(t *testing.T) {
			groups, factory := getGroups()
			DivideSize([]int{1, 2, 3, 4, 5}, 3, factory)
			require.Equal(t, [][]int{
				{1, 2, 3}, {4, 5},
			}, *groups)
		})
	})
	t.Run("by groups", func(t *testing.T) {
		t.Run("empty", func(t *testing.T) {
			_, factory := getGroups()
			got := DivideSize([]int{}, 5, factory)
			require.Empty(t, got)
		})
		t.Run("slice is smaller", func(t *testing.T) {
			groups, factory := getGroups()
			DivideSegments([]int{1, 2, 3, 4}, 5, factory)
			require.Equal(t, [][]int{
				{1}, {2}, {3}, {4},
			}, *groups)
		})
		t.Run("slice is bigger", func(t *testing.T) {
			groups, factory := getGroups()
			DivideSegments([]int{1, 2, 3, 4}, 2, factory)
			require.Equal(t, [][]int{
				{1, 2}, {3, 4},
			}, *groups)
		})
		t.Run("odd number 4 / 3", func(t *testing.T) {
			groups, factory := getGroups()
			DivideSegments([]int{1, 2, 3, 4}, 3, factory)
			require.Equal(t, [][]int{
				{1}, {2}, {3, 4},
			}, *groups)
		})
		t.Run("odd number 5 / 3", func(t *testing.T) {
			groups, factory := getGroups()
			DivideSegments([]int{1, 2, 3, 4, 5}, 3, factory)
			require.Equal(t, [][]int{
				{1}, {2, 3}, {4, 5},
			}, *groups)
		})
	})
}

func Test_Example01(t *testing.T) {
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
	pipego.Run(ctx,
		pipego.Parallel(3,
			DivideSize(data.values, 3, sumAgg)...,
		),
	)
	require.EqualValues(t, 45, sum)
}
