package handler

import (
	"deployhub/db"
	"deployhub/jwt"
	"deployhub/schema"
	"deployhub/worker"
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
)

func DeployHandler(c *gin.Context) {
	var req schema.DeployRequest
	if err := c.BindJSON(&req); err != nil {
		c.Error(err)
		return
	}

	if db.ProjectExists(c.Request.Context(), req.ServiceName) {
		fmt.Println("ALREADY EXISTS")
		c.AbortWithError(http.StatusConflict, fmt.Errorf("Sorry this name already exists"))
		return
	}
	token, err := c.Cookie("token")
	if err != nil {
		c.Error(err)
		return
	}

	user, err := jwt.Verify_JWT(token)
	if err != nil {
		c.Error(err)
		return
	}

	accessToken, err := db.UserToken(c, user)
	if err != nil {
		c.Error(err)
		return
	}

	params := schema.DeployParams{
		User:        user,
		AccessToken: accessToken,
		GitURL:      req.GitURL,
		ServiceName: req.ServiceName,
		Env:         req.Env,
		Tag:         "latest",
	}

	DeploymentID, err := worker.Worker(c.Request.Context(), params)
	if err != nil {
		c.Error(err)
		return
	}

	c.JSON(200, schema.DeployResponse{
		DeploymentID: DeploymentID,
	})
}

// Deploy is the main deployment function that can be called from handlers or webhooks
