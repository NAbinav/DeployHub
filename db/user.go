package db

import (
	"context"
)

func UserToken(username string, ctx context.Context) (string, error) {
	query := "select token from users where username=$1;"
	var token string
	err := DB.QueryRow(ctx, query, username).Scan(&token)
	if err != nil {
		return "", err
	}
	return token, nil
}
