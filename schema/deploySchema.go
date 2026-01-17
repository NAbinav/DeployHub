package schema

import (
	"golang.org/x/oauth2"
)

type DeployRequest struct {
	GitURL      string            `json:"git_url"`
	ServiceName string            `json:"name"`
	Env         map[string]string `json:"env"`
}

type DeployResponse struct {
	URL          string `json:"url"`
	DeploymentID string `json:"deployment_id"`
	Error        string `json:"error,omitempty"`
}

type DeployParams struct {
	User        string
	AccessToken oauth2.Token
	Tag         string
	GitURL      string
	ServiceName string
	Env         map[string]string
}

type DeployResult struct {
	ServiceURL   string
	DeploymentID string
	Framework    string
	Error        error
}

type DeploymentConfig struct {
	ProjectID    string
	Region       string
	RepoID       string
	ImageName    string
	ImagePath    string
	User         string
	AccessToken  oauth2.Token
	GitURL       string
	GitCleanURL  string
	DeploymentID string
}
