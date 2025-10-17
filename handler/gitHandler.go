package handler

import (
	"deployhub/cmd"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

func CloneHandler(c *gin.Context) {
	url := c.Query("url")
	fmt.Println(c.Request.Host)
	if url == "" {
		c.String(400, "URL parameter is required")
		return
	}
	var TempFolderName = uuid.NewString()
	err := cmd.CloneRepo(url, TempFolderName)
	if err != nil {
		fmt.Fprintf(c.Writer, "\nClone failed: %v\n", err)
	} else {
		fmt.Fprintln(c.Writer, "\nClone completed successfully!")
	}
}
