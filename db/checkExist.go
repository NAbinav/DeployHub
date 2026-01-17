package db

import (
	"context"
	"fmt"
)

func ProjectExists(ctx context.Context, pname string) bool {
	query := `SELECT EXISTS(SELECT 1 FROM projects WHERE pname = ?) AS "exists";`
	p, err := ExecuteQuerySingle(ctx, query, pname)
	if err != nil {
		return false
	}

	existsValue, ok := p["exists"]
	if !ok {
		return false
	}

	return fmt.Sprint(existsValue) == "1"
}
