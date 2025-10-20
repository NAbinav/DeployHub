package main

import (
	"deployhub/auth"
	"deployhub/db"
	"deployhub/handler"
	"fmt"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/github"
	"log"
	"net/http"
	"os"
)

func main() {
	err := db.Init_db()
	if err != nil {
		fmt.Println(err)
		return
	}
	client := auth.Client{}
	client.Config = &oauth2.Config{
		ClientID:     os.Getenv("GITHUB_CLIENT_ID"),
		ClientSecret: os.Getenv("GITHUB_CLIENT_SECRET"),
		RedirectURL:  "http://localhost:8080/callback",
		Scopes:       []string{"repo"},

		Endpoint: github.Endpoint,
	}

	http.HandleFunc("GET /", handler.GetWebsite)
	http.HandleFunc("POST /", handler.DeployHandler)
	http.HandleFunc("GET /login", client.LoginHandler)
	http.HandleFunc("GET /callback", client.CallbackHandler)

	server := &http.Server{
		Addr: ":8080",
	}

	fmt.Println("running on :8080")
	log.Fatal(server.ListenAndServe())
}
