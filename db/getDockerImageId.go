package db

import "context"

func GetDockerIds(ctx context.Context, pname string) ([]map[string]any, error) {
	query := `select image_id from docker_images where pname=?`
	ids, err := ExecuteQueryMultiple(ctx, query, pname)
	return ids, err
}
