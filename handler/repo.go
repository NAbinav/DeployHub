package handler

import (
	"deployhub/db"
	"deployhub/helper"
	"deployhub/jwt"
	"fmt"
	"strings"

	"github.com/gin-gonic/gin"
)

func RepoName(c *gin.Context) {
	jwt_token, err := c.Cookie("token")
	if err != nil {
		c.AbortWithError(400, err)
	}
	username, err := jwt.Verify_JWT(jwt_token)

	if err != nil {
		c.AbortWithError(400, err)
	}
	fmt.Println(username)
	token, err := db.UserToken(username, c)
	if err != nil {
		c.AbortWithError(400, err)
	}
	repos, err := helper.GetRepo(token, username)
	if err != nil {
		fmt.Println(err)
		return
	}
	var private_repo []string
	var public_repo []string
	for _, j := range repos {
		repo_name := strings.Split(j["html_url"].(string), "/")
		if j["private"] == true {
			private_repo = append(private_repo, repo_name[len(repo_name)-1])
		} else {

			public_repo = append(public_repo, repo_name[len(repo_name)-1])
		}
	}
	c.String(200, (fmt.Sprintf("public: %v \nprivate: %v\n\nTO START DEPLOY YOUR PROJECT GO TO /<PROJECT_NAME>", public_repo, private_repo)))
	fmt.Println(private_repo, public_repo)
}
