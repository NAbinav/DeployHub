package handler

//
// import (
// 	"context"
// 	"deployhub/db"
// 	"deployhub/jwt"
// 	"fmt"
// 	"time"
//
// 	"github.com/gin-gonic/gin"
// )
//
// func DeployLogsHandler(c *gin.Context) {
// 	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Minute)
// 	defer cancel()
//
// 	deploymentID := c.Param("deployment_id")
// 	if deploymentID == "" {
// 		c.JSON(400, gin.H{"error": "Deployment ID is required"})
// 		return
// 	}
//
// 	token, err := c.Cookie("token")
// 	if err != nil {
// 		c.JSON(401, gin.H{"error": "Invalid token"})
// 		return
// 	}
//
// 	user, err := jwt.Verify_JWT(token)
// 	if err != nil {
// 		c.JSON(401, gin.H{"error": "JWT verification failed"})
// 		return
// 	}
//
// 	// Set SSE headers
// 	c.Writer.Header().Set("Content-Type", "text/event-stream")
// 	c.Writer.Header().Set("Cache-Control", "no-cache")
// 	c.Writer.Header().Set("Connection", "keep-alive")
//
// 	// Stream existing logs
// 	logs, err := db.AddLog(deploymentID, user, ctx)
// 	if err != nil {
// 		c.SSEvent("error", fmt.Sprintf("Failed to retrieve logs: %v", err))
// 		c.Writer.Flush()
// 		return
// 	}
//
// 	for _, log := range logs {
// 		c.SSEvent("message", log)
// 		c.Writer.Flush()
// 	}
//
// 	// Keep the connection open to stream new logs (optional, depends on db implementation)
// 	// If db.GetLogs can watch for new logs, implement here
// 	ticker := time.NewTicker(1 * time.Second)
// 	defer ticker.Stop()
//
// 	for {
// 		select {
// 		case <-ctx.Done():
// 			c.SSEvent("message", "Log streaming ended")
// 			c.Writer.Flush()
// 			return
// 		case <-ticker.C:
// 			// Check for new logs (requires db.WatchLogs or similar)
// 			newLogs, err := db.GetLogs(deploymentID, user, ctx)
// 			if err != nil {
// 				c.SSEvent("error", fmt.Sprintf("Failed to retrieve new logs: %v", err))
// 				c.Writer.Flush()
// 				continue
// 			}
// 			// Send only new logs
// 			for _, log := range newLogs[len(logs):] {
// 				c.SSEvent("message", log)
// 				c.Writer.Flush()
// 			}
// 			logs = newLogs
// 		case <-c.Request.Context().Done():
// 			c.SSEvent("message", "Client disconnected")
// 			c.Writer.Flush()
// 			return
// 		}
// 	}
// }
