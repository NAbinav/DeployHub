package db

import (
	"context"
	"fmt"
	"os"

	"github.com/cloudflare/cloudflare-go/v3"
	"github.com/cloudflare/cloudflare-go/v3/d1"
)

func ExecuteQuery(ctx context.Context, query string, params ...string) (*[]d1.QueryResult, error) {
	resp, err := Client.D1.Database.Query(
		ctx,
		os.Getenv("CLOUDFLARE_D1_DB"),
		d1.DatabaseQueryParams{
			AccountID: cloudflare.F(os.Getenv("CLOUDFLARE_ACCOUNT_ID")),
			Sql:       cloudflare.F(query),
			Params:    cloudflare.F(params),
		},
	)

	if err != nil {
		return nil, fmt.Errorf("query execution failed: %w", err)
	}

	return resp, nil
}

func ExecuteQuerySingle(ctx context.Context, query string, params ...string) (map[string]any, error) {
	resp, err := ExecuteQuery(ctx, query, params...)
	if err != nil {
		return nil, err
	}

	if resp == nil {
		return nil, fmt.Errorf("no response from database")
	}

	for _, page := range *resp {
		if len(page.Results) == 0 {
			return nil, fmt.Errorf("no results found")
		}
		return page.Results[0].(map[string]any), nil
	}

	return nil, fmt.Errorf("no result returned")
}

func ExecuteQueryMultiple(ctx context.Context, query string, params ...string) ([]map[string]any, error) {
	resp, err := ExecuteQuery(ctx, query, params...)
	if err != nil {
		return nil, err
	}

	if resp == nil {
		return nil, fmt.Errorf("no response from database")
	}

	var results []map[string]any
	for _, page := range *resp {
		for _, row := range page.Results {
			results = append(results, row.(map[string]any))
		}
	}

	return results, nil
}
