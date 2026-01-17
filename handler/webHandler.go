package handler

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"strings"

	"deployhub/db"
	"deployhub/schema"

	"github.com/gin-gonic/gin"
)

type GitHubWebhookPayload struct {
	Ref        string `json:"ref"`
	After      string `json:"after"`
	Repository struct {
		CloneURL string `json:"clone_url"`
		Name     string `json:"name"`
		Owner    struct {
			Login string `json:"login"`
		} `json:"owner"`
	} `json:"repository"`
	Sender struct {
		Login string `json:"login"`
	} `json:"sender"`
	PullRequest *struct {
		Number int `json:"number"`
		Head   struct {
			Ref string `json:"ref"`
			SHA string `json:"sha"`
		} `json:"head"`
		State string `json:"state"`
	} `json:"pull_request,omitempty"`
}

func sanitizeTag(s string) string {
	s = strings.ToLower(strings.ReplaceAll(strings.ReplaceAll(s, "/", "-"), "_", "-"))
	var result strings.Builder
	for _, ch := range s {
		if (ch >= 'a' && ch <= 'z') || (ch >= '0' && ch <= '9') || ch == '-' {
			result.WriteRune(ch)
		}
	}
	return result.String()
}

func verifySignature(payload []byte, signature, secret string) bool {
	if secret == "" {
		return true
	}
	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write(payload)
	expectedMAC := "sha256=" + hex.EncodeToString(mac.Sum(nil))
	return hmac.Equal([]byte(expectedMAC), []byte(signature))
}

func injectTokenIntoGitURL(cloneURL, token string) string {
	if strings.HasPrefix(cloneURL, "https://") {
		return strings.Replace(cloneURL, "https://", fmt.Sprintf("https://%s@", token), 1)
	}
	return cloneURL
}

func HandleGitHubWebhook(c *gin.Context) {
	body, err := io.ReadAll(c.Request.Body)
	if err != nil {
		c.JSON(400, gin.H{"error": "Failed to read request body"})
		return
	}

	if !verifySignature(body, c.GetHeader("X-Hub-Signature-256"), getWebhookSecret()) {
		c.JSON(401, gin.H{"error": "Invalid signature"})
		return
	}

	var payload GitHubWebhookPayload
	if err := json.Unmarshal(body, &payload); err != nil {
		c.JSON(400, gin.H{"error": "Invalid JSON payload"})
		return
	}

	eventType := c.GetHeader("X-GitHub-Event")
	serviceName := c.Param("id")
	var tag, previewURL string
	var shouldDeploy bool

	switch eventType {
	case "push":
		if !strings.HasPrefix(payload.Ref, "refs/heads/") {
			c.JSON(200, gin.H{"message": "Ignoring non-branch push"})
			return
		}
		branch := strings.TrimPrefix(payload.Ref, "refs/heads/")
		if branch == "main" || branch == "master" {
			tag = "latest"
		} else {
			tag = sanitizeTag(branch)
			previewURL = fmt.Sprintf("%s-%s.yourdomain.com", tag, payload.Repository.Name)
		}
		shouldDeploy = true

	case "pull_request":
		if payload.PullRequest == nil || payload.PullRequest.State != "open" {
			c.JSON(200, gin.H{"message": "Ignoring closed or invalid PR"})
			return
		}
		tag = fmt.Sprintf("pr-%d", payload.PullRequest.Number)
		previewURL = fmt.Sprintf("%s-%s.yourdomain.com", tag, payload.Repository.Name)
		shouldDeploy = true

	default:
		c.JSON(200, gin.H{"message": fmt.Sprintf("Ignoring event type: %s", eventType)})
		return
	}

	if !shouldDeploy {
		c.JSON(200, gin.H{"message": "No deployment triggered"})
		return
	}

	accessToken, err := db.UserToken(c.Request.Context(), payload.Repository.Owner.Login)
	if err != nil {
		c.JSON(500, gin.H{"error": fmt.Sprintf("Failed to get user token: %v", err)})
		return
	}

	params := schema.DeployParams{
		User:        payload.Repository.Owner.Login,
		AccessToken: accessToken,
		GitURL:      injectTokenIntoGitURL(payload.Repository.CloneURL, accessToken.AccessToken),
		ServiceName: serviceName,
		Tag:         tag,
		Env:         map[string]string{},
	}

	go func() {
		if result := Deploy(c.Request.Context(), params); result.Error != nil {
			fmt.Printf("Deploy failed for %s: %v\n", serviceName, result.Error)
		} else {
			fmt.Printf("Deploy succeeded for %s: %s\n", serviceName, previewURL)
		}
	}()

	response := gin.H{"message": "Deployment started", "service_name": serviceName, "tag": tag}
	if previewURL != "" {
		response["preview_url"] = previewURL
	}
	c.JSON(202, response)
}

func getWebhookSecret() string {
	return ""
}
