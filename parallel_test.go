package pipego_test

import (
	"context"
	"sync"
	"testing"

	"github.com/sonalys/pipego"
	"github.com/stretchr/testify/require"
)

func Test_Parallel(t *testing.T) {
	ctx := context.Background()
	t.Run("empty", func(t *testing.T) {
		err := pipego.Parallel(5)(ctx)
		require.NoError(t, err)
	})
	t.Run("0 parallelism", func(t *testing.T) {
		err := pipego.Parallel(0)(ctx)
		require.Equal(t, pipego.ZeroParallelismErr, err)
	})
	t.Run("is parallel", func(t *testing.T) {
		var wg, ready sync.WaitGroup
		wg.Add(1)
		ready.Add(2)
		var a, b int
		go pipego.Parallel(2,
			func(ctx context.Context) (err error) {
				a = 1
				ready.Done()
				wg.Wait()
				return nil
			},
			func(ctx context.Context) (err error) {
				b = 2
				ready.Done()
				wg.Wait()
				return nil
			},
		)(ctx)
		ready.Wait()
		require.Equal(t, 1, a)
		require.Equal(t, 2, b)
		wg.Done()
	})
	t.Run("runs at the specified parallelism number", func(t *testing.T) {
		var wg, ready sync.WaitGroup
		wg.Add(1)
		ready.Add(1)
		var a, b int
		go require.NotPanics(t, func() {
			err := pipego.Parallel(1,
				func(ctx context.Context) (err error) {
					a = 1
					ready.Done()
					wg.Wait()
					return nil
				},
				func(ctx context.Context) (err error) {
					b = 2
					ready.Done() // If you set parallelism = 2 you will see this panics, because weight is 1.
					wg.Wait()
					return nil
				},
			)(ctx)
			require.NoError(t, err)
		})
		ready.Wait()
		require.Equal(t, 1, a)
		require.Equal(t, 0, b)
		ready.Add(1)
		wg.Done()
	})
	t.Run("context is cancelled when step errors", func(t *testing.T) {
		var ready sync.WaitGroup
		ready.Add(1)
		err := pipego.Parallel(1,
			func(ctx context.Context) (err error) {
				defer ready.Done()
				return pipego.NilFieldError
			},
			func(ctx context.Context) (err error) {
				ready.Wait()
				require.Error(t, ctx.Err())
				return nil
			},
		)(ctx)
		require.Equal(t, pipego.NilFieldError, err)
	})
}
