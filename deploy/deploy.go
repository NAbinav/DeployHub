package deploy

import (
	"bytes"
	"context"
	"deployhub/db"
	"deployhub/helper"
	"deployhub/schema"
	"deployhub/utils"
	"deployhub/webhook"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	artifactregistry "cloud.google.com/go/artifactregistry/apiv1"
	"cloud.google.com/go/artifactregistry/apiv1/artifactregistrypb"
	run "cloud.google.com/go/run/apiv2"
	"cloud.google.com/go/run/apiv2/runpb"
	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
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
	result.Status = "running"

	config, err := initializeDeploymentConfig(params, result.DeploymentID)
	if err != nil {
		result.Error = err
		result.Status = "failed"
		return result
	}

	tempDir, err := cloneRepository(config)
	if err != nil {
		result.Error = err
		result.Status = "failed"
		return result
	}
	defer os.RemoveAll(tempDir)

	fmt.Println("cloned!!")
	fmt.Println(time.Since(start))

	framework, err := prepareDockerfile(tempDir)
	if err != nil {
		result.Error = err
		result.Status = "failed"
		return result
	}
	result.Framework = framework

	if err := ensureArtifactRegistry(deployCtx, config); err != nil {
		result.Error = err
		result.Status = "failed"
		return result
	}

	if err := buildAndPushDockerImage(config, tempDir, start); err != nil {
		result.Error = err
		result.Status = "failed"
		return result
	}

	if err := DeployToCloudRun(deployCtx, config, params); err != nil {
		result.Error = err
		result.Status = "failed"
		return result
	}

	if err := configurePublicAccess(config, params.ServiceName); err != nil {
		fmt.Println("Warning: failed to configure public access:", err)
	}

	fmt.Println("started deployment!!!!!!")
	var buf bytes.Buffer
	// Use inspect to get the RepoDigest (the cloud address)
	err = utils.RunCommandWithOutput(
		"docker",
		[]string{"inspect", "--format", "{{index .RepoDigests 0}}", config.ImagePath},
		&buf,
	)

	if err != nil {
		fmt.Println("Failed getting registry digest:", err)
		return result
	}

	// Result looks like: region-docker.pkg.dev/project/repo/name@sha256:abcdef...
	rawPath := strings.TrimSpace(buf.String())
	parts := strings.Split(rawPath, "@")

	if len(parts) < 2 {
		fmt.Println("No registry digest found in path")
		return result
	}

	// THIS is the ID you save to your database (e.g., "sha256:2273c0f9...")
	dockerImageID := parts[1]
	if config.Branch == "" {
		config.Branch = "main"
	}
	db.AddDockerID(context.Background(), params.ServiceName, dockerImageID, config.Branch)
	if err := setupWebhook(config, params); err != nil {
		fmt.Println("webhook error:", err)
	}

	fmt.Println(time.Since(start))

	result.ServiceURL = fmt.Sprintf("https://%s.brogramiz.info", params.ServiceName)

	fmt.Println(result.DeploymentID, result.ServiceURL, framework, params.ServiceName, result.Status)
	err = db.UpdateProjectAfterDeploy(ctx, result.ServiceURL, framework, params.ServiceName, result.Status)
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
		Branch:       params.Branch,
		GitURL:       gitURL,
		GitCleanURL:  gitCleanURL,
		DeploymentID: deploymentID,
	}, nil
}

