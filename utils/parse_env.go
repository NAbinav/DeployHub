package utils

import "strings"

func ParseENVString(envStr string) map[string]string {
	env := make(map[string]string)

	lines := strings.SplitSeq(envStr, "\n")
	for line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}

		parts := strings.SplitN(line, "=", 2)
		if len(parts) == 2 {
			env[parts[0]] = parts[1]
		} else if len(parts) == 1 {
			env[parts[0]] = ""
		}
	}

	return env
}
