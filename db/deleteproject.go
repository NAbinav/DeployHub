package db

import (
	"context"
	"fmt"
)

func UpdateProjectName(ctx context.Context, oldPname, newPname string) error {

	getQuery := "SELECT username, git_url, project_url, framework FROM projects WHERE pname = ?;"
	row, err := ExecuteQuerySingle(ctx, getQuery, oldPname)
	if err != nil {
		return fmt.Errorf("project not found: %w", err)
	}

	username := row["username"].(string)
	gitUrl := row["git_url"].(string)
	projectUrl, _ := row["project_url"].(string)
	framework, _ := row["framework"].(string)

	deleteQuery := "DELETE FROM projects WHERE pname = ?;"
	_, err = ExecuteQuery(ctx, deleteQuery, oldPname)
	if err != nil {
		return fmt.Errorf("failed to delete old project: %w", err)
	}

	insertQuery := `
		INSERT INTO projects (pname, username, git_url, project_url, framework)
		VALUES (?, ?, ?, ?, ?);
	`
	_, err = ExecuteQuery(ctx, insertQuery, newPname, username, gitUrl, projectUrl, framework)
	if err != nil {
		return fmt.Errorf("failed to insert project with new name: %w", err)
	}

	return nil
}
