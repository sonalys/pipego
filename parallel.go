package pp

import (
	"context"

	"golang.org/x/sync/errgroup"
)

// Parallel runs all the given steps in parallel,
// It cancels context for the first non-nil error and returns.
// It runs 'n' go-routines at a time.
func Parallel(n uint16, steps ...Step) Step {
	return func(ctx context.Context) (err error) {
		if n <= 0 {
			n = uint16(len(steps))
		}

		errgrp, ctx := errgroup.WithContext(ctx)
		errgrp.SetLimit(int(n))

		for _, step := range steps {
			errgrp.Go(func() error {
				return step(ctx)
			})
		}

		return errgrp.Wait()
	}
}
