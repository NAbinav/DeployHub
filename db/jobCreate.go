package db

import (
	"context"
	"deployhub/schema"
	"fmt"
)

func EnqueueJob(ctx context.Context, job schema.DeploymentJob) error {
	query := `
        INSERT INTO deployment_jobs (
            id, user_name, service_name, git_url, env, tag, status
        ) VALUES (?, ?, ?, ?, ?, ?, ?)
    `

	result, err := ExecuteQuery(ctx, query,
		job.ID,
		job.UserName,
		job.ServiceName,
		job.GitURL,
		job.Env,
		job.Tag,
		"QUEUED",
	)
	fmt.Println(result)
	return err
}
