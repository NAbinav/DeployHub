package handler

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"strings"

	"deployhub/db"
	"deployhub/deploy"
	"deployhub/schema"

	"github.com/gin-gonic/gin"
)

type GitHubWebhookPayload struct {
	Ref        string `json:"ref"`
	After      string `json:"after"`
	Repository struct {
		Name     string `json:"name"`
		CloneURL string `json:"clone_url"`
		Owner    struct {
			Login string `json:"login"`
		} `json:"owner"`
	} `json:"repository"`
	PullRequest *struct {
		Number int    `json:"number"`
		State  string `json:"state"`
		Head   struct {
			Ref string `json:"ref"`
			SHA string `json:"sha"`
		} `json:"head"`
	} `json:"pull_request,omitempty"`
}

/* -------------------- helpers -------------------- */

func sanitizeTag(s string) string {
	s = strings.ToLower(s)
	s = strings.ReplaceAll(s, "/", "-")
	s = strings.ReplaceAll(s, "_", "-")

	var out strings.Builder
	for _, r := range s {
		if (r >= 'a' && r <= 'z') || (r >= '0' && r <= '9') || r == '-' {
			out.WriteRune(r)
		}
	}
	return out.String()
}

func verifySignature(payload []byte, signature, secret string) bool {
	if secret == "" {
		return true // allow if secret not configured
	}

	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write(payload)
	expected := "sha256=" + hex.EncodeToString(mac.Sum(nil))

	return hmac.Equal([]byte(expected), []byte(signature))
}

func injectGitHubToken(cloneURL, token string) string {
	if strings.HasPrefix(cloneURL, "https://") {
		return strings.Replace(
			cloneURL,
			"https://",
			fmt.Sprintf("https://x-access-token:%s@", token),
			1,
		)
	}
	return cloneURL
}

func getWebhookSecret() string {
	// TODO: load from env
	return ""
}

/* -------------------- handler -------------------- */

func HandleGitHubWebhook(c *gin.Context) {
	serviceName := c.Param("id")

	body, err := io.ReadAll(c.Request.Body)
	if err != nil {
		c.JSON(400, gin.H{"error": "unable to read body"})
		return
	}

	if !verifySignature(
		body,
		c.GetHeader("X-Hub-Signature-256"),
		getWebhookSecret(),
	) {
		c.JSON(401, gin.H{"error": "invalid signature"})
		return
	}

	var payload GitHubWebhookPayload
	if err := json.Unmarshal(body, &payload); err != nil {
		c.JSON(400, gin.H{"error": "invalid json payload"})
		return
	}

	event := c.GetHeader("X-GitHub-Event")

	var (
		tag        string
		previewURL string
	)

	switch event {

	case "push":
		if !strings.HasPrefix(payload.Ref, "refs/heads/") {
			c.JSON(200, gin.H{"message": "ignored non-branch push"})
			return
		}

		branch := strings.TrimPrefix(payload.Ref, "refs/heads/")
		if branch == "main" || branch == "master" {
			tag = "latest"
		} else {
			tag = sanitizeTag(branch)
			previewURL = fmt.Sprintf(
				"%s-%s.yourdomain.com",
				tag,
				payload.Repository.Name,
			)
		}

	case "pull_request":
		if payload.PullRequest == nil || payload.PullRequest.State != "open" {
			c.JSON(200, gin.H{"message": "ignored closed pr"})
			return
		}

		tag = fmt.Sprintf("pr-%d", payload.PullRequest.Number)
		previewURL = fmt.Sprintf(
			"%s-%s.yourdomain.com",
			tag,
			payload.Repository.Name,
		)

	default:
		c.JSON(200, gin.H{"message": "event ignored"})
		return
	}

	accessToken, err := db.UserToken(
		context.Background(),
		payload.Repository.Owner.Login,
	)
	if err != nil {
		c.JSON(500, gin.H{"error": "failed to fetch user token"})
		return
	}

	deployParams := schema.DeployParams{
		User:        payload.Repository.Owner.Login,
		AccessToken: accessToken,
		GitURL:      payload.Repository.Name,
		ServiceName: serviceName,
		Tag:         tag,
		Env:         map[string]string{},
	}

	go func() {
		ctx := context.Background()
		result := deploy.Deploy(ctx, deployParams)

		if result.Error != nil {
			log.Printf(
				"[DEPLOY FAILED] service=%s tag=%s err=%v",
				serviceName,
				tag,
				result.Error,
			)
			return
		}

		log.Printf(
			"[DEPLOY OK] service=%s tag=%s url=%s",
			serviceName,
			tag,
			previewURL,
		)
	}()

	resp := gin.H{
		"message": "deployment started",
		"service": serviceName,
		"tag":     tag,
	}

	if previewURL != "" {
		resp["preview_url"] = previewURL
	}

	c.JSON(202, resp)
}
