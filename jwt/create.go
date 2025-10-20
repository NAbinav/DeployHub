package jwt

import (
	"github.com/golang-jwt/jwt/v5"
	"time"
)

var Secret_key = []byte("q2j3y3oovsf4ubeyv2")

func Create_JWT(email_id string) (string, error) {
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"email_id": email_id,
		"exp":      time.Now().Add(24 * time.Hour).Unix(),
	})
	tokenstring, err := token.SignedString(Secret_key)
	if err != nil {
		return "", err
	}
	return tokenstring, nil

}
