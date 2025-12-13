package utils

import (
	"bufio"
	"fmt"
	"io"
	"os/exec"
)

func RunCommandWithOutput(name string, args []string, output io.Writer) error {
	if output != nil {
		fmt.Fprintf(output, "🛠 Running: %s %v\n", name, args)
	}

	cmd := exec.Command(name, args...)

	if output == nil {
		return cmd.Run()
	}

	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return fmt.Errorf("failed to get stdout pipe: %v", err)
	}

	stderr, err := cmd.StderrPipe()
	if err != nil {
		return fmt.Errorf("failed to get stderr pipe: %v", err)
	}

	if err := cmd.Start(); err != nil {
		return fmt.Errorf("failed to start command: %v", err)
	}

	multi := io.MultiWriter(output)
	scannerOut := bufio.NewScanner(stdout)
	scannerErr := bufio.NewScanner(stderr)

	go func() {
		for scannerOut.Scan() {
			fmt.Fprintln(multi, scannerOut.Text())
		}
	}()

	go func() {
		for scannerErr.Scan() {
			fmt.Fprintln(multi, scannerErr.Text())
		}
	}()

	err = cmd.Wait()
	if err != nil {
		return fmt.Errorf("command failed: %v", err)
	}

	return nil
}
