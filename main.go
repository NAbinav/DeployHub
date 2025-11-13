package main

import (
	"deployhub/auth"
	"deployhub/db"
	"deployhub/handler"
	"deployhub/jwt"
	"fmt"
	"github.com/gin-gonic/gin"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/github"
	"log"
	"net/http"
	"os"
)

func allowedHostsMiddleware(allowedHosts ...string) gin.HandlerFunc {
	hostsMap := make(map[string]bool)
	for _, host := range allowedHosts {
		hostsMap[host] = true
	}

	return func(c *gin.Context) {
		if !hostsMap[c.Request.Host] {
			handler.GetWebsite(c)
			c.Abort()
			return
		}
		c.Next()
	}
}

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

	// Apply middleware to API routes group
	api := r.Group("/api", allowedHostsMiddleware("localhost:8080", "brogramiz.info"))
	{
		api.GET("/login", func(c *gin.Context) {
			client.LoginHandler(c.Writer, c.Request)
		})

		api.GET("/signup", func(c *gin.Context) {
			client.LoginHandler(c.Writer, c.Request)
		})

		api.GET("/callback", func(c *gin.Context) {
			client.CallbackHandler(c.Writer, c.Request, c)
		})

		api.POST("/deploy/", handler.DeployHandler)

		api.GET("/repo", handler.RepoName)

		api.GET("/deploy/:id", handler.DeployIDHandler)

		api.GET("/check", func(c *gin.Context) {
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
		})

		api.GET("/projects", handler.GetProject)

		api.DELETE("/projects", handler.DeleteService)
	}

	r.NoRoute(func(c *gin.Context) {
		handler.GetWebsite(c)
	})

	fmt.Println("running on :8080")
	log.Fatal(r.Run(":8080"))
}
