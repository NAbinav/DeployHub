package deploy

import (
	"context"
	"deployhub/db"
	"deployhub/schema"
	"deployhub/utils"
	"deployhub/webhook"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"time"

	artifactregistry "cloud.google.com/go/artifactregistry/apiv1"
	"cloud.google.com/go/artifactregistry/apiv1/artifactregistrypb"
	run "cloud.google.com/go/run/apiv2"
	"cloud.google.com/go/run/apiv2/runpb"
	"github.com/google/uuid"
	"google.golang.org/api/option"
)

func Deploy(ctx context.Context, params schema.DeployParams) schema.DeployResult {
	start := time.Now()

	// Create context with timeout if parent context doesn't have one
	deployCtx, cancel := context.WithTimeout(ctx, 15*time.Minute)
	defer cancel()

	result := schema.DeployResult{
		DeploymentID: uuid.New().String(),
	}

	config, err := initializeDeploymentConfig(params, result.DeploymentID)
	if err != nil {
		result.Error = err
		return result
	}

	tempDir, err := cloneRepository(config)
	if err != nil {
		result.Error = err
		return result
	}
	defer os.RemoveAll(tempDir)

	fmt.Println("cloned!!")
	fmt.Println(time.Since(start))

	framework, err := prepareDockerfile(tempDir)
	if err != nil {
		result.Error = err
		return result
	}
	result.Framework = framework

	if err := ensureArtifactRegistry(deployCtx, config); err != nil {
		result.Error = err
		return result
	}

	if err := buildAndPushDockerImage(config, tempDir, start); err != nil {
		result.Error = err
		return result
	}

	if err := deployToCloudRun(deployCtx, config, params); err != nil {
		result.Error = err
		return result
	}

	if err := configurePublicAccess(config, params.ServiceName); err != nil {
		fmt.Println("Warning: failed to configure public access:", err)
	}

	fmt.Println("started deployment!!!!!!")

	if err := setupWebhook(config, params); err != nil {
		fmt.Println("webhook error:", err)
	}

	fmt.Println(time.Since(start))

	result.ServiceURL = fmt.Sprintf("https://%s.brogramiz.info", params.ServiceName)

	status := "running"
	fmt.Println(result.ServiceURL, framework, params.ServiceName, status)
	err = db.UpdateProjectAfterDeploy(ctx, result.ServiceURL, framework, params.ServiceName, status)
	if err != nil {
		fmt.Println("UPDATE DB:", err)
	}

	return result
}

func initializeDeploymentConfig(params schema.DeployParams, deploymentID string) (*schema.DeploymentConfig, error) {
	projectID := os.Getenv("GCLOUD_PROJECT_ID")
	if projectID == "" {
		return nil, errors.New("GCLOUD_PROJECT_ID not set")
	}

	region := "asia-south1"
	repoID := "deploy-hub"
	imageName := fmt.Sprintf("%s-image", params.ServiceName)
	imagePath := fmt.Sprintf("%s-docker.pkg.dev/%s/%s/%s:%s", region, projectID, repoID, imageName, params.Tag)
	gitURL := "https://" + params.AccessToken.AccessToken + "@github.com/" + params.User + "/" + params.GitURL
	gitCleanURL := "https://www.github.com/" + params.User + "/" + params.GitURL

	return &schema.DeploymentConfig{
		ProjectID:    projectID,
		Region:       region,
		RepoID:       repoID,
		ImageName:    imageName,
		ImagePath:    imagePath,
		User:         params.User,
		AccessToken:  params.AccessToken,
		GitURL:       gitURL,
		GitCleanURL:  gitCleanURL,
		DeploymentID: deploymentID,
	}, nil
}

func cloneRepository(config *schema.DeploymentConfig) (string, error) {
	tempDir := filepath.Join("/tmp", fmt.Sprintf("repo-%d", time.Now().UnixNano()))
	if err := utils.RunCommandWithOutput("git", []string{"clone", config.GitURL, tempDir}, nil); err != nil {
		return "", err
	}
	return tempDir, nil
}

func prepareDockerfile(tempDir string) (string, error) {
	framework := utils.DetectFramework(tempDir)
	if err := utils.CreateDockerfile(tempDir, framework); err != nil {
		return "", err
	}
	return framework, nil
}

func ensureArtifactRegistry(ctx context.Context, config *schema.DeploymentConfig) error {
	arClient, err := artifactregistry.NewClient(ctx)
	if err != nil {
		return fmt.Errorf("artifactregistry.NewClient: %v", err)
	}
	defer arClient.Close()

	repoParent := fmt.Sprintf("projects/%s/locations/%s", config.ProjectID, config.Region)
	repoName := fmt.Sprintf("%s/repositories/%s", repoParent, config.RepoID)

	_, err = arClient.GetRepository(ctx, &artifactregistrypb.GetRepositoryRequest{Name: repoName})
	if err != nil {
		_, err = arClient.CreateRepository(ctx, &artifactregistrypb.CreateRepositoryRequest{
			Parent:       repoParent,
			RepositoryId: config.RepoID,
			Repository: &artifactregistrypb.Repository{
				Format:      artifactregistrypb.Repository_DOCKER,
				Description: "DeployHub managed Docker repo",
			},
		})
		if err != nil {
			return fmt.Errorf("CreateRepository: %v", err)
		}
	}
	return nil
}

