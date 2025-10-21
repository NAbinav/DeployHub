package handler

import (
	"deployhub/db"
	"deployhub/helper"
	"deployhub/jwt"
	"deployhub/utils"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
)

func DeployIDHandler(c *gin.Context) {
	repo := c.Param("id")
	repo = strings.ToLower(repo)
	projectID := os.Getenv("GCLOUD_PROJECT_ID")
	if projectID == "" {
		c.Error(errors.New("PROJECT_ID environment variable not set"))
		return
	}
	token, err := c.Cookie("token")
	if err != nil {
		fmt.Println(err)
		return
	}
	fmt.Println(token)
	username, err := jwt.Verify_JWT(token)
	if err != nil {
		fmt.Println(err)
		c.String(400, err.Error())
		return
	}
	access_token, err := db.UserToken(username, c.Request.Context())
	map_token, err := helper.StringToToken(access_token)
	git_url := "https://" + map_token["access_token"].(string) + "@github.com/" + username + "/" + repo
	tempDir := filepath.Join("/tmp", fmt.Sprintf("repo-%d", time.Now().UnixNano()))
	if err := utils.RunCommand("git", "clone", git_url, tempDir); err != nil {
		c.Error(err)
		c.String(400, err.Error())
		return
	}

	framework := utils.DetectFramework(tempDir)
	fmt.Println("🔍 Detected framework:", framework)

	if err := utils.CreateDockerfile(tempDir, framework); err != nil {
		c.Error(err)
		c.String(400, err.Error())
		return
	}

	if err := utils.CreateDockerignore(tempDir); err != nil {
		fmt.Println("Warning: Failed to create .dockerignore:", err)
		c.String(400, err.Error())
	}

	if err := os.Chdir(tempDir); err != nil {
		c.Error(err)
		c.String(400, err.Error())
		return
	}

	imageName := fmt.Sprintf("%s-image", repo)
	repoName := "deploy-hub"
	repoPath := fmt.Sprintf("asia-south1-docker.pkg.dev/%s/%s/%s", projectID, repoName, imageName)
	region := "asia-south1"

	if err := utils.RunCommand("docker", "build", "-t", imageName, "."); err != nil {
		c.Error(err)
		c.String(400, err.Error())
		return
	}

	if err := utils.RunCommand("docker", "tag", imageName, repoPath); err != nil {
		c.Error(err)
		c.String(400, err.Error())
		return
	}

	if err := utils.RunCommand("gcloud", "auth", "configure-docker", "asia-south1-docker.pkg.dev", "--quiet"); err != nil {
		c.Error(err)
		c.String(400, err.Error())
		return
	}

	if err := utils.RunCommand("docker", "push", repoPath); err != nil {
		c.Error(err)
		c.String(400, err.Error())
		return
	}

	// env_string := utils.ENVString(.Env)
	// fmt.Println(env_string)
	deploy_cmd := fmt.Sprintf("gcloud run deploy %s --image %s --platform managed --region %s --allow-unauthenticated --timeout 300s --memory 512M", repo, repoPath, region)
	if _, err := utils.RunCommandArray(strings.Split(deploy_cmd, " ")); err != nil {
		c.Error(err)
		log_cmd := "gcloud alpha run services logs read " + repo
		out, err := utils.RunCommandArray(strings.Split(log_cmd, " "))
		if err != nil {
			c.String(400, error.Error(err))
		}
		c.String(400, out)
		return
	}

	os.Chdir("/tmp")
	_ = os.RemoveAll(tempDir)
	_ = utils.RunCommand("docker", "rmi", imageName)
	serviceURL := fmt.Sprintf("https://%s-%s.a.run.app", repo, region)
	c.JSON(200, DeployResponse{URL: serviceURL})

}
