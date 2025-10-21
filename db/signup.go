package db

import (
	"context"
	"fmt"

	"github.com/google/uuid"
)

func SignUp(username any, token string, pfp_url any) error {
	id := uuid.NewString()
	query := `INSERT INTO users (id, username, token, pfp_link)
VALUES ($1, $2, $3, $4)
ON CONFLICT (username) 
DO UPDATE SET token = EXCLUDED.token;
`
	fmt.Println(token)
	_, err := DB.Exec(context.Background(), query, id, username, token, pfp_url)
	if err != nil {
		return err
	}
	return nil
}
