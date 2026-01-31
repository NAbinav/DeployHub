package worker

import (
	"context"
	"deployhub/db"
	"deployhub/deploy"
	"deployhub/schema"
	"deployhub/utils"
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
	go func() {
		deployCtx, _ := context.WithTimeout(ctx, 15*time.Minute)
		deploy.Deploy(deployCtx, params)
	}()
	return job.ID, nil

}
