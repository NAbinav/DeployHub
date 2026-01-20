package utils

import "strings"

func ENVString(env map[string]string) string {
	var output strings.Builder
	for key, val := range env {
		output.WriteString(key)
		output.WriteString("=")
		output.WriteString(val)
		output.WriteString("\n")
	}
	return output.String()
}
