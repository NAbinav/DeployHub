package handler

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
)

func WebhookHandler(c *gin.Context) {
	id := c.Param("id")

	var payload map[string]any

	if err := c.ShouldBindJSON(&payload); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "invalid JSON payload",
		})
		return
	}

	fmt.Println("Webhook ID:", id)
	fmt.Println("Payload:", payload)

	c.JSON(http.StatusOK, gin.H{
		"status": "received",
	})
}
