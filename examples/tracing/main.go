package main

import (
	"context"
	"errors"
	"os"
	"time"

	"github.com/rs/zerolog"
	pp "github.com/sonalys/pipego"
	"github.com/sonalys/pipego/retry"
)

func getLogger(ctx pp.Context) zerolog.Logger {
	return zerolog.New(ctx.GetWriter())
}

func main() {
	ctx := context.Background()
	r, _ := pp.Run(ctx,
		pp.Parallel(2,
			func(ctx pp.Context) (err error) {
				log := getLogger(ctx)
				log.Info().Msg("testing info")
				return
			},
			func(ctx pp.Context) (err error) {
				ctx = ctx.Section("test section")
				log := getLogger(ctx)
				log.Error().Msg("from inside section")
				return
			},
			retry.Constant(3, time.Second,
				func(ctx pp.Context) (err error) {
					log := getLogger(ctx)
					log.Error().Msg("from inside retry")
					return errors.New("error")
				},
			),
		),
	)
	// Note that the section log can also be customized by modifying pp.NewSectionFormatter.
	r.LogNode.Tree(os.Stdout)
	// [root] new context initialized
	//         [parallel] n = 2 steps = 3
	//                 [step-2]
	//                         [retry] n = 3 r = retry.constantRetry
	//                                 {"level":"error","message":"from inside retry"}
	//                                 {"level":"error","message":"from inside retry"}
	//                                 {"level":"error","message":"from inside retry"}
	//                 [step-0]
	//                 [step-1]
	//                         [test section]
	//                                 {"level":"error","message":"from inside section"}
}
