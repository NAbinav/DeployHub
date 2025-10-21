package db

import (
	"context"

	"github.com/google/uuid"
)

func SignUp(username any, token string, pfp_url any) error {
	id := uuid.NewString()
	query := `INSERT INTO users (id,username, token, pfp_link) SELECT $1, $2, $3,$4
	WHERE NOT EXISTS (SELECT 1 FROM users WHERE username = $2);`
	_, err := DB.Exec(context.Background(), query, id, username, token, pfp_url)
	if err != nil {
		return err
	}
	return nil
}
