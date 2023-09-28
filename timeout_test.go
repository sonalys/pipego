package pp

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestTimeout(t *testing.T) {
	ctx := NewContext()
	tests := []struct {
		name string
		run  func(t *testing.T)
	}{
		{
			name: "empty",
			run: func(t *testing.T) {
				require.NotPanics(t, func() {
					resp := Timeout(0)
					require.Empty(t, resp)
				})
			},
		},
		{
			name: "dont timeout",
			run: func(t *testing.T) {
				a := 0
				f := func(_ Context) (err error) {
					a++
					return
				}
				steps := Timeout(time.Second,
					f, f, f,
				)
				_, err := New(steps...).Run(ctx)
				require.NoError(t, err)
				require.Equal(t, 3, a)
			},
		},
		{
			name: "timeout",
			run: func(t *testing.T) {
				a := 0
				f := func(ctx Context) (err error) {
					time.Sleep(400 * time.Millisecond)
					if ctx.Err() != nil {
						return
					}
					a++
					return
				}
				steps := Timeout(time.Second,
					f, f, f,
				)
				_, err := New(steps...).Run(ctx)
				require.Error(t, err)
				require.Equal(t, 2, a)
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.run(t)
		})
	}
}
