package pp

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestWrapErr(t *testing.T) {
	ctx := NewContext()
	type args struct {
		wrapper ErrorWrapper
		steps   []StepFunc
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
				steps: []StepFunc{
					func(ctx Context) (err error) {
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
