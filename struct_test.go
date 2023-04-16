package pp_test

import (
	"context"
	"testing"

	pp "github.com/sonalys/pipego"
	"github.com/stretchr/testify/require"
)

func Test_Struct(t *testing.T) {
	ctx := context.Background()
	t.Run("nil struct", func(t *testing.T) {
		err := pp.Struct(nil, func(ctx context.Context, t *int) error {
			return nil
		})(ctx)
		require.Equal(t, pp.NilStructErr, err)
	})
	t.Run("mutating one field", func(t *testing.T) {
		type str struct {
			a, b int
		}
		var got str
		err := pp.Struct(&got, func(ctx context.Context, t *str) error {
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
		err := pp.Struct(&got, func(ctx context.Context, t *str) error {
			t.a = 1
			return nil
		})(ctx)
		require.NoError(t, err)
		err = pp.Struct(&got, func(ctx context.Context, t *str) error {
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
