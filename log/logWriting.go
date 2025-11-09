package log

import (
	"context"
	"deployhub/db"
	"fmt"
)

type LogWriter struct {
	DeploymentID string
	User         string
	Ctx          context.Context
	Stream       chan string
}

func (lw *LogWriter) Write(p []byte) (n int, err error) {
	logMessage := string(p)

	if err := db.AddLog(lw.DeploymentID, lw.User, logMessage, lw.Ctx); err != nil {
		fmt.Printf("Failed to store log: %v\n", err)
	}

	select {
	case lw.Stream <- logMessage:
	default:
		fmt.Printf("SSE stream full, dropping log: %s\n", logMessage)
	}
	fmt.Print(logMessage)
	return len(p), nil
}
