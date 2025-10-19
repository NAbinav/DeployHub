package cmd

import (
	"fmt"
	"github.com/go-git/go-git/v5"
	"os"
	"path/filepath"
)

func CloneRepo(url string, name string) error {
	targetPath := filepath.Join("/tmp", name)
	fmt.Printf("📦 Cloning repository: %s → %s\n", url, targetPath)
	_, err := git.PlainClone(targetPath, false, &git.CloneOptions{
		URL:      url,
		Progress: os.Stdout,
	})
	if err != nil {
		return fmt.Errorf("failed to clone repository: %w", err)
	}
	fmt.Println("✅ Clone completed successfully!")
	return nil
}
