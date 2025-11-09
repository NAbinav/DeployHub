package helper

import (
	"context"
	"encoding/json"
	"fmt"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/github"
	"io"
)

func GetRepo(tokenString oauth2.Token, username string) ([]map[string]any, error) {

	conf := &oauth2.Config{
		ClientID:     "",
		ClientSecret: "",
		Endpoint:     github.Endpoint,
	}

	client := conf.Client(context.Background(), &tokenString)
	url := "https://api.github.com/user/repos?visibility=all&affiliation=owner&per_page=100"

	res, err := client.Get(url)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()

	body, err := io.ReadAll(res.Body)
	if err != nil {
		return nil, err
	}

	// Check if the response is an error
	if res.StatusCode != 200 {
		var errResp map[string]any
		fmt.Println(err)
		if err := json.Unmarshal(body, &errResp); err != nil {
			return nil, fmt.Errorf("API error (status %d): %s", res.StatusCode, string(body))
		}
		return nil, fmt.Errorf("API error (status %d): %v", res.StatusCode, errResp)
	}

	var repos []map[string]any
	if err := json.Unmarshal(body, &repos); err != nil {
		return nil, fmt.Errorf("failed to parse repos (response: %s): %w", string(body), err)
	}
	fmt.Println(len(repos))

	return repos, nil
}
