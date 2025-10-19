package utils

import (
	"fmt"
	"os"
	"os/exec"
)

func RunCommand(name string, args ...string) error {
	fmt.Printf("🛠 Running: %s %v\n", name, args)
	cmd := exec.Command(name, args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Stdin = os.Stdin
	return cmd.Run()
}
func RunCommandArray(cmdParts []string) error {
	fmt.Printf("🛠 Running: %s \n", cmdParts)
	cmd := exec.Command(cmdParts[0], cmdParts[1:]...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Stdin = os.Stdin
	return cmd.Run()
}
