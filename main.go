package main

import (
	"deployhub/auth"
	"deployhub/db"
	"deployhub/handler"
	"fmt"
	"github.com/gin-gonic/gin"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/github"
	"log"
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
		RedirectURL:  os.Getenv("CALLBACK_URL"),
		Scopes:       []string{"repo"},
		Endpoint:     github.Endpoint,
	}

	r := gin.Default()

	r.GET("/login", func(c *gin.Context) {
		if c.Request.Host == "localhost:8080" || c.Request.Host == "brogramiz.info" {
			client.LoginHandler(c.Writer, c.Request)
		} else {
			handler.GetWebsite(c)
		}
	})

	r.GET("/signup", func(c *gin.Context) {
		if c.Request.Host == "localhost:8080" || c.Request.Host == "brogramiz.info" {
			client.LoginHandler(c.Writer, c.Request)
		} else {
			handler.GetWebsite(c)
		}
	})

	r.GET("/callback", func(c *gin.Context) {
		if c.Request.Host == "localhost:8080" || c.Request.Host == "brogramiz.info" {
			client.CallbackHandler(c.Writer, c.Request, c)
		} else {
			handler.GetWebsite(c)
		}
	})

	r.POST("/", func(c *gin.Context) {
		fmt.Println(c.Request.Host)
		if c.Request.Host == "localhost:8080" || c.Request.Host == "brogramiz.info" {
			handler.DeployHandler(c)
		} else {
			handler.GetWebsite(c)
		}
	})

	r.GET("/repo", handler.RepoName)
	r.GET("/:id", func(c *gin.Context) {
		fmt.Println(c.Request.Host)
		if c.Request.Host == "localhost:8080" || c.Request.Host == "brogramiz.info" {
			handler.DeployIDHandler(c)
		} else {
			handler.GetWebsite(c)
		}
	})
	r.NoRoute(func(c *gin.Context) {
		handler.GetWebsite(c)
	})

	fmt.Println("running on :8080")
	log.Fatal(r.Run(":8080"))
}
