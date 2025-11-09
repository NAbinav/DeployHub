package main

import (
	"deployhub/auth"
	"deployhub/db"
	"deployhub/handler"
	"deployhub/jwt"
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/gin-gonic/gin"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/github"
)

func main() {
	err := db.Init_Cloudflare()
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

	r.GET("/api/login", func(c *gin.Context) {
		if c.Request.Host == "localhost:8080" || c.Request.Host == "brogramiz.info" {
			client.LoginHandler(c.Writer, c.Request)
		} else {
			handler.GetWebsite(c)
		}
	})
	r.GET("/api/signup", func(c *gin.Context) {
		if c.Request.Host == "localhost:8080" || c.Request.Host == "brogramiz.info" {
			client.LoginHandler(c.Writer, c.Request)
		} else {
			handler.GetWebsite(c)
		}
	})
	r.GET("/api/callback", func(c *gin.Context) {
		if c.Request.Host == "localhost:8080" || c.Request.Host == "brogramiz.info" {
			client.CallbackHandler(c.Writer, c.Request, c)
		} else {
			handler.GetWebsite(c)
		}
	})
	r.POST("/api/deploy/", func(c *gin.Context) { // Changed from "/"
		fmt.Println(c.Request.Host)
		if c.Request.Host == "localhost:8080" || c.Request.Host == "brogramiz.info" {
			handler.DeployHandler(c)
		} else {
			handler.GetWebsite(c)
		}
	})
	r.GET("/api/repo", func(c *gin.Context) {
		if c.Request.Host == "localhost:8080" || c.Request.Host == "brogramiz.info" {
			handler.RepoName(c)
		} else {
			handler.GetWebsite(c)
		}
	})
	r.GET("/api/deploy/:id", func(c *gin.Context) { // Changed from "/:id"
		fmt.Println(c.Request.Host)
		if c.Request.Host == "localhost:8080" || c.Request.Host == "brogramiz.info" {
			handler.DeployIDHandler(c)
		} else {
			handler.GetWebsite(c)
		}
	})
	r.GET("/api/check", func(c *gin.Context) {
		if c.Request.Host == "localhost:8080" || c.Request.Host == "brogramiz.info" {
			token, err := c.Cookie("token")
			if err != nil {
				c.JSON(http.StatusUnauthorized, gin.H{"status": "no token"})
				return
			}
			username, err := jwt.Verify_JWT(token)
			if err != nil {
				c.JSON(http.StatusUnauthorized, gin.H{"status": "invalid token"})
				return
			}
			c.JSON(http.StatusOK, gin.H{"status": "success", "user": username})
		} else {
			handler.GetWebsite(c)
		}
	})
	r.GET("/api/projects", func(c *gin.Context) {
		fmt.Println(c.Request.Host)
		if c.Request.Host == "localhost:8080" || c.Request.Host == "brogramiz.info" {
			handler.GetProject(c)
		} else {
			handler.GetWebsite(c)
		}
	})
	r.DELETE("/api/projects", func(c *gin.Context) {
		if c.Request.Host == "localhost:8080" || c.Request.Host == "brogramiz.info" {
			handler.DeleteService(c)
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
