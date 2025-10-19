package utils

func ENVString(env map[string]string) string {
	var output string
	for key, val := range env {
		output += key + "=" + val
	}
	return output
}
