package db

import "context"

func SignUp(username string, token string, pfp_url string) {
	query := "insert into users values ($1,$2,#3);"
	DB.Exec(context.Background(), query)
}
