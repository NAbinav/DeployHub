package db

import (
	"context"
	"fmt"
)

func ProjectExists(ctx context.Context, pname string) bool {
	query := `SELECT EXISTS(SELECT 1 FROM projects WHERE pname = ?);`
	p, err := ExecuteQuerySingle(ctx, query, pname)
	if err != nil {
		return true
	}
	fmt.Println(p[query])
	return p[query] == 1
}
