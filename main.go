package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httputil"
	"net/url"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"
)

func GetWebsite(w http.ResponseWriter, r *http.Request) {
	projectName := strings.Split(r.Host, ".")[0]
	targetURL := "https://" + projectName + "-" + os.Getenv("GCLOUD_PROJECT_NUMBER") + ".asia-south1.run.app"
	target, err := url.Parse(targetURL)
	if err != nil {
		http.Error(w, "Bad target URL", http.StatusInternalServerError)
		fmt.Println("URL parse error:", err)
		return
	}
	proxy := httputil.NewSingleHostReverseProxy(target)
	r.Host = target.Host
	fmt.Println("Proxying request to:", targetURL)
	proxy.ServeHTTP(w, r)
}

type DeployRequest struct {
	GitURL      string `json:"git_url"`
	ServiceName string `json:"name"`
}

type DeployResponse struct {
	URL   string `json:"url"`
	Error string `json:"error,omitempty"`
}

func runCommand(name string, args ...string) error {
	fmt.Printf("Running: %s %v\n", name, args)
	cmd := exec.Command(name, args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Stdin = os.Stdin
	return cmd.Run()
}

func deployHandler(w http.ResponseWriter, r *http.Request) {
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
	if err := runCommand("git", "clone", req.GitURL, tempDir); err != nil {
		json.NewEncoder(w).Encode(DeployResponse{Error: err.Error()})
		return
	}

	// Change working directory
	if err := os.Chdir(tempDir); err != nil {
		json.NewEncoder(w).Encode(DeployResponse{Error: err.Error()})
		return
	}

	imageName := fmt.Sprintf("%s-image", req.ServiceName)
	repoName := "deploy-hub"
	repoPath := fmt.Sprintf("asia-south1-docker.pkg.dev/%s/%s/%s", projectID, repoName, imageName)
	region := "asia-south1"

	// Build with pack
	if err := runCommand("pack", "build", imageName, "--builder", "gcr.io/buildpacks/builder"); err != nil {
		json.NewEncoder(w).Encode(DeployResponse{Error: err.Error()})
		return
	}

	// Tag & push
	if err := runCommand("docker", "tag", imageName, repoPath); err != nil {
		json.NewEncoder(w).Encode(DeployResponse{Error: err.Error()})
		return
	}

	// Authenticate Docker to Artifact Registry
	if err := runCommand("gcloud", "auth", "configure-docker", "asia-south1-docker.pkg.dev", "--quiet"); err != nil {
		json.NewEncoder(w).Encode(DeployResponse{Error: err.Error()})
		return
	}

	if err := runCommand("docker", "push", repoPath); err != nil {
		json.NewEncoder(w).Encode(DeployResponse{Error: err.Error()})
		return
	}

	// Deploy to Cloud Run
	if err := runCommand("gcloud", "run", "deploy", req.ServiceName,
		"--image", repoPath,
		"--platform", "managed",
		"--region", region,
		"--allow-unauthenticated"); err != nil {
		json.NewEncoder(w).Encode(DeployResponse{Error: err.Error()})
		return
	}

	// Clean up
	os.Chdir("/tmp")
	_ = os.RemoveAll(tempDir)
	_ = runCommand("docker", "rmi", imageName)

	// Return Cloud Run URL
	serviceURL := fmt.Sprintf("https://%s-%s.a.run.app", req.ServiceName, region)
	json.NewEncoder(w).Encode(DeployResponse{URL: serviceURL})
}

func main() {
	http.HandleFunc("/", GetWebsite)
	http.HandleFunc("/deploy", deployHandler)
	fmt.Println("Proxy server running on :8080")
	http.ListenAndServe(":8080", nil)
}
