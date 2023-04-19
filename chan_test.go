package pp

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestChanDivide(t *testing.T) {
	ctx := NewContext()
	t.Run("empty", func(t *testing.T) {
		ch := make(chan int, 1)
		step := ChanDivide(ch, func(_ Context, i int) error {
			return fmt.Errorf("failed")
		})
		close(ch)
		err := step(ctx)
		require.NoError(t, err)
	})
	t.Run("context cancelled", func(t *testing.T) {
		ch := make(chan int, 1)
		ctx, cancel := ctx.WithCancel()
		value := 0
		step := ChanDivide(ch, func(_ Context, _ int) error {
			value = 1
			return fmt.Errorf("failed")
		})
		cancel()
		err := step(ctx)
		require.NoError(t, err)
		require.Equal(t, 0, value)
		close(ch)
	})
	t.Run("success", func(t *testing.T) {
		ch := make(chan int, 1)
		value := 0
		step := ChanDivide(ch, func(_ Context, i int) error {
			value = i
			return nil
		})
		ch <- 1
		close(ch)
		err := step(ctx)
		require.NoError(t, err)
		require.Equal(t, 1, value)
	})
}
