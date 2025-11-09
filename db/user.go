package db

import (
	"context"
	"encoding/json"
	"golang.org/x/oauth2"
)

func UserToken(ctx context.Context, username string) (oauth2.Token, error) {
	query := "SELECT token FROM users WHERE username = ?;"
	row, err := ExecuteQuerySingle(ctx, query, username)
	if err != nil {
		return oauth2.Token{}, err
	}
	tokenValue := row["token"].(string)

	var token oauth2.Token
	if err := json.Unmarshal([]byte(tokenValue), &token); err != nil {
		return oauth2.Token{}, err
	}
	return token, nil
}
