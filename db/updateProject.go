package db

import (
	"context"
	"fmt"
)

func UpdateProjectAfterDeploy(ctx context.Context, project_url, framework, pname, status string) error {
	query := `update projects set project_url= ? , framework=? ,status=? where pname=?`
	result, err := ExecuteQuery(ctx, query, project_url, framework, pname, status)
	fmt.Println(result)
	return err
}
