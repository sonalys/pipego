package pp_test

import (
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
					s.b = 1
					ready.Done() // If you set parallelism = 2 you will see this panics, because weight is 1.
					wg.Wait()
					return nil
				},
			)(ctx)
			require.NoError(t, err)
		})
		ready.Wait()
		require.NotEqual(t, s.b, s.a)
		ready.Add(1)
		wg.Done()
	})
}
