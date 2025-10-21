package helper

import (
	"encoding/json"
	"golang.org/x/oauth2"
)

func TokenToString(t *oauth2.Token) (string, error) {
	jsonBytes, err := json.Marshal(t)
	if err != nil {
		return "", err
	}
	return string(jsonBytes), nil
}
