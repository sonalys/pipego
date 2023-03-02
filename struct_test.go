package pipego_test

import (
	"context"
	"testing"

	"github.com/sonalys/pipego"
	"github.com/stretchr/testify/require"
)

func Test_Struct(t *testing.T) {
	ctx := context.Background()
	t.Run("nil struct", func(t *testing.T) {
		err := pipego.Struct(nil, func(ctx context.Context, t *int) error {
			return nil
		})(ctx)
		require.Equal(t, pipego.NilStructErr, err)
	})
	t.Run("mutating one field", func(t *testing.T) {
		type str struct {
			a, b int
		}
		var got str
		err := pipego.Struct(&got, func(ctx context.Context, t *str) error {
			t.a = 1
			return nil
		})(ctx)
		require.NoError(t, err)
		require.Equal(t, str{
			a: 1,
		}, got)
	})
	t.Run("mutating two fields", func(t *testing.T) {
		type str struct {
			a, b int
		}
		var got str
		err := pipego.Struct(&got, func(ctx context.Context, t *str) error {
			t.a = 1
			return nil
		})(ctx)
		require.NoError(t, err)
		err = pipego.Struct(&got, func(ctx context.Context, t *str) error {
			t.b = 2
			return nil
		})(ctx)
		require.NoError(t, err)
		require.Equal(t, str{
			a: 1,
			b: 2,
		}, got)
	})
}
