package pp

import (
	"context"
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestWrapErr(t *testing.T) {
	ctx := context.Background()
	type args struct {
		wrapper ErrorWrapper
		steps   Steps
	}
	tests := []struct {
		name string
		args args
		exp  error
	}{
		{
			name: "simple case",
			exp:  fmt.Errorf("mock"),
			args: args{
				wrapper: func(err error) error {
					return fmt.Errorf("mock")
				},
				steps: Steps{
					func(ctx context.Context) (err error) {
						return fmt.Errorf("test")
					},
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotOut := WrapErr(tt.args.wrapper, tt.args.steps...)[0](ctx)
			require.Equal(t, tt.exp, gotOut)
		})
	}
}
