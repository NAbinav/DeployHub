package db

import (
	"context"
	"fmt"
)

func UpdateProjectEnv(ctx context.Context, pname, env string) error {
	query := "UPDATE projects SET env = ? WHERE pname = ?;"
	_, err := ExecuteQuery(ctx, query, env, pname)
	if err != nil {
		return fmt.Errorf("failed to update env: %w", err)
	}
	return nil
}
