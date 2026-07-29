package main

import (
	"context"
	"deployhub/db"
	"deployhub/deploy"
	"deployhub/schema"
	"deployhub/utils"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/hibiken/asynq"
	"github.com/joho/godotenv"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)




func HandleWorker(ctx context.Context, t *asynq.Task) error {
	var params schema.DeployParams
	if err := json.Unmarshal(t.Payload(), &params); err != nil {
		return fmt.Errorf("json.Unmarshal failed: %v: %w", err, asynq.SkipRetry)
	}

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
		return err
	}
	env_string := utils.ENVString(params.Env)
	clean_url := fmt.Sprintf("https://www.github.com/%s/%s", params.User, params.GitURL)
	db.AddProject(ctx, params.User, clean_url, "", "", params.ServiceName, env_string)
	deployCtx, cancel := context.WithTimeout(ctx, 15*time.Minute)
	defer cancel()
	result := deploy.Deploy(deployCtx, params)
	fmt.Println(job.ID)
	db.UpdateQueue(ctx, job.ID, result.ServiceURL, result.Framework, result.Status)
	db.UpdateProjectAfterDeploy(ctx, result.ServiceURL, result.Framework, params.GitURL, result.Status)
	return nil
}

func main() {
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}

	err = db.Init_Cloudflare()
	if err != nil {
		fmt.Println(err)
		return
	}


	srv := asynq.NewServer(
		asynq.RedisClientOpt{Addr: os.Getenv("REDIS_URL")},
		asynq.Config{
			Concurrency: 10,
			Queues: map[string]int{
				"critical": 6,
				"default":  3,
				"low":      1,
			},
		},
	)
	mux := asynq.NewServeMux()
	mux.HandleFunc("deploy", HandleWorker)

	go func() {
		http.Handle("/metrics", promhttp.Handler())
		if err := http.ListenAndServe(":8081", nil); err != nil {
			log.Printf("metrics server error: %v", err)
		}
	}()

	if err := srv.Run(mux); err != nil {
		log.Fatalf("could not run server: %v", err)
	}
}
