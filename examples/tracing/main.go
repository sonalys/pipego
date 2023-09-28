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

func funcA(ctx pp.Context) (err error) {
	log := getLogger(ctx)
	log.Info().Msg("testing info")
	return
}

func funcB(ctx pp.Context) (err error) {
	ctx = ctx.Section("test section")
	log := getLogger(ctx)
	log.Error().Msg("from inside section")
	return
}

func funcC(ctx pp.Context) (err error) {
	log := getLogger(ctx)
	log.Error().Msg("from inside retry")
	return errors.New("error")
}

func main() {
	ctx := context.Background()
	resp, _ := pp.New(
		pp.Parallel(2,
			funcA,
			funcB,
		),
		retry.Constant(3, time.Second,
			funcC,
		),
	).
		// WithSections allows us to use our pp.Context.Section function, segmentating logs by sections
		// Pipego is capable of using reflection to automatically segmentate functions by name.
		WithOptions(
			pp.WithAutomaticSections(),
			pp.WithSections(),
		).
		Run(ctx)
	// Note that the section log can also be customized by modifying pp.NewSectionFormatter.
	resp.LogTree(os.Stdout)
	// [root]
	// [main.main.Parallel.func1] step=0
	// 				[main.funcB] step=1
	// 								[test section]
	// 												{"level":"error","message":"from inside section"}
	// 				[main.funcA] step=0
	// 								{"level":"info","message":"testing info"}
	// [main.main.Constant.newRetry.func4] step=1
	// 				[main.funcC] step=0
	// 								{"level":"error","message":"from inside retry"}
	// 								{"level":"error","message":"from inside retry"}
	// 								{"level":"error","message":"from inside retry"}
}
