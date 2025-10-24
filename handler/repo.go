package handler

import (
	"deployhub/db"
	"deployhub/helper"
	"deployhub/jwt"
	"fmt"

	"github.com/gin-gonic/gin"
)

type Repo struct {
	Id   float64 `json:"id"`
	Name string  `json:"name"`
	Url  string  `json:"url"`
}

func RepoName(c *gin.Context) {
	jwt_token, err := c.Cookie("token")
	if err != nil {
		c.JSON(400, err)
	}
	username, err := jwt.Verify_JWT(jwt_token)

	if err != nil {
		c.JSON(400, err)
	}
	fmt.Println(username)
	token, err := db.UserToken(username, c)
	if err != nil {
		c.JSON(400, err)
	}
	repos, err := helper.GetRepo(token, username)
	if err != nil {
		fmt.Println(err)
		return
	}
	var all_repo []Repo
	for _, j := range repos {
		var repo Repo
		repo.Id = j["id"].(float64)
		repo.Name = j["name"].(string)
		repo.Url = j["html_url"].(string)
		all_repo = append(all_repo, repo)
	}
	c.JSON(200, all_repo)

}
