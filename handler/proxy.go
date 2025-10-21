package handler

import (
	"fmt"
	"net/http"
	"net/http/httputil"
	"net/url"
	"os"
	"strings"

	"github.com/gin-gonic/gin"
)

func GetWebsite(c *gin.Context) {
	fmt.Println(c.Request.Host)
	if c.Request.Host == "localhost:8080" || c.Request.Host == "brogramiz.info" {
		_, err := c.Cookie("token")
		if err != nil {
			http.Redirect(c.Writer, c.Request, "/login", 307)
		} else {
			RepoName(c)
		}
		return
	}

	projectName := strings.Split(c.Request.Host, ".")[0]
	targetURL := "https://" + projectName + "-" + os.Getenv("GCLOUD_PROJECT_NUMBER") + ".asia-south1.run.app"
	target, err := url.Parse(targetURL)
	if err != nil {
		c.Error(fmt.Errorf("not found"))
		fmt.Println("URL parse error:", err)
		return
	}
	proxy := httputil.NewSingleHostReverseProxy(target)
	c.Request.Host = target.Host
	fmt.Println("Proxying request to:", targetURL)
	proxy.ServeHTTP(c.Writer, c.Request)
}
