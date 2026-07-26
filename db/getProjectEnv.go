package db

import "context"

func GetProjectEnv(ctx context.Context, pname string) ([]map[string]any, error) {
	query := "SELECT env FROM projects WHERE pname = ?;"
	env, err := ExecuteQueryMultiple(ctx, query, pname)
	return env, err
}
