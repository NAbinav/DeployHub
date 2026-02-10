package db

import (
	"context"
	"fmt"
)

func GetDockerIds(ctx context.Context, pname string) ([]map[string]any, error) {
	query := `select image_id , branch from docker_image where pname=?`
	ids, err := ExecuteQueryMultiple(ctx, query, pname)
	fmt.Println(ids, err)
	return ids, err
}
