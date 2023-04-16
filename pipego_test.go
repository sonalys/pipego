package pp_test

import (
	"context"
	"fmt"
	"testing"
	"time"

	pp "github.com/sonalys/pipego"
	"github.com/stretchr/testify/require"
)

func Test_Run(t *testing.T) {
	ctx := context.Background()
	t.Run("no steps", func(t *testing.T) {
		_, err := pp.Run(ctx)
		require.NoError(t, err)
	})
	t.Run("with steps", func(t *testing.T) {
		run := false
		_, err := pp.Run(ctx, func(_ pp.Context) (err error) {
			run = true
			return
		})
		require.NoError(t, err)
		require.True(t, run)
	})
	t.Run("with warnings", func(t *testing.T) {
		run := false
		r, err := pp.Run(ctx, func(ctx pp.Context) (err error) {
			run = true
			ctx.Warn("warn")
			return
		})
		require.NoError(t, err)
		require.True(t, run)
		require.Len(t, r.Logs(pp.ErrLevelWarn), 1)
	})
	t.Run("with duration", func(t *testing.T) {
		run := false
		delay := 100 * time.Millisecond
		r, err := pp.Run(ctx, func(_ pp.Context) (err error) {
			run = true
			time.Sleep(delay)
			return
		})
		require.NoError(t, err)
		require.True(t, run)
		require.InDelta(t, delay, r.Duration, float64(delay)*0.1)
	})
	t.Run("keep step order", func(t *testing.T) {
		var i int
		_, err := pp.Run(ctx,
			func(_ pp.Context) (err error) {
				require.Equal(t, 0, i)
				i++
				return err
			},
			func(_ pp.Context) (err error) {
				require.Equal(t, 1, i)
				i++
				return err
			},
		)
		require.NoError(t, err)
		require.Equal(t, 2, i)
	})
	t.Run("stop on error", func(t *testing.T) {
		_, err := pp.Run(ctx,
			func(_ pp.Context) (err error) {
				return fmt.Errorf("mock")
			},
			func(_ pp.Context) (err error) {
				require.Fail(t, "should not run")
				return
			},
		)
		require.Equal(t, fmt.Errorf("mock"), err)
	})
}
