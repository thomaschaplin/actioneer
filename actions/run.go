package actions

import (
	"os/exec"
	"strings"
)

func RunCommand(jobName, command string) error {
	action := func() (string, string, error) {
		cmd := exec.Command("bash", "-c", command)
		cmd.Dir = "repo"

		var out, stderr strings.Builder
		cmd.Stdout = &out
		cmd.Stderr = &stderr

		err := cmd.Run()
		return out.String(), stderr.String(), err
	}

	return WithTiming(jobName, "RunCommand", command, action)
}
