package handler

import (
	"deployhub/db"
	"deployhub/jwt"

	"github.com/gin-gonic/gin"
)

func GetProject(c *gin.Context) {
	token, err := c.Cookie("token")
	if err != nil {
		c.JSON(400, err)
		return
	}
	user, err := jwt.Verify_JWT(token)
	projects, err := db.GetUserProject(c, user)
	if err != nil {
		c.JSON(400, err)
		return
	}
	c.JSON(200, projects)
}
