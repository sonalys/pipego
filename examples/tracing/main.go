package main

import (
	"context"
	"errors"
	"os"
	"time"

	pp "github.com/sonalys/pipego"
	"github.com/sonalys/pipego/retry"
	"golang.org/x/exp/slog"
)

func getLogger(ctx pp.Context) slog.Logger {
	return *slog.New(slog.NewJSONHandler(ctx.GetWriter(), nil)).With(slog.Attr{
		Key:   "section",
		Value: slog.AnyValue(ctx.GetPath()),
	})
}

func funcA(ctx pp.Context) (err error) {
	log := getLogger(ctx)
	log.Info("testing info")
	return
}

func funcB(ctx pp.Context) (err error) {
	ctx = ctx.SetSection("test section")
	log := getLogger(ctx)
	log.Error("from inside section")
	return
}

func funcC(ctx pp.Context) (err error) {
	log := getLogger(ctx)
	log.Error("from inside retry")
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
	//	[main.main.Parallel.func1] step=0
	//		[main.funcB] step=1
	//			[test section]
	//				{"level":"error","section":"test section","message":"from inside section"}
	//		[main.funcA] step=0
	//			{"level":"info","section":"main.funcA","message":"testing info"}
	//	[main.main.Constant.newRetry.func4] step=1
	//		[main.funcC] step=0
	//			{"level":"error","section":"main.funcC","message":"from inside retry"}
	//			{"level":"error","section":"main.funcC","message":"from inside retry"}
	//			{"level":"error","section":"main.funcC","message":"from inside retry"}
}
