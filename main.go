package main

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"io/ioutil"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"thomaschaplin/github-actions-poc/actions"
	"thomaschaplin/github-actions-poc/executor"
	"thomaschaplin/github-actions-poc/logger"
	"thomaschaplin/github-actions-poc/parser"
	"time"

	"github.com/google/uuid"
)

var webhookSecret = os.Getenv("ACTIONEER_WEBHOOK_SECRET")

// validateSignature checks if the request signature matches the expected signature
func validateSignature(secret, body []byte, signature string) bool {
	h := hmac.New(sha256.New, secret)
	h.Write(body)
	expectedSignature := "sha256=" + hex.EncodeToString(h.Sum(nil))
	return hmac.Equal([]byte(expectedSignature), []byte(signature))
}

// webhookHandler handles incoming webhook requests
func webhookHandler(w http.ResponseWriter, r *http.Request) {
	// Read the payload
	body, err := ioutil.ReadAll(r.Body)
	defer r.Body.Close()
	if err != nil {
		logger.LogJSON(map[string]interface{}{
			"event":     "error",
			"message":   "Error reading request body",
			"error":     err.Error(),
			"timestamp": time.Now().Format(time.RFC3339),
		})
		http.Error(w, "unable to read request body", http.StatusBadRequest)
		return
	}

	// Validate the GitHub signature
	signature := r.Header.Get("X-Hub-Signature-256")
	if signature == "" {
		logger.LogJSON(map[string]interface{}{
			"event":     "security_alert",
			"message":   "Missing X-Hub-Signature-256",
			"status":    "403 Forbidden",
			"timestamp": time.Now().Format(time.RFC3339),
		})
		http.Error(w, "Missing signature", http.StatusForbidden)
		return
	}
	if !validateSignature([]byte(webhookSecret), body, signature) {
		logger.LogJSON(map[string]interface{}{
			"event":     "security_alert",
			"message":   "Invalid signature",
			"status":    "403 Forbidden",
			"timestamp": time.Now().Format(time.RFC3339),
		})
		http.Error(w, "invalid signature", http.StatusForbidden)
		return
	}

	// Parse the payload
	var payload map[string]interface{}
	if err := json.Unmarshal(body, &payload); err != nil {
		logger.LogJSON(map[string]interface{}{
			"event":     "error",
			"message":   "Invalid JSON payload",
			"error":     err.Error(),
			"timestamp": time.Now().Format(time.RFC3339),
		})
		http.Error(w, "invalid JSON payload", http.StatusBadRequest)
		return
	}

	// Log webhook reception
	logger.LogJSON(map[string]interface{}{
		"event":     "webhook_received",
		"message":   "Received GitHub webhook",
		"payload":   payload,
		"timestamp": time.Now().Format(time.RFC3339),
	})

	if eventName := r.Header.Get("X-GitHub-Event"); eventName == "push" {
		processPushEvent(payload)
	}

	w.WriteHeader(http.StatusOK)
	w.Write([]byte("Webhook received"))
}

func processPushEvent(payload map[string]interface{}) {
	var ref, commitMessage, author, commitSha string

	// Extract the ref
	if r, ok := payload["ref"].(string); ok {
		ref = r
	}

	// Extract the commit data
	if headCommit, ok := payload["head_commit"].(map[string]interface{}); ok {
		if id, exists := headCommit["id"].(string); exists {
			commitSha = id
		}
		if message, exists := headCommit["message"].(string); exists {
			commitMessage = message
		}
		if authorData, exists := headCommit["author"].(map[string]interface{}); exists {
			if name, exists := authorData["name"].(string); exists {
				author = name
			}
		}
	}

	// Extract the repository data
	repoData, repoExists := payload["repository"].(map[string]interface{})
	if !repoExists {
		logger.LogJSON(map[string]interface{}{
			"event":     "error",
			"message":   "Repository data is missing",
			"timestamp": time.Now().Format(time.RFC3339),
		})
		return
	}

	// Extract the repository URL
	repoUrl, ok := repoData["html_url"].(string)
	if !ok {
		logger.LogJSON(map[string]interface{}{
			"event":     "error",
			"message":   "repoUrl is not a string",
			"timestamp": time.Now().Format(time.RFC3339),
		})
		return
	}

	// Log commit details
	logger.LogJSON(map[string]interface{}{
		"event":          "push_event_received",
		"repository":     repoUrl,
		"commit_sha":     commitSha,
		"commit_message": commitMessage,
		"author":         author,
		"timestamp":      time.Now().Format(time.RFC3339),
	})

	deleteErr := logger.DeleteLogs()
	if deleteErr != nil {
		logger.LogJSON(map[string]interface{}{
			"event":     "configuration_error",
			"message":   "ACTIONEER_WEBHOOK_SECRET is not set",
			"error":     deleteErr.Error(),
			"timestamp": time.Now().Format(time.RFC3339),
		})
	}

	// Clone the repository
	if ref != "" {
		actions.Checkout("Checkout", repoUrl, strings.TrimPrefix(ref, "refs/heads/"), "./repo")
	} else {
		logger.LogJSON(map[string]interface{}{
			"event":     "error",
			"message":   "ref is empty",
			"timestamp": time.Now().Format(time.RFC3339),
		})
		return
	}

	// Parse the workflow file
	workflowFile := "./repo/workflow.yaml"
	workflow, err := parser.ParseWorkflow(workflowFile)
	if err != nil {
		logger.LogJSON(map[string]interface{}{
			"event":     "workflow_parse_error",
			"message":   "Failed to parse workflow",
			"error":     err.Error(),
			"timestamp": time.Now().Format(time.RFC3339),
		})
		return
	}

	// Generate execution ID
	executionId := uuid.New().String()

	// Log execution start
	logger.LogJSON(map[string]interface{}{
		"id":        executionId,
		"event":     "workflow_execution_start",
		"message":   "Executing workflow",
		"workflow":  workflow,
		"timestamp": time.Now().Format(time.RFC3339),
	})

	executor.ExecuteWorkflow(workflow)

	// Log execution completion
	logger.LogJSON(map[string]interface{}{
		"id":        executionId,
		"event":     "workflow_execution_completed",
		"message":   "Execution completed",
		"timestamp": time.Now().Format(time.RFC3339),
	})
}

func main() {
	webhookSecret := os.Getenv("ACTIONEER_WEBHOOK_SECRET")
	if webhookSecret == "" {
		logger.LogJSON(map[string]interface{}{
			"event":     "configuration_error",
			"message":   "ACTIONEER_WEBHOOK_SECRET is not set",
			"timestamp": time.Now().Format(time.RFC3339),
		})
		os.Exit(1)
	}

	logID := uuid.New().String()

	// Log server startup
	logger.LogJSON(map[string]interface{}{
		"id":        logID,
		"event":     "server_start",
		"message":   "Starting server",
		"port":      8080,
		"timestamp": time.Now().Format(time.RFC3339),
	})

	server := &http.Server{Addr: ":8080"}
	go gracefulShutdown(server)

	http.HandleFunc("/", webhookHandler)
	if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		logger.LogJSON(map[string]interface{}{
			"event":     "server_error",
			"message":   "Server failed",
			"error":     err.Error(),
			"timestamp": time.Now().Format(time.RFC3339),
		})
	}
}

// Graceful shutdown for handling OS signals
func gracefulShutdown(server *http.Server) {
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt, syscall.SIGTERM)

	<-stop
	logger.LogJSON(map[string]interface{}{
		"event":     "server_shutdown",
		"message":   "Shutting down server",
		"timestamp": time.Now().Format(time.RFC3339),
	})

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	server.Shutdown(ctx)
}
