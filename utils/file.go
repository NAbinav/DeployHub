package utils

import (
	"os"
	"path/filepath"
	"strings"
)

func fileExists(path string) bool {
	_, err := os.Stat(path)
	return !os.IsNotExist(err)
}

func DetectFramework(projectPath string) string {
	if fileExists(filepath.Join(projectPath, "package.json")) {
		data, _ := os.ReadFile(filepath.Join(projectPath, "package.json"))
		str := string(data)
		if strings.Contains(str, "next") {
			return "nextjs"
		}
		if strings.Contains(str, "vite") {
			return "vite"
		}
		if strings.Contains(str, "react") {
			return "react"
		}
		if strings.Contains(str, "express") {
			return "express"
		}

		return "node"
	}
	if fileExists(filepath.Join(projectPath, "requirements.txt")) {
		data, _ := os.ReadFile(filepath.Join(projectPath, "requirements.txt"))
		str := string(data)
		if strings.Contains(str, "fastapi") {
			return "fastapi"
		}
		if strings.Contains(str, "flask") {
			return "flask"
		}
		return "python"
	}
	if fileExists(filepath.Join(projectPath, "go.mod")) {
		return "go"
	}
	if fileExists(filepath.Join(projectPath, "index.html")) {
		return "static"
	}
	return "unknown"
}
