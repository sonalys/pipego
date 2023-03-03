package main

import (
	"context"
	"fmt"

	"github.com/sonalys/pipego"
)

// Object model from database.
type DatabaseObject struct{}

// fetchAPI usual api signature.
func fetchAPI(ctx context.Context, id string) (*DatabaseObject, error) {
	return &DatabaseObject{}, nil
}

// wrappedFetch adapts api to pipego signature.
func wrappedFetch(id string) pipego.FetchField[DatabaseObject] {
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
	report, err := pipego.Run(ctx,
		pipego.Field(&data.a1, wrappedFetch("a1")),
		pipego.Field(&data.a2, wrappedFetch("a2")),
		pipego.Field(&data.a3, wrappedFetch("a3")),
		func(ctx context.Context) (err error) {
			pipego.Warn(ctx, "warning: %s", data.a1)
			return
		},
	)
	if err != nil {
		println(err.Error())
		return
	}
	fmt.Printf("execution took %s.\n%d warnings:\n%s\n", report.Duration, len(report.Warnings), report.Warnings)
}
