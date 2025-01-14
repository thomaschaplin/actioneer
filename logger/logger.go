package logger

import (
	"fmt"
	"os"
	"time"
)

func LogExecutionDetails(job, step, command string, startTime, endTime time.Time, output, errorOutput string) error {
	duration := endTime.Sub(startTime)

	// Convert duration to a human-readable format
	var durationStr string
	if duration < time.Second {
		durationStr = fmt.Sprintf("%.2f ms", float64(duration.Microseconds())/1000)
	} else if duration < time.Minute {
		durationStr = fmt.Sprintf("%.2f seconds", duration.Seconds())
	} else {
		durationStr = fmt.Sprintf("%.2f minutes", duration.Minutes())
	}

	// Open the logs.md file in append mode, creating it if it doesn't exist
	file, err := os.OpenFile("logs.md", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return err
	}
	defer file.Close()

	// Write the log entry
	logEntry := fmt.Sprintf("<details><summary>%s > %s > %s</summary>\n\n", job, step, command)
	logEntry += fmt.Sprintf("- Start Time: `%s`\n", startTime.Format(time.RFC3339))
	logEntry += fmt.Sprintf("- End Time: `%s`\n", endTime.Format(time.RFC3339))
	logEntry += fmt.Sprintf("- Duration: `%s`\n", durationStr)
	if len(output) > 0 {
		logEntry += fmt.Sprintf("<pre>%s</pre>\n", output)
	}
	if len(errorOutput) > 0 {
		logEntry += fmt.Sprintf("<pre>%s</pre>\n", errorOutput)
	}
	logEntry += "</details>\n\n"

	_, err = file.WriteString(logEntry)
	return err
}
