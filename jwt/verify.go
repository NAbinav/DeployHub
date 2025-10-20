package jwt

import (
	"fmt"
	"github.com/golang-jwt/jwt/v5"
)

func Verify_JWT(token_string string) (string, error) {
	token, err := jwt.Parse(token_string, func(t *jwt.Token) (any, error) {
		return Secret_key, nil
	})
	if err != nil {
		return "", err
	}
	if !token.Valid {
		return "", fmt.Errorf("Invalid Token")
	}
	return (token.Claims.(jwt.MapClaims))["email_id"].(string), nil
}
