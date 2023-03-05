package main

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/sonalys/pipego"
	"github.com/sonalys/pipego/retry"
)

func main() {
	ctx := context.Background()
	// Do not print any logs in stdOut.
	pipego.DefaultLogger = nil
	report, err := pipego.Run(ctx,
		pipego.Parallel(2,
			func(ctx context.Context) (err error) {
				pipego.Warn(ctx, "parallel 1 warn")
				return
			},
			func(ctx context.Context) (err error) {
				pipego.Warn(ctx, "parallel 2 warn")
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
	fmt.Printf("finished in %s with %d warnings.\n\n", report.Duration, len(report.Logs(pipego.ErrLevelWarn)))

	println("reconstructing log tree:")
	println(report.LogTree(pipego.ErrLevelTrace))
}
