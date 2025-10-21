package auth

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/google/uuid"
	"golang.org/x/oauth2"
	"io"
	"net/http"
)

type Client struct {
	Config *oauth2.Config
}

func (c *Client) CallbackHandler(w http.ResponseWriter, r *http.Request) {
	code := r.URL.Query().Get("code")
	token, err := c.Config.Exchange(context.Background(), code)
	req, _ := http.NewRequest("GET", "https://api.github.com/user/repos", nil)
	req.Header.Set("Authorization", "Bearer "+token.AccessToken)

	resp, err := http.DefaultClient.Do(req)
	// defer resp.Body.Close()
	body, _ := io.ReadAll(resp.Body)
	// fmt.Println(string(body))
	fmt.Println(string(body[0]))

	if err != nil {
		fmt.Println(err)
		return
	}
	client := c.Config.Client(context.Background(), token)
	response, err := client.Get("https://api.github.com/user")
	output, _ := io.ReadAll(response.Body)
	client_json := make(map[string]any)
	json.Unmarshal(output, &client_json)
	fmt.Println(client_json)

}
func (c *Client) LoginHandler(w http.ResponseWriter, r *http.Request) {
	url := c.Config.AuthCodeURL(uuid.NewString())
	http.Redirect(w, r, url, http.StatusTemporaryRedirect)
}
