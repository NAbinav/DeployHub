package handler

import (
	"deployhub/cmd"
	"fmt"
	"strings"

	"github.com/gin-gonic/gin"
)

func CloneHandler(c *gin.Context) {
	url := c.Query("url")
	if url == "" {
		c.String(400, "URL parameter is required")
		return
	}
	if !strings.HasPrefix(url, "https://") && !strings.HasPrefix(url, "git@") {
		c.String(400, "Invalid URL")
		return
	}
	c.Writer.Header().Set("Content-Type", "text/plain")
	c.Writer.Header().Set("Transfer-Encoding", "chunked")
	c.Writer.Flush()
	err := cmd.CloneRepo(url)
	if err != nil {
		fmt.Fprintf(c.Writer, "\nClone failed: %v\n", err)
	} else {
		fmt.Fprintln(c.Writer, "\nClone completed successfully!")
	}
}
