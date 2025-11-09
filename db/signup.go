package db

import (
	"context"
	"fmt"
	"github.com/google/uuid"
)

func SignUp(ctx context.Context, username, token, pfpURL string) error {
	id := uuid.NewString()
	query := `
		INSERT INTO users (id, username, token, pfp_link)
		VALUES (?, ?, ?, ?)
		ON CONFLICT(username) DO UPDATE SET
			token = excluded.token,
			pfp_link = excluded.pfp_link;    `
	_, err := ExecuteQuery(ctx, query, id, username, token, pfpURL)

	if err != nil {
		return fmt.Errorf("failed to upsert user: %w", err)
	}

	return nil

}
