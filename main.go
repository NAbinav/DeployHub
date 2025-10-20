package main

import (
	"deployhub/auth"
	"deployhub/handler"
	"fmt"
	"log"
	"net/http"
	"os"

	"golang.org/x/oauth2"
	"golang.org/x/oauth2/github"
)

func main() {
	client := auth.Client{}
	client.Config = &oauth2.Config{
		ClientID:     os.Getenv("GITHUB_CLIENT_ID"),
		ClientSecret: os.Getenv("GITHUB_CLIENT_SECRET"),
		RedirectURL:  "http://localhost:8080/callback",
		Endpoint:     github.Endpoint,
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
