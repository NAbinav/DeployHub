package db

import (
	"context"
	"fmt"
)

func UpdateQueue(ctx context.Context, deployment_id, projectURL, framework, status string) error {
	query := `update deployment_jobs set service_url=?,framework=?,status=? where id=?;`
	result, err := ExecuteQuery(ctx, query, projectURL, framework, status, deployment_id)
	fmt.Println(result, err)
	return err
}
