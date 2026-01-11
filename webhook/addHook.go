package webhook

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"golang.org/x/oauth2"
	"golang.org/x/oauth2/github"
)

func AddWebhook(tokenString oauth2.Token, owner, repo, webhookURL string) error {
	conf := &oauth2.Config{
		ClientID:     "",
		ClientSecret: "",
		Endpoint:     github.Endpoint,
	}
	client := conf.Client(context.Background(), &tokenString)

	payload := map[string]any{
		"name":   "web",
		"active": true,
		"events": []string{"push"},
		"config": map[string]string{
			"url":          webhookURL,
			"content_type": "json",
		},
	}

	body, _ := json.Marshal(payload)
	url := fmt.Sprintf("https://api.github.com/repos/%s/%s/hooks", owner, repo)
	req, _ := http.NewRequest("POST", url, bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")

	res, err := client.Do(req)
	if err != nil {
		return err
	}
	defer res.Body.Close()

	if res.StatusCode != 201 {
		return fmt.Errorf("failed: status %d", res.StatusCode)
	}
	return nil
}
