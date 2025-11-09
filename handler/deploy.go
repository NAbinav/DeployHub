package handler

import (
	"context"
	"deployhub/db"
	// "deployhub/helper"
	"deployhub/jwt"
	"deployhub/log"
	"deployhub/utils"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"cloud.google.com/go/artifactregistry/apiv1"
	artifactregistrypb "cloud.google.com/go/artifactregistry/apiv1/artifactregistrypb"
	run "cloud.google.com/go/run/apiv2"
	runpb "cloud.google.com/go/run/apiv2/runpb"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"golang.org/x/oauth2"
	"google.golang.org/api/option"
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

func DeployHandler(c *gin.Context) {
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Minute)
	defer cancel()
	var req DeployRequest
	if err := c.BindJSON(&req); err != nil {
		c.Error(err)
		return
	}

	// Generate unique deployment ID
	deploymentID := uuid.New().String()

	// Set up SSE stream
	c.Writer.Header().Set("Content-Type", "text/event-stream")
	c.Writer.Header().Set("Cache-Control", "no-cache")
	c.Writer.Header().Set("Connection", "keep-alive")
	c.SSEvent("message", "Deployment started for "+req.ServiceName)

	logStream := make(chan string, 100) // Buffered channel to avoid blocking
	defer close(logStream)

	logWriter := &log.LogWriter{
		DeploymentID: deploymentID,
		User:         "",
		Ctx:          ctx,
		Stream:       logStream,
	}

	go func() {
		for log := range logStream {
			c.SSEvent("message", log)
			c.Writer.Flush()
		}
	}()

	projectID := os.Getenv("GCLOUD_PROJECT_ID")
	if projectID == "" {
		logStream <- "GCLOUD_PROJECT_ID not set"
		c.Error(errors.New("GCLOUD_PROJECT_ID not set"))
		return
	}

	region := "asia-south1"
	repoID := "deploy-hub"
	imageName := fmt.Sprintf("%s-image", req.ServiceName)
	imagePath := fmt.Sprintf("%s-docker.pkg.dev/%s/%s/%s:latest", region, projectID, repoID, imageName)

	token, err := c.Cookie("token")
	if err != nil {
		logStream <- fmt.Sprintf("Invalid token: %v", err)
		c.Error(err)
		return
	}

	user, err := jwt.Verify_JWT(token)
	if err != nil {
		logStream <- fmt.Sprintf("JWT verification failed: %v", err)
		c.Error(err)
		return
	}
	logWriter.User = user // Set user for LogWriter
	var access_token oauth2.Token
	access_token, err = db.UserToken(c, user)
	if err != nil {
		logStream <- fmt.Sprintf("Failed to get user token: %v", err)
		c.Error(err)
		return
	}

	logStream <- "Starting deployment for " + req.ServiceName

	gitURL := "https://" + access_token.AccessToken + "@github.com/" + user + "/" + req.GitURL
	gitCleanURL := "https://www.github.com/" + user + "/" + req.GitURL

	tempDir := filepath.Join("/tmp", fmt.Sprintf("repo-%d", time.Now().UnixNano()))
	if err := utils.RunCommandWithOutput("git", []string{"clone", gitURL, tempDir}, logWriter); err != nil {
		logStream <- fmt.Sprintf("Git clone failed: %v", err)
		c.Error(err)
		return
	}
	defer os.RemoveAll(tempDir) // Clean up

	framework := utils.DetectFramework(tempDir)
	logStream <- "Detected framework: " + framework

	if err := utils.CreateDockerfile(tempDir, framework); err != nil {
		logStream <- fmt.Sprintf("Dockerfile creation failed: %v", err)
		c.Error(err)
		return
	}

	arClient, err := artifactregistry.NewClient(ctx, option.WithCredentialsFile(os.Getenv("GOOGLE_APPLICATION_CREDENTIALS")))
	if err != nil {
		logStream <- fmt.Sprintf("artifactregistry.NewClient: %v", err)
		c.Error(fmt.Errorf("artifactregistry.NewClient: %v", err))
		return
	}
	defer arClient.Close()

	repoParent := fmt.Sprintf("projects/%s/locations/%s", projectID, region)
	repoName := fmt.Sprintf("%s/repositories/%s", repoParent, repoID)

	_, err = arClient.GetRepository(ctx, &artifactregistrypb.GetRepositoryRequest{Name: repoName})
	if err != nil {
		logStream <- "Repository not found — creating: " + repoID
		_, err = arClient.CreateRepository(ctx, &artifactregistrypb.CreateRepositoryRequest{
			Parent:       repoParent,
			RepositoryId: repoID,
			Repository: &artifactregistrypb.Repository{
				Format:      artifactregistrypb.Repository_DOCKER,
				Description: "DeployHub managed Docker repo",
			},
		})
		if err != nil {
			logStream <- fmt.Sprintf("CreateRepository: %v", err)
			c.Error(fmt.Errorf("CreateRepository: %v", err))
			return
		}
	}

	logStream <- "Configuring Docker authentication..."
	if err := utils.RunCommandWithOutput("gcloud", []string{"auth", "configure-docker", fmt.Sprintf("%s-docker.pkg.dev", region), "--quiet"}, logWriter); err != nil {
		logStream <- fmt.Sprintf("Docker auth configuration failed: %v", err)
		c.Error(fmt.Errorf("Docker auth configuration failed: %v", err))
		return
	}

	logStream <- "Building Docker image..."
	if err := utils.RunCommandWithOutput("docker", []string{"build", "-t", imagePath, tempDir}, logWriter); err != nil {
		logStream <- fmt.Sprintf("Docker build failed: %v", err)
		c.Error(fmt.Errorf("Docker build failed: %v", err))
		return
	}

	logStream <- "Pushing image to Artifact Registry..."
	if err := utils.RunCommandWithOutput("docker", []string{"push", imagePath}, logWriter); err != nil {
		logStream <- fmt.Sprintf("Docker push failed: %v", err)
		c.Error(fmt.Errorf("Docker push failed: %v", err))
		return
	}

	logStream <- "Image pushed successfully!"

	runClient, err := run.NewServicesClient(ctx, option.WithCredentialsFile(os.Getenv("GOOGLE_APPLICATION_CREDENTIALS")))
	if err != nil {
		logStream <- fmt.Sprintf("run.NewServicesClient: %v", err)
		c.Error(fmt.Errorf("run.NewServicesClient: %v", err))
		return
	}
	defer runClient.Close()

	envVars := []*runpb.EnvVar{}
	for k, v := range req.Env {
		envVars = append(envVars, &runpb.EnvVar{Name: k, Values: &runpb.EnvVar_Value{Value: v}})
	}

	parent := fmt.Sprintf("projects/%s/locations/%s", projectID, region)

	_, err = runClient.CreateService(ctx, &runpb.CreateServiceRequest{
		Parent:    parent,
		ServiceId: req.ServiceName,
		Service: &runpb.Service{
			Template: &runpb.RevisionTemplate{
				Containers: []*runpb.Container{
					{
						Image: imagePath,
						Env:   envVars,
						Resources: &runpb.ResourceRequirements{
							Limits: map[string]string{"memory": "512Mi"},
						},
					},
				},
			},
			Ingress: runpb.IngressTraffic_INGRESS_TRAFFIC_ALL,
		},
	})
	if err != nil {
		logStream <- fmt.Sprintf("Cloud Run deploy failed: %v", err)
		c.Error(fmt.Errorf("Cloud Run deploy failed: %v", err))
		return
	}

	logStream <- "Making service publicly accessible..."
	if err := utils.RunCommandWithOutput("gcloud", []string{"run", "services", "add-iam-policy-binding", req.ServiceName,
		"--region=" + region,
		"--member=allUsers",
		"--role=roles/run.invoker",
		"--quiet"}, logWriter); err != nil {
		logStream <- fmt.Sprintf("Warning: Failed to make service public: %v", err)
	}

	serviceURL := fmt.Sprintf("https://%s.brogramiz.info", req.ServiceName)
	if err := db.AddProject(c, user, gitCleanURL, serviceURL, framework, req.ServiceName); err != nil {
		logStream <- fmt.Sprintf("Failed to save project: %v", err)
		c.Error(err)
		return
	}

	logStream <- "Deployment successful! Service URL: " + serviceURL
	c.JSON(200, DeployResponse{URL: serviceURL, DeploymentID: deploymentID})
}
