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
	if code == "" {
		return fmt.Errorf("no code in callback")
	}

	token, err := c.Config.Exchange(context.Background(), code)
	if err != nil {
		return fmt.Errorf("token exchange failed: %w", err)
	}

	fmt.Printf("Got token: %s (type: %s)\n", token.AccessToken, token.TokenType)

	client := c.Config.Client(context.Background(), token)

	response, err := client.Get("https://api.github.com/user")
	if err != nil {
		return fmt.Errorf("failed to get user info: %w", err)
	}
	defer response.Body.Close()

	if response.StatusCode != 200 {
		body, _ := io.ReadAll(response.Body)
		return fmt.Errorf("GitHub API error (status %d): %s", response.StatusCode, string(body))
	}

	output, err := io.ReadAll(response.Body)
	if err != nil {
		return fmt.Errorf("failed to read response: %w", err)
	}

	client_json := make(map[string]any)
	if err := json.Unmarshal(output, &client_json); err != nil {
		return fmt.Errorf("failed to parse user info: %w", err)
	}

	username, ok := client_json["login"].(string)
	if !ok {
		return fmt.Errorf("invalid username in response")
	}

	avatarURL, _ := client_json["avatar_url"].(string)

	tokenString, err := helper.TokenToString(token)
	if err != nil {
		return fmt.Errorf("failed to serialize token: %w", err)
	}

	fmt.Printf("Storing token for user %s: %s\n", username, tokenString)

	err = db.SignUp(ctx, username, tokenString, avatarURL)
	if err != nil {
		return fmt.Errorf("failed to store user: %w", err)
	}

	jwt_token, err := jwt.Create_JWT(username)
	if err != nil {
		return fmt.Errorf("failed to create JWT: %w", err)
	}

	ctx.SetCookie("token", jwt_token, 7200, "/", "brogramiz.info", true, true)

	ctx.Redirect(http.StatusTemporaryRedirect, "/")
	return nil
}

func (c *Client) LoginHandler(w http.ResponseWriter, r *http.Request) {
	url := c.Config.AuthCodeURL(uuid.NewString())
	http.Redirect(w, r, url, http.StatusTemporaryRedirect)
}
