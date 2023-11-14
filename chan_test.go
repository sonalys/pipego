package pp

import (
	"context"
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestChanDivide(t *testing.T) {
	ctx := context.Background()
	// Since we are obliged to work with channel pointers in this context, we need this tricky function to get
	// the casting right.
	getCh := func() (chan int, *<-chan int) {
		ch := make(chan int, 1)
		var recv <-chan int
		recv = ch
		return ch, &recv
	}
	t.Run("empty", func(t *testing.T) {
		ch, recv := getCh()
		step := ChanDivide(recv, func(_ context.Context, i int) error {
			return fmt.Errorf("failed")
		})
		close(ch)
		err := step(ctx)

		require.NoError(t, err)
	})
	t.Run("context cancelled", func(t *testing.T) {
		ch, recv := getCh()
		ctx, cancel := context.WithCancel(ctx)
		value := 0
		step := ChanDivide(recv, func(_ context.Context, _ int) error {
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
		ch, recv := getCh()
		value := 0
		step := ChanDivide(recv, func(_ context.Context, i int) error {
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
