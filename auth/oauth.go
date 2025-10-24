package auth

import (
	"context"
	"deployhub/db"
	"deployhub/helper"
	"deployhub/jwt"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"golang.org/x/oauth2"
)

type Client struct {
	Config *oauth2.Config
}

func (c *Client) CallbackHandler(w http.ResponseWriter, r *http.Request, ctx *gin.Context) error {
	code := r.URL.Query().Get("code")
	token, err := c.Config.Exchange(context.Background(), code)
	client := c.Config.Client(context.Background(), token)
	response, err := client.Get("https://api.github.com/user")
	output, _ := io.ReadAll(response.Body)
	client_json := make(map[string]any)
	json.Unmarshal(output, &client_json)
	tokenString, err := helper.TokenToString(token)
	if err != nil {
		return err
	}
	err = db.SignUp(client_json["login"], tokenString, client_json["avatar_url"])
	jwt_token, err := jwt.Create_JWT(client_json["login"].(string))
	if err != nil {
		fmt.Println(err)
		return err
	}
	ctx.SetCookie("token", jwt_token, 7200, "/", "brogramiz.info", true, true)
	// ctx.JSON(http.StatusOK, gin.H{
	// 	"status":   "success",
	// 	"redirect": "http://localhost:5173",
	// })
	ctx.Redirect(307, "/")

	return nil

}
func (c *Client) LoginHandler(w http.ResponseWriter, r *http.Request) {
	url := c.Config.AuthCodeURL(uuid.NewString())
	http.Redirect(w, r, url, http.StatusTemporaryRedirect)
}
