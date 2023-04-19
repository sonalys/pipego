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
				ctx.Info("parallel 0 warn")
				return
			},
			func(ctx pp.Context) (err error) {
				ctx.Info("parallel 1 info")
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
	fmt.Printf("finished in %s with %d warnings.\n\n", report.Duration, len(report.Logs(pp.Info)))

	println("reconstructing log tree:")
	println(report.LogTree(pp.Trace))
	// 	error
	// finished in 3.001202993s with 4 warnings.
	//
	// reconstructing log tree:
	// [root] new context initialized
	//       2023-04-19T16:08:26+02:00: starting Run method with 1 steps
	//       2023-04-19T16:08:26+02:00: running step[0]
	//       [parallel] n = 2 steps = 3
	//               2023-04-19T16:08:26+02:00: waiting tasks to finish
	//               [step-2]
	//                       2023-04-19T16:08:26+02:00: queued
	//                       2023-04-19T16:08:26+02:00: running
	//                       [retry] n = 3 r = retry.constantRetry
	//                               2023-04-19T16:08:26+02:00: retry failed #1: error
	//                               2023-04-19T16:08:27+02:00: retry failed #2: error
	//                               2023-04-19T16:08:28+02:00: retry failed #3: error
	//               [step-0]
	//                       2023-04-19T16:08:26+02:00: queued
	//                       2023-04-19T16:08:26+02:00: running
	//                       2023-04-19T16:08:26+02:00: parallel 0 warn
	//                       2023-04-19T16:08:26+02:00: finished
	//               [step-1]
	//                       2023-04-19T16:08:26+02:00: queued
	//                       2023-04-19T16:08:26+02:00: running
	//                       2023-04-19T16:08:26+02:00: parallel 1 info
	//                       2023-04-19T16:08:26+02:00: finished
	//               2023-04-19T16:08:29+02:00: closing errChan
	//               2023-04-19T16:08:29+02:00: parallel method finished in 3.00027916s
	//       2023-04-19T16:08:29+02:00: Run method finished in 3.00030725s
}
