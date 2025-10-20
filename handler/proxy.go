package handler

import (
	"fmt"
	"net/http"
	"net/http/httputil"
	"net/url"
	"os"
	"strings"
)

func GetWebsite(w http.ResponseWriter, r *http.Request) {
	if r.Host == "localhost:8080" || r.Host == "yourmaindomain.com" {
		http.NotFound(w, r) // let main server handle /login, /callback, etc.
		return
	}

	projectName := strings.Split(r.Host, ".")[0]
	targetURL := "https://" + projectName + "-" + os.Getenv("GCLOUD_PROJECT_NUMBER") + ".asia-south1.run.app"
	target, err := url.Parse(targetURL)
	if err != nil {
		http.Error(w, "Bad target URL", http.StatusInternalServerError)
		fmt.Println("URL parse error:", err)
		return
	}
	proxy := httputil.NewSingleHostReverseProxy(target)
	r.Host = target.Host
	fmt.Println("Proxying request to:", targetURL)
	proxy.ServeHTTP(w, r)
}
