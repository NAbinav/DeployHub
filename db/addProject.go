package db

import "context"

type Project struct {
	Project_name string `json:"project_name"`
	User         string `json:"user"`
	Git_url      string `json:"git_url"`
	Project_url  string `json:"project_url"`
	Framework    string `json:"framework"`
}

func AddProject(ctx context.Context, project_name string, user string, git_url string, project_url string, framework string) error {
	query := "insert into projects (pname,username,git_url,project_url,framework) values ($1,$2,$3,$4,$5)"
	_, err := DB.Exec(ctx, query, project_name, user, git_url, project_url, framework)
	return err
}

func GetUserProject(ctx context.Context, username string) ([]Project, error) {
	var AllProject []Project
	query := "select pname,username,git_url,project_url,framework  from projects where username=$1"
	rows, err := DB.Query(ctx, query, username)
	if err != nil {
		return []Project{}, err
	}
	defer rows.Close()
	for rows.Next() {
		var project Project
		if err := rows.Scan(&project.Project_name, &project.User, &project.Git_url, &project.Project_url, &project.Framework); err != nil {
			return []Project{}, err
		}
		AllProject = append(AllProject, project)
	}
	return AllProject, nil
}
