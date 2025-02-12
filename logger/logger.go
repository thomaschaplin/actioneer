package logger

import (
	"encoding/json"
	"fmt"
	"os"
	"time"
)

func LogJSON(logData map[string]interface{}) {
	jsonData, err := json.MarshalIndent(logData, "", "  ")
	if err != nil {
		fmt.Println(`{"event": "error", "message": "Error marshalling JSON", "error": "` + err.Error() + `"}`)
		return
	}
	fmt.Println(string(jsonData))
}

func DeleteLogs() error {
	err := os.Remove("logs.md")
	if err != nil {
		if os.IsNotExist(err) {
			LogJSON(map[string]interface{}{
				"event":     "log_file_cleanup",
				"message":   "logs.md does not exist, no need to delete",
				"timestamp": time.Now().Format(time.RFC3339),
			})
			return nil
		}
		LogJSON(map[string]interface{}{
			"event":     "error",
			"message":   "Failed to delete logs.md",
			"error":     err.Error(),
			"timestamp": time.Now().Format(time.RFC3339),
		})
		return err
	}

	LogJSON(map[string]interface{}{
		"event":     "log_file_cleanup",
		"message":   "Successfully deleted logs.md",
		"timestamp": time.Now().Format(time.RFC3339),
	})
	return nil
}

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
		LogJSON(map[string]interface{}{
			"event":     "error",
			"message":   "Failed to open logs.md",
			"error":     err.Error(),
			"timestamp": time.Now().Format(time.RFC3339),
		})
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
	if err != nil {
		LogJSON(map[string]interface{}{
			"event":     "error",
			"message":   "Failed to write to logs.md",
			"error":     err.Error(),
			"timestamp": time.Now().Format(time.RFC3339),
		})
		return err
	}

	LogJSON(map[string]interface{}{
		"event":     "log_entry_written",
		"message":   "Execution details logged",
		"job":       job,
		"step":      step,
		"command":   command,
		"timestamp": time.Now().Format(time.RFC3339),
	})
	return nil
}
