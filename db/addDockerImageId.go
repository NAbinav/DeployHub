package db

import "context"

func AddDockerID(ctx context.Context, pname, image_id string) error {
	query := `insert into docker_image (image_id,pname) values (?,?)`
	_, err := ExecuteQuery(ctx, query, pname, image_id)
	return err
}