func cloneRepository(config *schema.DeploymentConfig) (string, error) {
	// Create unique temp directory
	branch := config.Branch
	if branch == "" {
		branch = "main" // Default to main if not specified
	}
	fmt.Println(branch)
	tempDir := filepath.Join("/tmp", fmt.Sprintf("repo-%d", time.Now().UnixNano()))

	fmt.Printf("Cloning branch: %s from %s\n", config.Branch, config.GitURL)

	// Configure Clone Options
	cloneOptions := &git.CloneOptions{
		URL:           config.GitURL,
		Progress:      os.Stdout,
		SingleBranch:  true,                                    // Clone only the specific branch (lighter)
		Depth:         1,                                       // Shallow clone (faster)
		ReferenceName: plumbing.NewBranchReferenceName(branch), // Target specific branch
	}

	_, err := git.PlainClone(tempDir, false, cloneOptions)
	if err != nil {
		return "", fmt.Errorf("git clone failed: %v", err)
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

func DeployToCloudRun(ctx context.Context, config *schema.DeploymentConfig, params schema.DeployParams) error {
	client, err := run.NewServicesClient(
		ctx,
		option.WithCredentialsFile(os.Getenv("GOOGLE_APPLICATION_CREDENTIALS")),
	)
	if err != nil {
		return fmt.Errorf("run.NewServicesClient: %v", err)
	}
	defer client.Close()

	serviceID := fmt.Sprintf("projects/%s/locations/%s/services/%s", config.ProjectID, config.Region, params.ServiceName)
	parent := fmt.Sprintf("projects/%s/locations/%s", config.ProjectID, config.Region)

	cleanTag := helper.SanitizeTag(params.Tag)
	revisionName := fmt.Sprintf("%s-%s-%d", params.ServiceName, cleanTag, time.Now().Unix())

	//  Fetch existing service to preserve current traffic
	existingService, err := client.GetService(ctx, &runpb.GetServiceRequest{Name: serviceID})
	isNewService := err != nil

	var traffic []*runpb.TrafficTarget
	isProd := params.Tag == "latest" || params.Tag == "main" || params.Tag == "master"

	if isNewService {
		// First deploy: Always 100% to new revision
		traffic = []*runpb.TrafficTarget{
			{
				Type:     runpb.TrafficTargetAllocationType_TRAFFIC_TARGET_ALLOCATION_TYPE_REVISION,
				Revision: revisionName,
				Percent:  100,
				Tag:      cleanTag,
			},
		}
	} else if isProd {
		// Production deploy: Move 100% to new revision
		traffic = []*runpb.TrafficTarget{
			{
				Type:     runpb.TrafficTargetAllocationType_TRAFFIC_TARGET_ALLOCATION_TYPE_REVISION,
				Revision: revisionName,
				Percent:  100,
			},
		}
	} else {
		// Preview deploy: Keep existing 100% traffic, add new revision with 0% and tag
		currentProdRevision := ""
		for _, t := range existingService.Traffic {
			if t.Percent == 100 {
				currentProdRevision = t.Revision
				break
			}
		}

		// Keep old traffic
		if currentProdRevision != "" {
			traffic = append(traffic, &runpb.TrafficTarget{
				Type:     runpb.TrafficTargetAllocationType_TRAFFIC_TARGET_ALLOCATION_TYPE_REVISION,
				Revision: currentProdRevision,
				Percent:  100,
			})
		}

		// Add new preview
		traffic = append(traffic, &runpb.TrafficTarget{
			Type:     runpb.TrafficTargetAllocationType_TRAFFIC_TARGET_ALLOCATION_TYPE_REVISION,
			Revision: revisionName,
			Percent:  0,
			Tag:      cleanTag,
		})
	}

	// 3. Construct Service
	service := &runpb.Service{
		Template: &runpb.RevisionTemplate{
			Revision: revisionName,
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

	// 4. Execute Update or Create
	if isNewService {
		service.Name = "" // Must be empty for create
		req := &runpb.CreateServiceRequest{
			Parent:    parent,
			ServiceId: params.ServiceName,
			Service:   service,
		}
		_, err = client.CreateService(ctx, req)
	} else {
		service.Name = serviceID // Must be full path for update
		req := &runpb.UpdateServiceRequest{
			Service: service,
		}
		_, err = client.UpdateService(ctx, req)
	}

	if err != nil {
		return fmt.Errorf("Cloud Run deploy failed: %v", err)
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
