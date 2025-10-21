package helper

import (
	"context"
	"encoding/json"
	"fmt"
	"io"

	"golang.org/x/oauth2"
	"golang.org/x/oauth2/github"
)

func GetRepo(tokenString string, username string) ([]map[string]any, error) {
	var token oauth2.Token
	if err := json.Unmarshal([]byte(tokenString), &token); err != nil {
		return nil, fmt.Errorf("failed to parse token: %w", err)
	}
	conf := &oauth2.Config{
		ClientID:     "",
		ClientSecret: "",
		Endpoint:     github.Endpoint,
	}
	client := conf.Client(context.Background(), &token)
	url := "https://api.github.com/user/repos?visibility=all&affiliation=owner"
	res, err := client.Get(url)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()
	body, err := io.ReadAll(res.Body)
	if err != nil {
		return nil, err
	}
	var repos []map[string]any
	if err := json.Unmarshal(body, &repos); err != nil {
		return nil, err
	}
	return repos, nil
}
