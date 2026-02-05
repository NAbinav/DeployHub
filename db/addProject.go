package db

import (
	"context"
	"fmt"
)

type Project struct {
	Project_name string `json:"project_name"`
	User         string `json:"user"`
	Git_url      string `json:"git_url"`
	Project_url  string `json:"project_url"`
	Framework    string `json:"framework"`
	ENV          string `json:"env"`
}

func AddProject(ctx context.Context, username, gitURL, projectURL, framework, projectName, env string) error {
	query := `INSERT INTO projects (username, git_url, project_url, framework, pname, status,env)
VALUES (?, ?, ?, ?, ?, "deploying...",?);`
	_, err := ExecuteQuery(ctx, query, username, gitURL, projectURL, framework, projectName, env)
	if err != nil {
		return fmt.Errorf("failed to save project: %w", err)
	}

	return nil

}

func GetUserProject(ctx context.Context, username string) ([]map[string]any, error) {
	query := `
		SELECT username, git_url, project_url, framework,pname,status
		FROM projects
		WHERE username = ?
	`
	projects, err := ExecuteQueryMultiple(ctx, query, username)
	if err != nil {
		return nil, fmt.Errorf("failed to get projects: %w", err)
	}
	return projects, nil
}
