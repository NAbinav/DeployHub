package handler

import (
	"deployhub/cmd"
	"encoding/json"
	"fmt"
	"github.com/google/uuid"
	"net/http"
)

type CloneResponse struct {
	Status  string `json:"status"`
	Folder  string `json:"folder,omitempty"`
	Message string `json:"message"`
	Error   string `json:"error,omitempty"`
}

func CloneHandler(w http.ResponseWriter, r *http.Request) {
	// Parse query parameter
	url := r.URL.Query().Get("url")
	if url == "" {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(CloneResponse{
			Status:  "error",
			Message: "URL parameter is required",
		})
		return
	}

	// Generate temp folder name
	tempFolder := uuid.NewString()

	// Clone repository
	err := cmd.CloneRepo(url, tempFolder)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(CloneResponse{
			Status:  "failed",
			Message: fmt.Sprintf("Clone failed: %v", err),
			Error:   err.Error(),
		})
		return
	}

	// Success response
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(CloneResponse{
		Status:  "success",
		Folder:  tempFolder,
		Message: "Clone completed successfully!",
	})
}
