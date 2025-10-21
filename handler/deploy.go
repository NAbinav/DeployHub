package handler

import (
	"deployhub/utils"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
)

type DeployRequest struct {
	GitURL      string            `json:"git_url"`
	ServiceName string            `json:"name"`
	Env         map[string]string `json:"env"`
}

type DeployResponse struct {
	URL   string `json:"url"`
	Error string `json:"error,omitempty"`
}

func DeployHandler(c *gin.Context) {
	var req DeployRequest
	err := c.BindJSON(&req)
	if err != nil {
		c.Error(err)
	}
	projectID := os.Getenv("GCLOUD_PROJECT_ID")
	if projectID == "" {
		c.Error(errors.New("PROJECT_ID environment variable not set"))
		return
	}

	tempDir := filepath.Join("/tmp", fmt.Sprintf("repo-%d", time.Now().UnixNano()))
	if err := utils.RunCommand("git", "clone", req.GitURL, tempDir); err != nil {
		c.Error(err)
		return
	}

	framework := utils.DetectFramework(tempDir)
	fmt.Println("🔍 Detected framework:", framework)

	if err := utils.CreateDockerfile(tempDir, framework); err != nil {
		c.Error(err)
		return
	}

	if err := utils.CreateDockerignore(tempDir); err != nil {
		fmt.Println("Warning: Failed to create .dockerignore:", err)
	}

	if err := os.Chdir(tempDir); err != nil {
		c.Error(err)
		return
	}

	imageName := fmt.Sprintf("%s-image", req.ServiceName)
	repoName := "deploy-hub"
	repoPath := fmt.Sprintf("asia-south1-docker.pkg.dev/%s/%s/%s", projectID, repoName, imageName)
	region := "asia-south1"

	if err := utils.RunCommand("docker", "build", "-t", imageName, "."); err != nil {
		c.Error(err)
		return
	}

	if err := utils.RunCommand("docker", "tag", imageName, repoPath); err != nil {
		c.Error(err)
		return
	}

	if err := utils.RunCommand("gcloud", "auth", "configure-docker", "asia-south1-docker.pkg.dev", "--quiet"); err != nil {
		c.Error(err)
		return
	}

	if err := utils.RunCommand("docker", "push", repoPath); err != nil {
		c.Error(err)
		return
	}

	env_string := utils.ENVString(req.Env)
	fmt.Println(env_string)
	deploy_cmd := fmt.Sprintf("gcloud run deploy %s --image %s --platform managed --region %s --allow-unauthenticated --timeout 300s --memory 512M --set-env-vars %s", req.ServiceName, repoPath, region, env_string)
	if _, err := utils.RunCommandArray(strings.Split(deploy_cmd, " ")); err != nil {
		c.Error(err)
		return
	}

	os.Chdir("/tmp")
	_ = os.RemoveAll(tempDir)
	_ = utils.RunCommand("docker", "rmi", imageName)
	serviceURL := fmt.Sprintf("https://%s.brogramiz.info", req.ServiceName)
	c.JSON(200, DeployResponse{URL: serviceURL})

}
