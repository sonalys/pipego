package main

import (
	"context"
	"fmt"

	pp "github.com/sonalys/pipego"
)

// Object model from database.
type DatabaseObject struct{}

// fetchAPI usual api signature.
func fetchAPI(ctx context.Context, id string) (*DatabaseObject, error) {
	return &DatabaseObject{}, nil
}

// wrappedFetch adapts api to pipego signature.
func wrappedFetch(id string, ptr *DatabaseObject) pp.StepFunc {
	return func(ctx context.Context) (err error) {
		ptr, err = fetchAPI(ctx, id)
		return
	}
}

func main() {
	var data struct {
		a1 DatabaseObject
		a2 DatabaseObject
		a3 DatabaseObject
	}
	ctx := context.Background()
	report, err := pp.Run(ctx,
		wrappedFetch("a1", &data.a1),
		wrappedFetch("a2", &data.a2),
		wrappedFetch("a3", &data.a3),
		func(ctx context.Context) (err error) {
			pp.Warn(ctx, "warning: %s", data.a1)
			return
		},
	)
	if err != nil {
		println(err.Error())
		return
	}
	logs := report.Logs(pp.ErrLevelWarn)
	fmt.Printf("execution took %s.\n%d warnings:\n%s\n", report.Duration, len(logs), logs)
}
