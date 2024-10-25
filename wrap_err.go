package pp

import "context"

type ErrorWrapper func(error) error

// WrapErr encapsulates all given steps errors, if an error is returned, it will be wrapped by ErrorWrapper's error.
func WrapErr(wrapper ErrorWrapper, steps ...Step) (out Steps) {
	out = make(Steps, 0, len(steps))
	for _, step := range steps {
		out = append(out, func(ctx context.Context) (err error) {
			err = step(ctx)
			if err != nil {
				return wrapper(err)
			}
			return nil
		})
	}
	return
}
