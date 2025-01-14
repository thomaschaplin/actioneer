package actions

import (
	"time"
	"thomaschaplin/github-actions-poc/logger"
)

func WithTiming(job, step, command string, action func() (string, string, error)) error {
	startTime := time.Now()
	output, errorOutput, err := action()
	endTime := time.Now()

	logErr := logger.LogExecutionDetails(job, step, command, startTime, endTime, output, errorOutput)
	if logErr != nil {
		return logErr
	}

	return err
}
