package handler

import (
	"deployhub/db"
	"deployhub/jwt"
	"fmt"

	"github.com/gin-gonic/gin"
)

func GetProject(c *gin.Context) {
	token, err := c.Cookie("token")
	fmt.Println(token)
	if err != nil {
		c.JSON(400, err)
		fmt.Println(err)
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
