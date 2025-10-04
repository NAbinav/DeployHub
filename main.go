package main

import (
	"deployhub/handler"
	"github.com/gin-gonic/gin"
)

func main() {
	r := gin.Default()
	r.GET("/", handler.CloneHandler)
	r.Run(":8080")
}
