package actions

import (
	"fmt"
	"os"
	"os/exec"
	"strings"
)

func Checkout(jobName, repoURL, branch, dest string) error {
	action := func() (string, string, error) {
		// Check if the directory exists and remove it if it does
		if _, err := os.Stat(dest); err == nil {
			if err := os.RemoveAll(dest); err != nil {
				return "", "", fmt.Errorf("failed to remove existing directory: %v", err)
			}
		} else if !os.IsNotExist(err) {
			return "", "", fmt.Errorf("failed to check if directory exists: %v", err)
		}

		// Prepare and run the git clone command
		cmd := exec.Command("git", "clone", "--branch", branch, "--progress", repoURL, dest)
		var out strings.Builder
		cmd.Stdout = &out
		cmd.Stderr = &out

		err := cmd.Run()
		return out.String(), "", err
	}

	return WithTiming(jobName, "Checkout", fmt.Sprintf("git clone %s %s", repoURL, dest), action)
}
