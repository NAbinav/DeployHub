package handler

import (
	"context"
	"deployhub/deploy"
	"deployhub/schema"
	"fmt"
	"os"

	"github.com/gin-gonic/gin"
)

func HandleRollback(c *gin.Context) {
	serviceName := c.Param("pname")
	var body struct {
		ImageID string `json:"image_id"`
	}

	if err := c.ShouldBindJSON(&body); err != nil || body.ImageID == "" {
		c.JSON(400, gin.H{"error": "valid image_id (sha256 digest) is required"})
		return
	}

	projectID := os.Getenv("GCLOUD_PROJECT_ID")
	region := "asia-south1"
	repoID := "deploy-hub"

	// Deploy the old digest directly — no retag needed
	imagePath := fmt.Sprintf("%s-docker.pkg.dev/%s/%s/%s-image", region, projectID, repoID, serviceName)
	sourceImage := fmt.Sprintf("%s@%s", imagePath, body.ImageID)

	fmt.Printf("🛠 Rolling back to digest: %s\n", body.ImageID)

	ctx := context.Background()
	config := &schema.DeploymentConfig{
		ProjectID: projectID,
		Region:    region,
		ImagePath: sourceImage,
	}

	params := schema.DeployParams{
		ServiceName: serviceName,
		Tag:         "latest",
		Env:         map[string]string{},
	}

	if err := deploy.DeployToCloudRun(ctx, config, params); err != nil {
		c.JSON(500, gin.H{"error": fmt.Sprintf("Cloud Run update failed: %v", err)})
		return
	}

	c.JSON(200, gin.H{
		"message":      "Rollback successful",
		"active_image": body.ImageID,
	})
}
