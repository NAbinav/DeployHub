package helper

import (
	"fmt"
	"regexp"
	"strings"
	"time"
)

func SanitizeTag(s string) string {
	s = strings.ToLower(s)
	// Replace invalid chars with hyphen
	reg := regexp.MustCompile("[^a-z0-9]+")
	s = reg.ReplaceAllString(s, "-")
	// Trim hyphens
	s = strings.Trim(s, "-")
	// Limit length to 63 chars (DNS limit)
	if len(s) > 63 {
		s = s[:63]
	}
	// Fallback if empty
	if s == "" {
		s = "v" + fmt.Sprint(time.Now().Unix())
	}
	return s
}
