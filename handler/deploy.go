package handler

import (
	"deployhub/utils"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"
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

func DeployHandler(w http.ResponseWriter, r *http.Request) {
	var req DeployRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	projectID := os.Getenv("GCLOUD_PROJECT_ID")
	if projectID == "" {
		json.NewEncoder(w).Encode(DeployResponse{Error: "PROJECT_ID environment variable not set"})
		return
	}

	tempDir := filepath.Join("/tmp", fmt.Sprintf("repo-%d", time.Now().UnixNano()))
	if err := utils.RunCommand("git", "clone", req.GitURL, tempDir); err != nil {
		json.NewEncoder(w).Encode(DeployResponse{Error: err.Error()})
		return
	}

	framework := utils.DetectFramework(tempDir)
	fmt.Println("🔍 Detected framework:", framework)

	// Create Dockerfile and .dockerignore
	if err := utils.CreateDockerfile(tempDir, framework); err != nil {
		json.NewEncoder(w).Encode(DeployResponse{Error: "Failed to create Dockerfile: " + err.Error()})
		return
	}

	if err := utils.CreateDockerignore(tempDir); err != nil {
		fmt.Println("Warning: Failed to create .dockerignore:", err)
	}

	if err := os.Chdir(tempDir); err != nil {
		json.NewEncoder(w).Encode(DeployResponse{Error: err.Error()})
		return
	}

	imageName := fmt.Sprintf("%s-image", req.ServiceName)
	repoName := "deploy-hub"
	repoPath := fmt.Sprintf("asia-south1-docker.pkg.dev/%s/%s/%s", projectID, repoName, imageName)
	region := "asia-south1"

	if err := utils.RunCommand("docker", "build", "-t", imageName, "."); err != nil {
		json.NewEncoder(w).Encode(DeployResponse{Error: "Docker build failed: " + err.Error()})
		return
	}

	if err := utils.RunCommand("docker", "tag", imageName, repoPath); err != nil {
		json.NewEncoder(w).Encode(DeployResponse{Error: err.Error()})
		return
	}

	if err := utils.RunCommand("gcloud", "auth", "configure-docker", "asia-south1-docker.pkg.dev", "--quiet"); err != nil {
		json.NewEncoder(w).Encode(DeployResponse{Error: err.Error()})
		return
	}

	if err := utils.RunCommand("docker", "push", repoPath); err != nil {
		json.NewEncoder(w).Encode(DeployResponse{Error: err.Error()})
		return
	}

	env_string := utils.ENVString(req.Env)
	fmt.Println(env_string)
	deploy_cmd := fmt.Sprintf("gcloud run deploy %s --image %s --platform managed --region %s --allow-unauthenticated --timeout 300s --memory 512M --set-env-vars %s", req.ServiceName, repoPath, region, env_string)
	if err := utils.RunCommandArray(strings.Split(deploy_cmd, " ")); err != nil {
		json.NewEncoder(w).Encode(DeployResponse{Error: err.Error()})
		return
	}

	os.Chdir("/tmp")
	_ = os.RemoveAll(tempDir)
	_ = utils.RunCommand("docker", "rmi", imageName)
	serviceURL := fmt.Sprintf("https://%s-%s.a.run.app", req.ServiceName, region)
	json.NewEncoder(w).Encode(DeployResponse{URL: serviceURL})

}
