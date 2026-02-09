package handler

import (
	"deployhub/db"
	"deployhub/jwt"

	"github.com/gin-gonic/gin"
)

func Rollback(c *gin.Context) {
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

	_, err = db.UserToken(c, user)
	if err != nil {
		c.Error(err)
		return
	}
	pname := c.Param("pname")
	ids, err := db.GetDockerIds(c, pname)

	if err != nil {
		c.Errors.JSON()
		return
	}
	c.JSON(200, ids)
}
