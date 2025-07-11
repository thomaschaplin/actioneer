package internal

import (
	"bytes"
	"os"
	"path/filepath"
	"testing"
)

func TestGetAction(t *testing.T) {
	action, ok := GetAction("actions/checkout@v1")
	if !ok || action == nil {
		t.Error("expected to find built-in action 'actions/checkout@v1'")
	}
}

func TestCheckoutAction_MissingURL(t *testing.T) {
	step := Step{Uses: "actions/checkout@v1"}
	err := CheckoutAction(step)
	if err == nil {
		t.Error("expected error when URL is missing")
	}
}

func TestCheckoutAction_InvalidRepo(t *testing.T) {
	step := Step{Uses: "actions/checkout@v1", URL: "https://invalid/repo/url"}
	err := CheckoutAction(step)
	if err == nil {
		t.Error("expected error for invalid repo URL")
	}
	_ = os.RemoveAll(".actioneer-tmp")
}

func TestCatAction_NoFiles(t *testing.T) {
	step := Step{Uses: "actions/cat@v1"}
	err := CatAction(step)
	if err == nil {
		t.Error("expected error when files are missing")
	}
}

func TestShellAction_NoCommand(t *testing.T) {
	step := Step{Uses: "actions/shell@v1"}
	err := ShellAction(step)
	if err == nil {
		t.Error("expected error when command is missing")
	}
}

func TestEnvAction_NoEnv(t *testing.T) {
	step := Step{Uses: "actions/env@v1"}
	err := EnvAction(step)
	if err == nil {
		t.Error("expected error when env is missing")
	}
}

func TestEnvSecureAction_NoEnv(t *testing.T) {
	step := Step{Uses: "actions/env-secure@v1"}
	err := EnvSecureAction(step)
	if err == nil {
		t.Error("expected error when env is missing")
	}
}

func TestEnvUnsetAction_NoFiles(t *testing.T) {
	step := Step{Uses: "actions/env-unset@v1"}
	err := EnvUnsetAction(step)
	if err == nil {
		t.Error("expected error when files are missing")
	}
}

func TestCatAction_Success(t *testing.T) {
	// Setup: create checkout dir and test file
	workdir := filepath.Join(".actioneer-tmp", "checkout")
	os.MkdirAll(workdir, 0o755)
	testFile := filepath.Join(workdir, "hello.txt")
	content := []byte("Hello, Actioneer!\n")
	if err := os.WriteFile(testFile, content, 0644); err != nil {
		t.Fatalf("failed to write test file: %v", err)
	}
	defer os.RemoveAll(".actioneer-tmp")

	step := Step{Uses: "actions/cat@v1", Files: []string{"hello.txt"}}
	err := CatAction(step)
	if err != nil {
		t.Errorf("expected no error, got: %v", err)
	}
}

func TestEnvAction_Success(t *testing.T) {
	EnvVars = map[string]string{}
	step := Step{Uses: "actions/env@v1", Env: map[string]string{"FOO": "bar"}}
	err := EnvAction(step)
	if err != nil {
		t.Errorf("expected no error, got: %v", err)
	}
	if EnvVars["FOO"] != "bar" {
		t.Errorf("expected EnvVars[FOO] to be 'bar', got '%s'", EnvVars["FOO"])
	}
}

func TestEnvSecureAction_Success(t *testing.T) {
	EnvVars = map[string]string{}
	SecureEnvVars = map[string]struct{}{}
	step := Step{Uses: "actions/env-secure@v1", Env: map[string]string{"SECRET": "shh"}}
	err := EnvSecureAction(step)
	if err != nil {
		t.Errorf("expected no error, got: %v", err)
	}
	if EnvVars["SECRET"] != "shh" {
		t.Errorf("expected EnvVars[SECRET] to be 'shh', got '%s'", EnvVars["SECRET"])
	}
	if _, ok := SecureEnvVars["SECRET"]; !ok {
		t.Error("expected SECRET to be in SecureEnvVars")
	}
}

func TestEnvUnsetAction_Success(t *testing.T) {
	EnvVars = map[string]string{"FOO": "bar"}
	SecureEnvVars = map[string]struct{}{"FOO": {}}
	step := Step{Uses: "actions/env-unset@v1", Files: []string{"FOO"}}
	err := EnvUnsetAction(step)
	if err != nil {
		t.Errorf("expected no error, got: %v", err)
	}
	if _, ok := EnvVars["FOO"]; ok {
		t.Error("expected FOO to be removed from EnvVars")
	}
	if _, ok := SecureEnvVars["FOO"]; ok {
		t.Error("expected FOO to be removed from SecureEnvVars")
	}
}

func TestPrintEnvAction_SecureMasking(t *testing.T) {
	EnvVars = map[string]string{"FOO": "bar", "SECRET": "shh"}
	SecureEnvVars = map[string]struct{}{"SECRET": {}}
	step := Step{Uses: "actions/print-env@v1"}

	// Capture stdout
	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	err := PrintEnvAction(step)
	w.Close()
	os.Stdout = oldStdout

	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}

	var buf bytes.Buffer
	buf.ReadFrom(r)
	output := buf.String()

	if !bytes.Contains([]byte(output), []byte(`"FOO": "bar"`)) {
		t.Errorf("expected FOO to be visible, got output: %s", output)
	}
	if !bytes.Contains([]byte(output), []byte(`"SECRET": "***"`)) {
		t.Errorf("expected SECRET to be masked, got output: %s", output)
	}
}

func TestUploadAndDownloadArtifactAction(t *testing.T) {
	// Setup: create checkout dir and test file
	checkoutDir := filepath.Join(".actioneer-tmp", "checkout")
	os.MkdirAll(checkoutDir, 0o755)
	testFile := filepath.Join(checkoutDir, "artifact.txt")
	content := []byte("artifact content\n")
	if err := os.WriteFile(testFile, content, 0644); err != nil {
		t.Fatalf("failed to write test file: %v", err)
	}
	defer os.RemoveAll(".actioneer-tmp")

	// Upload
	uploadStep := Step{Uses: "actions/upload-artifact@v1", Files: []string{"artifact.txt"}}
	err := UploadArtifactAction(uploadStep)
	if err != nil {
		t.Fatalf("UploadArtifactAction failed: %v", err)
	}

	uploadedFile := filepath.Join(".actioneer-tmp", "upload", "artifact.txt")
	if _, err := os.Stat(uploadedFile); err != nil {
		t.Errorf("expected uploaded file to exist: %v", err)
	}

	// Remove from checkout to ensure download works
	os.Remove(testFile)

	// Download
	downloadStep := Step{Uses: "actions/download-artifact@v1", Files: []string{"artifact.txt"}}
	err = DownloadArtifactAction(downloadStep)
	if err != nil {
		t.Fatalf("DownloadArtifactAction failed: %v", err)
	}

	downloadedFile := filepath.Join(checkoutDir, "artifact.txt")
	data, err := os.ReadFile(downloadedFile)
	if err != nil {
		t.Errorf("expected downloaded file to exist: %v", err)
	}
	if string(data) != string(content) {
		t.Errorf("expected downloaded file content to match, got: %s", string(data))
	}
}
