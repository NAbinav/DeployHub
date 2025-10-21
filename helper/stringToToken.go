package helper

import "encoding/json"

func StringToToken(token string) (map[string]any, error) {
	var map_token map[string]any
	if err := json.Unmarshal([]byte(token), &map_token); err != nil {
		return map[string]any{}, err
	}
	return map_token, nil
}
