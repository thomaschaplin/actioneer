package main

import (
	"encoding/json"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"github.com/thomaschaplin/actioneer/internal"
)

type LogEntry struct {
	Level   string `json:"level"`
	Message string `json:"message"`
	Step    string `json:"step,omitempty"`
	Detail  any    `json:"detail,omitempty"`
}

func logJSON(level, msg, step string, detail any) {
	entry := LogEntry{Level: level, Message: msg, Step: step, Detail: detail}
	b, _ := json.Marshal(entry)
	log.Println(string(b))
}

func executeWorkflow(workflowFile string) {
	start := time.Now()
	cfg, err := internal.LoadConfig(workflowFile)
	if err != nil {
		logJSON("fatal", "failed to load config", workflowFile, err.Error())
		return
	}
	logJSON("info", "Loaded pipeline", workflowFile, cfg.Pipeline.Name)
	for _, step := range cfg.Pipeline.Steps {
		logJSON("info", "Starting step", step.Name, nil)
		if step.Uses != "" {
			logJSON("info", "Using action", step.Name, step.Uses)
			if action, ok := internal.GetAction(step.Uses); ok {
				if step.Uses == "actions/shell@v1" {
					// Set env for shell action
					env := os.Environ()
					for k, v := range internal.EnvVars {
						env = append(env, k+"="+v)
					}
					os.Clearenv()
					for _, e := range env {
						parts := []rune(e)
						for i, c := range parts {
							if c == '=' {
								os.Setenv(string(parts[:i]), string(parts[i+1:]))
								break
							}
						}
					}
				}
				start := time.Now()
				if err := action(step); err != nil {
					duration := time.Since(start)
					logJSON("error", "Action error", step.Name, map[string]any{"error": err.Error(), "duration": duration.String()})
					return
				}
				duration := time.Since(start)
				logJSON("info", "Completed step", step.Name, map[string]any{"duration": duration.String()})
			} else {
				logJSON("error", "Unknown action", step.Name, step.Uses)
				return
			}
		}
	}
	duration := time.Since(start)
	logJSON("info", "Pipeline completed", cfg.Pipeline.Name, map[string]any{"duration": duration.String()})
}

func main() {
	if len(os.Args) == 2 && os.Args[1] == "--webhook" {
		http.HandleFunc("/webhook", func(w http.ResponseWriter, r *http.Request) {
			if r.Method != http.MethodPost {
				w.WriteHeader(http.StatusMethodNotAllowed)
				return
			}
			var payload struct {
				Workflow string `json:"workflow"`
			}
			if err := json.NewDecoder(r.Body).Decode(&payload); err != nil || payload.Workflow == "" {
				w.WriteHeader(http.StatusBadRequest)
				w.Write([]byte("Missing or invalid workflow field"))
				return
			}
			file := filepath.Join(".actioneer", "workflows", payload.Workflow)
			go executeWorkflow(file)
			w.WriteHeader(http.StatusAccepted)
			w.Write([]byte("Workflow execution started"))
		})
		logJSON("info", "Starting webhook server on :8080", "", nil)
		if err := http.ListenAndServe(":8080", nil); err != nil {
			logJSON("fatal", "HTTP server error", "", err.Error())
		}
		return
	}
	// CLI mode (default)
	if len(os.Args) < 2 {
		logJSON("fatal", "no workflow file specified", "", "Usage: actioneer <workflow.yaml> | actioneer webhook")
		return
	}
	file := filepath.Join(".actioneer", "workflows", os.Args[1])
	executeWorkflow(file)
}
