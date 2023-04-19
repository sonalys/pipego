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
	// 	error
	// finished in 3.001202993s with 4 warnings.
	//
	// reconstructing log tree:
	// [6568b9a8-b050-4e40-8285-6f3b32fd04c9] root
	// [4ef26c3a-d32a-4615-92a1-c3a25c5cb0e8] run
	//         2023-04-19T14:33:06+02:00: starting Run method with 1 steps
	//         2023-04-19T14:33:06+02:00: running step[0]
	//         [b531dfad-c47f-4773-a53a-aad2372edc74] parallel
	//                 2023-04-19T14:33:06+02:00: starting parallelism = 2 with 3 steps
	//                 2023-04-19T14:33:06+02:00: step[0] is queued
	//                 2023-04-19T14:33:06+02:00: step[0] is running
	//                 2023-04-19T14:33:06+02:00: step[1] is queued
	//                 2023-04-19T14:33:06+02:00: step[1] is running
	//                 2023-04-19T14:33:06+02:00: step[2] is queued
	//                 2023-04-19T14:33:06+02:00: parallel 2 info
	//                 2023-04-19T14:33:06+02:00: step[1] is finished
	//                 2023-04-19T14:33:06+02:00: step[2] is running
	//                 2023-04-19T14:33:06+02:00: waiting tasks to finish
	//                 2023-04-19T14:33:06+02:00: parallel 1 warn
	//                 2023-04-19T14:33:09+02:00: step[0] errored: error. finishing execution
	//                 2023-04-19T14:33:09+02:00: Run method finished in 3.001202993s
	//                 [20be3fde-a176-40de-808f-0c95d6082cef] retry
	//                         2023-04-19T14:33:06+02:00: retry failed #1: error
	//                         2023-04-19T14:33:06+02:00: step[0] is finished
	//                         2023-04-19T14:33:07+02:00: retry failed #2: error
	//                         2023-04-19T14:33:08+02:00: retry failed #3: error
	//                         2023-04-19T14:33:09+02:00: step[2] failed: error. finishing execution
	//                         2023-04-19T14:33:09+02:00: step[2] is finished
	//                         2023-04-19T14:33:09+02:00: closing errChan
	//                         2023-04-19T14:33:09+02:00: Parallel method finished in 3.001174873s
}
