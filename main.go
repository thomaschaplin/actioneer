package main

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"
	"thomaschaplin/github-actions-poc/actions"
	"thomaschaplin/github-actions-poc/executor"
	"thomaschaplin/github-actions-poc/parser"
)

// Replace with your actual webhook secret
const webhookSecret = "mywebhooksecret"

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
	if err != nil {
		http.Error(w, "unable to read request body", http.StatusBadRequest)
		return
	}
	defer r.Body.Close()

	// Validate the GitHub signature
	signature := r.Header.Get("X-Hub-Signature-256")
	if !validateSignature([]byte(webhookSecret), body, signature) {
		http.Error(w, "invalid signature", http.StatusForbidden)
		return
	}

	fmt.Println("Valid webhook received")

	// Parse the payload
	var payload map[string]interface{}
	if err := json.Unmarshal(body, &payload); err != nil {
		http.Error(w, "invalid JSON payload", http.StatusBadRequest)
		return
	}

	// Log the received payload
	// fmt.Printf("Received payload: %+v\n", payload)

	// Check for push event and process it
	if eventName := r.Header.Get("X-GitHub-Event"); eventName == "push" {
		processPushEvent(payload)
	}

	// Respond to GitHub
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("Webhook received"))
}

// processPushEvent handles the push event payload
func processPushEvent(payload map[string]interface{}) {
	var ref string
	if r, ok := payload["ref"].(string); ok {
		ref = r
		fmt.Printf("Push event received for ref: %s\n", ref)
	}

	if headCommit, ok := payload["head_commit"].(map[string]interface{}); ok {
		if message, exists := headCommit["message"].(string); exists {
			fmt.Printf("Commit message: %s\n", message)
		}
		if author, exists := headCommit["author"].(map[string]interface{}); exists {
			if name, exists := author["name"].(string); exists {
				fmt.Printf("Author: %s\n", name)
			}
		}
	}

	repoData, repoExists := payload["repository"].(map[string]interface{})
	if !repoExists {
		fmt.Println("Error: repository data is missing")
		return
	}

	repoUrl, ok := repoData["html_url"].(string)
	if !ok {
		fmt.Println("Error: repoUrl is not a string")
		return
	}

	fmt.Println("Processing push event for repo: ", repoUrl)
	if ref != "" {
		actions.Checkout("Checkout", repoUrl, strings.TrimPrefix(ref, "refs/heads/"), "./repo")
	} else {
		fmt.Println("Error: ref is empty")
	}

	workflowFile := "./repo/workflow.yaml"

	workflow, err := parser.ParseWorkflow(workflowFile)
	if err != nil {
		fmt.Printf("Error parsing workflow: %v\n", err)
		return
	}

	fmt.Printf("Parsed workflow: %+v\n", workflow)

	fmt.Println("\nExecuting workflow... \n")
	executor.ExecuteWorkflow(workflow)
	fmt.Println("\nExecution completed! \n")
}

func main() {
	http.HandleFunc("/", webhookHandler)
	fmt.Println("Starting server on :8080...")
	if err := http.ListenAndServe(":8080", nil); err != nil {
		panic(err)
	}
}
