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
			func(ctx pp.Context) (err error) {
				ctx.Warn("parallel 1 warn")
				return
			},
			func(ctx pp.Context) (err error) {
				ctx.Info("parallel 2 info")
				return
			},
			retry.Constant(3, time.Second,
				func(ctx pp.Context) (err error) {
					return errors.New("error")
				},
			),
		),
	)
	if err != nil {
		println(err.Error())
	}
	fmt.Printf("finished in %s with %d warnings.\n\n", report.Duration, len(report.Logs(pp.Warn)))

	println("reconstructing log tree:")
	println(report.LogTree(pp.Trace))
}
