package pipego_test

import (
	"context"
	"testing"
	"time"

	"github.com/sonalys/pipego"
	"github.com/stretchr/testify/require"
)

func Test_Run(t *testing.T) {
	ctx := context.Background()
	t.Run("no steps", func(t *testing.T) {
		_, err := pipego.Run(ctx)
		require.NoError(t, err)
	})
	t.Run("with steps", func(t *testing.T) {
		run := false
		_, err := pipego.Run(ctx, func(ctx context.Context) (err error) {
			run = true
			return
		})
		require.NoError(t, err)
		require.True(t, run)
	})
	t.Run("with warnings", func(t *testing.T) {
		run := false
		report, err := pipego.Run(ctx, func(ctx context.Context) (err error) {
			run = true
			pipego.Warn(ctx, "warn")
			return
		})
		require.NoError(t, err)
		require.True(t, run)
		require.Len(t, report.Logs, 1)
	})
	t.Run("with duration", func(t *testing.T) {
		run := false
		delay := 100 * time.Millisecond
		report, err := pipego.Run(ctx, func(ctx context.Context) (err error) {
			run = true
			time.Sleep(delay)
			return
		})
		require.NoError(t, err)
		require.True(t, run)
		require.InDelta(t, delay, report.Duration, float64(delay)*0.1)
	})
	t.Run("keep step order", func(t *testing.T) {
		var i int
		_, err := pipego.Run(ctx,
			func(ctx context.Context) (err error) {
				require.Equal(t, 0, i)
				i++
				return
			},
			func(ctx context.Context) (err error) {
				require.Equal(t, 1, i)
				i++
				return
			},
		)
		require.NoError(t, err)
		require.Equal(t, 2, i)
	})
	t.Run("stop on error", func(t *testing.T) {
		_, err := pipego.Run(ctx,
			func(ctx context.Context) (err error) {
				return pipego.NilFieldError
			},
			func(ctx context.Context) (err error) {
				require.Fail(t, "should not run")
				return
			},
		)
		require.Equal(t, pipego.NilFieldError, err)
	})
}