func buildAndPushDockerImage(config *schema.DeploymentConfig, tempDir string, start time.Time) error {
	dockerRegistry := fmt.Sprintf("%s-docker.pkg.dev", config.Region)
	if err := utils.RunCommandWithOutput("gcloud", []string{"auth", "configure-docker", dockerRegistry, "--quiet"}, nil); err != nil {
		return fmt.Errorf("Docker auth configuration failed: %v", err)
	}

	if err := utils.RunCommandWithOutput("docker", []string{"build", "-t", config.ImagePath, tempDir}, nil); err != nil {
		return fmt.Errorf("Docker build failed: %v", err)
	}

	fmt.Println("docker image built!!")
	fmt.Println(time.Since(start))

	if err := utils.RunCommandWithOutput("docker", []string{"push", config.ImagePath}, nil); err != nil {
		return fmt.Errorf("Docker push failed: %v", err)
	}

	fmt.Println("Pushed docker!!")
	fmt.Println(time.Since(start))

	return nil
}

func deployToCloudRun(ctx context.Context, config *schema.DeploymentConfig, params schema.DeployParams) error {
	client, err := run.NewServicesClient(
		ctx,
		option.WithCredentialsFile(os.Getenv("GOOGLE_APPLICATION_CREDENTIALS")),
	)
	if err != nil {
		return fmt.Errorf("run.NewServicesClient: %v", err)
	}
	defer client.Close()

	serviceName := fmt.Sprintf(
		"projects/%s/locations/%s/services/%s",
		config.ProjectID,
		config.Region,
		params.ServiceName,
	)

	var traffic []*runpb.TrafficTarget
	if params.Tag != "" && params.Tag != "latest" {
		// For specific tags, create a tagged revision with 0% traffic
		traffic = []*runpb.TrafficTarget{
			{
				Type:    runpb.TrafficTargetAllocationType_TRAFFIC_TARGET_ALLOCATION_TYPE_LATEST,
				Percent: 100,
				Tag:     params.Tag,
			},
		}
	} else {
		traffic = []*runpb.TrafficTarget{
			{
				Type:    runpb.TrafficTargetAllocationType_TRAFFIC_TARGET_ALLOCATION_TYPE_LATEST,
				Percent: 100,
			},
		}
	}
	service := &runpb.Service{
		Template: &runpb.RevisionTemplate{
			Containers: []*runpb.Container{
				{
					Image: config.ImagePath,
					Env:   buildEnvVars(params.Env),
					Resources: &runpb.ResourceRequirements{
						Limits:  map[string]string{"memory": "512Mi"},
						CpuIdle: true,
					},
				},
			},
		},
		Traffic: traffic,
		Ingress: runpb.IngressTraffic_INGRESS_TRAFFIC_ALL,
	}

	// Try updating the service first
	service.Name = serviceName // ONLY for update
	_, err = client.UpdateService(ctx, &runpb.UpdateServiceRequest{
		Service: service,
	})
	if err == nil {
		return nil
	}

	// If update fails, create service (first-time deploy)
	service.Name = "" // MUST be empty for create
	parent := fmt.Sprintf("projects/%s/locations/%s", config.ProjectID, config.Region)
	_, err = client.CreateService(ctx, &runpb.CreateServiceRequest{
		Parent:    parent,
		ServiceId: params.ServiceName,
		Service:   service,
	})
	if err != nil {
		return fmt.Errorf("Cloud Run create failed: %v", err)
	}

	return nil
}

func buildEnvVars(env map[string]string) []*runpb.EnvVar {
	envVars := make([]*runpb.EnvVar, 0, len(env))
	for k, v := range env {
		envVars = append(envVars, &runpb.EnvVar{
			Name:   k,
			Values: &runpb.EnvVar_Value{Value: v},
		})
	}
	return envVars
}

func configurePublicAccess(config *schema.DeploymentConfig, serviceName string) error {
	return utils.RunCommandWithOutput("gcloud", []string{
		"run", "services", "add-iam-policy-binding", serviceName,
		"--region=" + config.Region,
		"--project=" + config.ProjectID,
		"--member=allUsers",
		"--role=roles/run.invoker",
		"--quiet",
	}, nil)
}

func setupWebhook(config *schema.DeploymentConfig, params schema.DeployParams) error {
	webhookUrl := fmt.Sprintf("https://brogramiz.info/api/webhook/%s", params.ServiceName)
	err := webhook.AddWebhook(config.AccessToken, config.User, params.GitURL, webhookUrl)
	if err == nil {
		fmt.Println("webhook added guyss")
	}
	return err
}
