package handler

import (
	"context"
	"deployhub/db"
	// "deployhub/helper"
	"deployhub/jwt"
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

	projectID := os.Getenv("GCLOUD_PROJECT_ID")
	if projectID == "" {
		c.Error(errors.New("GCLOUD_PROJECT_ID not set"))
		return
	}

	region := "asia-south1"
	repoID := "deploy-hub"
	imageName := fmt.Sprintf("%s-image", req.ServiceName)
	imagePath := fmt.Sprintf("%s-docker.pkg.dev/%s/%s/%s:latest", region, projectID, repoID, imageName)

	token, err := c.Cookie("token")
	if err != nil {
		c.Error(err)
		return
	}

	user, err := jwt.Verify_JWT(token)
	if err != nil {
		c.Error(err)
		return
	}

	var access_token oauth2.Token
	access_token, err = db.UserToken(c, user)
	if err != nil {
		c.Error(err)
		return
	}

	gitURL := "https://" + access_token.AccessToken + "@github.com/" + user + "/" + req.GitURL
	gitCleanURL := "https://www.github.com/" + user + "/" + req.GitURL

	tempDir := filepath.Join("/tmp", fmt.Sprintf("repo-%d", time.Now().UnixNano()))
	if err := utils.RunCommandWithOutput("git", []string{"clone", gitURL, tempDir}, nil); err != nil {
		c.Error(err)
		return
	}
	defer os.RemoveAll(tempDir) // Clean up

	framework := utils.DetectFramework(tempDir)

	if err := utils.CreateDockerfile(tempDir, framework); err != nil {
		c.Error(err)
		return
	}

	arClient, err := artifactregistry.NewClient(ctx, option.WithCredentialsFile(os.Getenv("GOOGLE_APPLICATION_CREDENTIALS")))
	if err != nil {
		c.Error(fmt.Errorf("artifactregistry.NewClient: %v", err))
		return
	}
	defer arClient.Close()

	repoParent := fmt.Sprintf("projects/%s/locations/%s", projectID, region)
	repoName := fmt.Sprintf("%s/repositories/%s", repoParent, repoID)

	_, err = arClient.GetRepository(ctx, &artifactregistrypb.GetRepositoryRequest{Name: repoName})
	if err != nil {
		_, err = arClient.CreateRepository(ctx, &artifactregistrypb.CreateRepositoryRequest{
			Parent:       repoParent,
			RepositoryId: repoID,
			Repository: &artifactregistrypb.Repository{
				Format:      artifactregistrypb.Repository_DOCKER,
				Description: "DeployHub managed Docker repo",
			},
		})
		if err != nil {
			c.Error(fmt.Errorf("CreateRepository: %v", err))
			return
		}
	}

	if err := utils.RunCommandWithOutput("gcloud", []string{"auth", "configure-docker", fmt.Sprintf("%s-docker.pkg.dev", region), "--quiet"}, nil); err != nil {
		c.Error(fmt.Errorf("Docker auth configuration failed: %v", err))
		return
	}

	if err := utils.RunCommandWithOutput("docker", []string{"build", "-t", imagePath, tempDir}, nil); err != nil {
		c.Error(fmt.Errorf("Docker build failed: %v", err))
		return
	}

	if err := utils.RunCommandWithOutput("docker", []string{"push", imagePath}, nil); err != nil {
		c.Error(fmt.Errorf("Docker push failed: %v", err))
		return
	}

	runClient, err := run.NewServicesClient(ctx, option.WithCredentialsFile(os.Getenv("GOOGLE_APPLICATION_CREDENTIALS")))
	if err != nil {
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
		c.Error(fmt.Errorf("Cloud Run deploy failed: %v", err))
		return
	}

	if err := utils.RunCommandWithOutput("gcloud", []string{"run", "services", "add-iam-policy-binding", req.ServiceName,
		"--region=" + region,
		"--member=allUsers",
		"--role=roles/run.invoker",
		"--quiet"}, nil); err != nil {
		c.Error(fmt.Errorf("Issues in Deployment"))
	}

	serviceURL := fmt.Sprintf("https://%s.brogramiz.info", req.ServiceName)
	if err := db.AddProject(c, user, gitCleanURL, serviceURL, framework, req.ServiceName); err != nil {
		c.Error(err)
		return
	}

	c.JSON(200, DeployResponse{URL: serviceURL, DeploymentID: deploymentID})
}
