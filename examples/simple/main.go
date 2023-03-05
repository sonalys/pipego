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
func wrappedFetch(id string) pp.FetchField[DatabaseObject] {
	return func(ctx context.Context) (*DatabaseObject, error) {
		return fetchAPI(ctx, id)
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
		pp.Field(&data.a1, wrappedFetch("a1")),
		pp.Field(&data.a2, wrappedFetch("a2")),
		pp.Field(&data.a3, wrappedFetch("a3")),
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
