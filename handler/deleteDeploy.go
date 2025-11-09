package handler

import (
	"deployhub/helper"
	"fmt"

	"github.com/gin-gonic/gin"
)

func DeleteService(c *gin.Context) {
	name := c.Query("name")
	err := helper.DeleteDeploy(c, name)
	if err != nil {
		fmt.Println(err)
		c.Error(err)
		return
	}
	c.JSON(200, struct{ status string }{status: "sucess delete"})
}
