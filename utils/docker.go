package utils

import (
	"fmt"
	"os"
	"path/filepath"
)

func CreateDockerfile(projectPath, framework string) error {
	var dockerfile string

	switch framework {
	case "nextjs":
		dockerfile = `FROM node:20-alpine
WORKDIR /app
COPY package*.json ./
RUN if [ -f package-lock.json ]; then npm ci --legacy-peer-deps; else npm install --legacy-peer-deps; fi
COPY . .
RUN npm run build
ENV PORT=8080
EXPOSE 8080
CMD ["npx", "next", "start", "-p", "8080"]`

	case "vite", "react":
		buildDir := "dist"
		if framework == "react" {
			buildDir = "build"
		}
		dockerfile = fmt.Sprintf(`FROM node:20-alpine AS builder
WORKDIR /app
COPY package*.json ./
RUN if [ -f package-lock.json ]; then npm ci --legacy-peer-deps; else npm install --legacy-peer-deps; fi
COPY . .
RUN npm run build

FROM node:20-alpine
WORKDIR /app
RUN npm install -g serve
COPY --from=builder /app/%s ./dist
ENV PORT=8080
EXPOSE 8080
CMD ["serve", "-s", "dist", "-l", "8080"]`, buildDir)

	case "express":
		dockerfile = `FROM node:20-alpine
WORKDIR /app
COPY package*.json ./
RUN if [ -f package-lock.json ]; then npm ci --legacy-peer-deps; else npm install --legacy-peer-deps; fi
COPY . .
ENV PORT=8080
EXPOSE 8080
CMD ["node", "index.js"]
`
	case "node":
		dockerfile = `FROM node:20-alpine
WORKDIR /app
COPY package*.json ./
RUN if [ -f package-lock.json ]; then npm ci --legacy-peer-deps; else npm install --legacy-peer-deps; fi
COPY . .
ENV PORT=8080
EXPOSE 8080
CMD ["npm", "start"]`

	case "flask":
		dockerfile = `FROM python:3.11-slim
WORKDIR /app
COPY requirements.txt .
RUN pip install --no-cache-dir -r requirements.txt
COPY . .
ENV PORT=8080
ENV FLASK_RUN_HOST=0.0.0.0
ENV FLASK_RUN_PORT=8080
EXPOSE 8080
CMD ["flask", "run"]`

	case "fastapi":
		dockerfile = `FROM python:3.11-slim
WORKDIR /app
COPY requirements.txt .
RUN pip install --no-cache-dir -r requirements.txt
COPY . .
ENV PORT=8080
EXPOSE 8080
CMD ["uvicorn", "main:app", "--host", "0.0.0.0", "--port", "8080"]`

	case "go":
		dockerfile = `FROM golang:1.25-alpine AS builder
WORKDIR /app
COPY go.* ./
RUN go mod download
COPY . .
RUN go build -o server .

FROM alpine:latest
WORKDIR /app
COPY --from=builder /app/server .
ENV PORT=8080
EXPOSE 8080
CMD ["./server"]`

	case "static":
		dockerfile = `FROM node:20-alpine
WORKDIR /app
COPY . .
RUN npm install -g serve
ENV PORT=8080
EXPOSE 8080
CMD ["serve", "-s", ".", "-l", "8080"]`

	default:
		dockerfile = `FROM node:20-alpine
WORKDIR /app
COPY . .
ENV PORT=8080
EXPOSE 8080
CMD ["sh", "-c", "if [ -f package.json ]; then npm install && npm start; else python -m http.server 8080; fi"]`
	}
	dockerfilePath := filepath.Join(projectPath, "Dockerfile")
	return os.WriteFile(dockerfilePath, []byte(dockerfile), 0644)
}
func CreateDockerignore(projectPath string) error {
	dockerignore := `node_modules
npm-debug.log
.git
.gitignore
README.md
.env
.DS_Store
dist
build
.next`

	dockerignorePath := filepath.Join(projectPath, ".dockerignore")
	return os.WriteFile(dockerignorePath, []byte(dockerignore), 0644)
}
