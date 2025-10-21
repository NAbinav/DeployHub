package jwt

import (
	"time"

	"github.com/golang-jwt/jwt/v5"
)

var Secret_key = []byte("q2j3y3oovsf4ubeyv2")

func Create_JWT(username string) (string, error) {
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"username": username,
		"exp":      time.Now().Add(24 * time.Hour).Unix(),
	})
	tokenstring, err := token.SignedString(Secret_key)
	if err != nil {
		return "", err
	}
	return tokenstring, nil

}
