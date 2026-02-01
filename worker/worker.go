package worker

import (
	"context"
	"deployhub/db"
	"deployhub/deploy"
	"deployhub/schema"
	"deployhub/utils"
	"fmt"
	"time"
)

func Worker(ctx context.Context, params schema.DeployParams) (string, error) {
	var job schema.DeploymentJob
	job.ID = utils.GenerateRandomString(10)
	job.UserName = params.User
	job.ServiceName = params.ServiceName
	job.GitURL = params.GitURL
	job.Env = utils.ENVString(params.Env)
	job.Tag = params.GitURL
	job.Status = "queued"
	err := db.EnqueueJob(ctx, job)
	if err != nil {
		return "", err
	}
	env_string := utils.ENVString(params.Env)
	clean_url := fmt.Sprintf("https://www.github.com/%s/%s", params.User, params.GitURL)
	db.AddProject(ctx, params.User, clean_url, "", "", params.ServiceName, env_string)
	go func() {
		deployCtx, cancel := context.WithTimeout(ctx, 15*time.Minute)
		defer cancel()
		deploy.Deploy(deployCtx, params)
	}()
	return job.ID, nil
}
