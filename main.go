package main

import (
	"deployhub/handler"
	"fmt"
	"log"
	"net/http"
)

func main() {
	http.HandleFunc("/", handler.GetWebsite)
	http.HandleFunc("/deploy", handler.DeployHandler)

	server := &http.Server{
		Addr: ":8080",
	}

	fmt.Println("running on :8080")
	log.Fatal(server.ListenAndServe())
}
