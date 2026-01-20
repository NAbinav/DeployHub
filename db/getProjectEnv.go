package db

import "context"

func getProjectEnv(ctx context.Context, pname string) ([]map[string]any, error) {
	query := "SELECT username, git_url, project_url, framework FROM projects WHERE pname = ?;"
	env, err := ExecuteQueryMultiple(ctx, query, pname)
	return env, err
}
