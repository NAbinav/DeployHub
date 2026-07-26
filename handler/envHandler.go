package handler

import (
	"deployhub/db"
	"deployhub/jwt"
	"encoding/json"
	"net/http"

	"github.com/gin-gonic/gin"
)

func GetEnvHandler(c *gin.Context) {
	pname := c.Param("pname")

	token, err := c.Cookie("token")
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "no token"})
		return
	}
	_, err = jwt.Verify_JWT(token)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid token"})
		return
	}

	results, err := db.GetProjectEnv(c, pname)
	if err != nil || len(results) == 0 {
		c.JSON(404, gin.H{"error": "project not found"})
		return
	}

	envStr, ok := results[0]["env"].(string)
	if !ok || envStr == "" {
		c.JSON(200, map[string]string{})
		return
	}

	var envMap map[string]string
	if err := json.Unmarshal([]byte(envStr), &envMap); err != nil {
		c.JSON(500, gin.H{"error": "failed to parse env vars"})
		return
	}

	c.JSON(200, envMap)
}

func UpdateEnvHandler(c *gin.Context) {
	pname := c.Param("pname")

	token, err := c.Cookie("token")
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "no token"})
		return
	}
	_, err = jwt.Verify_JWT(token)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid token"})
		return
	}

	var newEnv map[string]string
	if err := c.BindJSON(&newEnv); err != nil {
		c.JSON(400, gin.H{"error": "invalid env data"})
		return
	}

	envBytes, err := json.Marshal(newEnv)
	if err != nil {
		c.JSON(500, gin.H{"error": "failed to encode env vars"})
		return
	}

	err = db.UpdateProjectEnv(c, pname, string(envBytes))
	if err != nil {
		c.JSON(500, gin.H{"error": "failed to save env vars"})
		return
	}

	c.JSON(200, gin.H{"status": "updated"})
}
