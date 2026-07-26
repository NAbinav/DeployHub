package jwt

import (
	"fmt"
	"os"

	"github.com/golang-jwt/jwt/v5"
)

func Verify_JWT(token_string string) (string, error) {
	token, err := jwt.Parse(token_string, func(t *jwt.Token) (any, error) {
		return []byte(os.Getenv("JWT_SECRET")), nil
	})
	if err != nil {
		return "", err
	}
	if !token.Valid {
		return "", fmt.Errorf("Invalid Token")
	}
	return (token.Claims.(jwt.MapClaims))["username"].(string), nil
}
