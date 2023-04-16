package main

import (
	"context"
	"errors"
	"fmt"
	"time"

	pp "github.com/sonalys/pipego"
	"github.com/sonalys/pipego/retry"
)

func main() {
	ctx := context.Background()
	// Do not print any logs in stdOut.
	pp.DefaultLogger = nil
	report, err := pp.Run(ctx,
		pp.Parallel(2,
			func(ctx context.Context) (err error) {
				pp.Warn(ctx, "parallel 1 warn")
				return
			},
			func(ctx context.Context) (err error) {
				pp.Warn(ctx, "parallel 2 warn")
				return
			},
			retry.Retry(3, retry.ConstantRetry(time.Second),
				func(ctx context.Context) (err error) {
					return errors.New("error")
				},
			),
		),
	)
	if err != nil {
		println(err.Error())
	}
	fmt.Printf("finished in %s with %d warnings.\n\n", report.Duration, len(report.Logs(pp.ErrLevelWarn)))

	println("reconstructing log tree:")
	println(report.LogTree(pp.ErrLevelTrace))
}
