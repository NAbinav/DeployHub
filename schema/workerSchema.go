package schema

type DeploymentJob struct {
	ID string `json:"id"`

	UserName    string `json:"user_name"`
	ServiceName string `json:"service_name"`
	GitURL      string `json:"git_url"`

	Env string `json:"env"`
	Tag string `json:"tag"`

	Status string `json:"status"`

	ServiceURL *string `json:"service_url,omitempty"`
	Framework  *string `json:"framework,omitempty"`
	Error      *string `json:"error,omitempty"`

	Attempts int `json:"attempts"`

	CreatedAt  string  `json:"created_at"`
	StartedAt  *string `json:"started_at,omitempty"`
	FinishedAt *string `json:"finished_at,omitempty"`
}
