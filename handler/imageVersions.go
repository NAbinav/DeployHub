package handler

import (
	"deployhub/db"
	"deployhub/jwt"
	"fmt"

	"github.com/gin-gonic/gin"
)

func ImageIds(c *gin.Context) {
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
	fmt.Println(pname)
	ids, err := db.GetDockerIds(c, pname)
	fmt.Println(ids)

	if err != nil {
		c.Errors.JSON()
		return
	}
	c.JSON(200, ids)
}
