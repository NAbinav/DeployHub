package db

import "context"

func AddDockerID(ctx context.Context, pname, image_id, branch string) error {
	query := `insert into docker_image (image_id,pname,branch) values (?,?,?)`
	_, err := ExecuteQuery(ctx, query, image_id, pname, branch)
	return err
}
