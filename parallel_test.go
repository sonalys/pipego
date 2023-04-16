package pp_test

import (
	"fmt"
	"sync"
	"testing"

	pp "github.com/sonalys/pipego"
	"github.com/stretchr/testify/require"
)

func Test_Parallel(t *testing.T) {
	ctx := pp.NewContext()
	t.Run("empty", func(t *testing.T) {
		err := pp.Parallel(5)(ctx)
		require.NoError(t, err)
	})
	t.Run("0 parallelism", func(t *testing.T) {
		err := pp.Parallel(0)(ctx)
		require.Equal(t, pp.ZeroParallelismErr, err)
	})
	t.Run("is parallel", func(t *testing.T) {
		var wg, ready sync.WaitGroup
		wg.Add(1)
		ready.Add(2)
		type state struct {
			a, b int
		}
		var s state
		go pp.Parallel(2,
			func(_ pp.Context) (err error) {
				s.a = 1
				ready.Done()
				wg.Wait()
				return nil
			},
			func(_ pp.Context) (err error) {
				s.b = 2
				ready.Done()
				wg.Wait()
				return nil
			},
		)(ctx)
		ready.Wait()
		require.Equal(t, 1, s.a)
		require.Equal(t, 2, s.b)
		wg.Done()
	})
	t.Run("runs at the specified parallelism number", func(t *testing.T) {
		var wg, ready sync.WaitGroup
		wg.Add(1)
		ready.Add(1)
		type state struct {
			a, b int
		}
		var s state
		go require.NotPanics(t, func() {
			err := pp.Parallel(1,
				func(_ pp.Context) (err error) {
					s.a = 1
					ready.Done()
					wg.Wait()
					return nil
				},
				func(_ pp.Context) (err error) {
					s.b = 2
					ready.Done() // If you set parallelism = 2 you will see this panics, because weight is 1.
					wg.Wait()
					return nil
				},
			)(ctx)
			require.NoError(t, err)
		})
		ready.Wait()
		require.Equal(t, 1, s.a)
		require.Equal(t, 0, s.b)
		ready.Add(1)
		wg.Done()
	})
	t.Run("context is cancelled when step errors", func(t *testing.T) {
		var ready sync.WaitGroup
		ready.Add(1)
		err := pp.Parallel(1,
			func(_ pp.Context) (err error) {
				defer ready.Done()
				return fmt.Errorf("mock")
			},
			func(_ pp.Context) (err error) {
				ready.Wait()
				require.Error(t, ctx.Err())
				return nil
			},
		)(ctx)
		require.Equal(t, fmt.Errorf("mock"), err)
	})
}
