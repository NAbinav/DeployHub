package handler

import (
	"fmt"
	"net/http"
	"net/http/httputil"
	"net/url"
	"os"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
)

func GetWebsite(c *gin.Context) {
	fmt.Println(c.Request)
	if c.Request.Host == "localhost:8080" || c.Request.Host == "brogramiz.info" {
		t, err := c.Cookie("token")
		fmt.Println(t)
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
	transport := &http.Transport{
		MaxIdleConns:          1000,
		MaxIdleConnsPerHost:   100,
		IdleConnTimeout:       90 * time.Second,
		TLSHandshakeTimeout:   5 * time.Second,
		ExpectContinueTimeout: 1 * time.Second,
		ForceAttemptHTTP2:     true,
	}

	// proxy := httputil.NewSingleHostReverseProxy(target)
	proxy := &httputil.ReverseProxy{
		Director: func(req *http.Request) {
			req.URL.Scheme = target.Scheme
			req.URL.Host = target.Host
			req.Host = target.Host

			// Preserve original path & query
			// ReverseProxy does this automatically

			// Forward headers
			req.Header.Set("X-Forwarded-Host", c.Request.Host)
			req.Header.Set("X-Forwarded-Proto", "https")
		},
		Transport:     transport,
		FlushInterval: -1, // streaming, no buffering
	}

	c.Request.Host = target.Host
	fmt.Println("Proxying request to:", targetURL)
	proxy.ServeHTTP(c.Writer, c.Request)
}
