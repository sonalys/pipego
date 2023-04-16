package pp

// func TestDivideSliceInSize(t *testing.T) {
// 	getGroups := func() (*[][]int, func(slice []int) StepFunc) {
// 		group := [][]int{}
// 		return &group, func(slice []int) StepFunc {
// 			group = append(group, slice)
// 			return nil
// 		}
// 	}
// 	t.Run("by size", func(t *testing.T) {
// 		t.Run("empty", func(t *testing.T) {
// 			_, factory := getGroups()
// 			got := DivideSliceInSize([]int{}, 5, factory)
// 			require.Empty(t, got)
// 		})
// 		t.Run("slice is smaller", func(t *testing.T) {
// 			groups, factory := getGroups()
// 			_ = DivideSliceInSize([]int{1, 2, 3, 4}, 5, factory)
// 			require.Equal(t, [][]int{
// 				{1, 2, 3, 4},
// 			}, *groups)
// 		})
// 		t.Run("slice is bigger", func(t *testing.T) {
// 			_, factory := getGroups()
// 			got := DivideSliceInSize([]int{1, 2, 3, 4}, 2, factory)
// 			require.Len(t, got, 2)
// 		})
// 		t.Run("odd length", func(t *testing.T) {
// 			groups, factory := getGroups()
// 			_ = DivideSliceInSize([]int{1, 2, 3, 4}, 3, factory)
// 			require.Equal(t, [][]int{
// 				{1, 2, 3},
// 				{4},
// 			}, *groups)
// 		})
// 		t.Run("odd number 4 / 3", func(t *testing.T) {
// 			groups, factory := getGroups()
// 			DivideSliceInSize([]int{1, 2, 3, 4}, 3, factory)
// 			require.Equal(t, [][]int{
// 				{1, 2, 3}, {4},
// 			}, *groups)
// 		})
// 		t.Run("odd number 5 / 3", func(t *testing.T) {
// 			groups, factory := getGroups()
// 			DivideSliceInSize([]int{1, 2, 3, 4, 5}, 3, factory)
// 			require.Equal(t, [][]int{
// 				{1, 2, 3}, {4, 5},
// 			}, *groups)
// 		})
// 	})
// 	t.Run("by groups", func(t *testing.T) {
// 		t.Run("empty", func(t *testing.T) {
// 			_, factory := getGroups()
// 			got := DivideSliceInSize([]int{}, 5, factory)
// 			require.Empty(t, got)
// 		})
// 		t.Run("slice is smaller", func(t *testing.T) {
// 			groups, factory := getGroups()
// 			DivideSliceInGroups([]int{1, 2, 3, 4}, 5, factory)
// 			require.Equal(t, [][]int{
// 				{1}, {2}, {3}, {4},
// 			}, *groups)
// 		})
// 		t.Run("slice is bigger", func(t *testing.T) {
// 			groups, factory := getGroups()
// 			DivideSliceInGroups([]int{1, 2, 3, 4}, 2, factory)
// 			require.Equal(t, [][]int{
// 				{1, 2}, {3, 4},
// 			}, *groups)
// 		})
// 		t.Run("odd number 4 / 3", func(t *testing.T) {
// 			groups, factory := getGroups()
// 			DivideSliceInGroups([]int{1, 2, 3, 4}, 3, factory)
// 			require.Equal(t, [][]int{
// 				{1}, {2}, {3, 4},
// 			}, *groups)
// 		})
// 		t.Run("odd number 5 / 3", func(t *testing.T) {
// 			groups, factory := getGroups()
// 			DivideSliceInGroups([]int{1, 2, 3, 4, 5}, 3, factory)
// 			require.Equal(t, [][]int{
// 				{1}, {2, 3}, {4, 5},
// 			}, *groups)
// 		})
// 	})
// }

// func Test_Example01(t *testing.T) {
// 	var data struct {
// 		values []int
// 	}
// 	data.values = []int{1, 2, 3, 4, 5, 6, 7, 8, 9}
// 	var sum int64
// 	sumAgg := func(v int) StepFunc {
// 		return func(ctx context.Context) error {
// 			atomic.AddInt64(&sum, int64(v))
// 			return nil
// 		}
// 	}
// 	ctx := context.Background()
// 	Run(ctx,
// 		Parallel(3,
// 			DivideSliceInSize(data.values, 3, sumAgg)...,
// 		),
// 	)
// 	require.EqualValues(t, 45, sum)
// }
