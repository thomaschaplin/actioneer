package internal

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"maps"
	"os"
	"os/exec"
	"path/filepath"

	git "github.com/go-git/go-git/v5"
)

type ActionFunc func(step Step) error

var EnvVars = map[string]string{}
var SecureEnvVars = map[string]struct{}{}

var builtInActions = map[string]ActionFunc{
	"actions/checkout@v1":          CheckoutAction,
	"actions/cat@v1":               CatAction,
	"actions/shell@v1":             ShellAction,
	"actions/env@v1":               EnvAction,
	"actions/env-secure@v1":        EnvSecureAction,
	"actions/env-unset@v1":         EnvUnsetAction,
	"actions/print-env@v1":         PrintEnvAction,
	"actions/upload-artifact@v1":   UploadArtifactAction,
	"actions/download-artifact@v1": DownloadArtifactAction,
}

func GetAction(name string) (ActionFunc, bool) {
	a, ok := builtInActions[name]
	return a, ok
}

func CheckoutAction(step Step) error {
	repo := step.URL // Expect the repo URL in the 'url' field
	if repo == "" {
		return fmt.Errorf("checkout action requires a repo URL in the 'url' field")
	}
	tmpDir := filepath.Join(".actioneer-tmp", "checkout")
	if err := os.RemoveAll(tmpDir); err != nil {
		return fmt.Errorf("failed to clean checkout dir: %w", err)
	}
	if err := os.MkdirAll(tmpDir, 0o755); err != nil {
		return fmt.Errorf("failed to create checkout dir: %w", err)
	}
	_, err := git.PlainClone(tmpDir, false, &git.CloneOptions{
		URL:      repo,
		Progress: nil,
	})
	if err != nil {
		return fmt.Errorf("git clone failed: %w", err)
	}
	return nil
}

func CatAction(step Step) error {
	if len(step.Files) == 0 {
		return fmt.Errorf("cat action requires a 'files' array field")
	}
	// Always run in the checkout directory
	workdir := filepath.Join(".actioneer-tmp", "checkout")
	for _, file := range step.Files {
		f, err := os.Open(filepath.Join(workdir, file))
		if err != nil {
			return fmt.Errorf("cat: could not open %s: %w", file, err)
		}
		defer f.Close()
		if _, err := io.Copy(os.Stdout, f); err != nil {
			return fmt.Errorf("cat: error reading %s: %w", file, err)
		}
	}
	return nil
}

// Helper: check if command references any secure env var
func containsSecureEnvVar(cmd string) (string, bool) {
	for k := range SecureEnvVars {
		if EnvVars[k] != "" && bytes.Contains([]byte(cmd), []byte(k)) {
			return k, true
		}
	}
	return "", false
}

// Helper: mask secure env values in output
func maskSecureEnvVars(line string) string {
	for k := range SecureEnvVars {
		if EnvVars[k] != "" && bytes.Contains([]byte(line), []byte(EnvVars[k])) {
			line = string(bytes.ReplaceAll([]byte(line), []byte(EnvVars[k]), []byte("***")))
		}
	}
	return line
}

func ShellAction(step Step) error {
	if step.Command == "" {
		return fmt.Errorf("shell action requires a 'command' field")
	}
	if k, found := containsSecureEnvVar(step.Command); found {
		return fmt.Errorf("shell command references secure env var '%s', which is not allowed", k)
	}
	workdir := filepath.Join(".actioneer-tmp", "checkout")
	cmd := exec.Command("sh", "-c", step.Command)
	cmd.Dir = workdir
	cmd.Env = os.Environ()
	for k, v := range EnvVars {
		cmd.Env = append(cmd.Env, k+"="+v)
	}
	stdout, _ := cmd.StdoutPipe()
	stderr, _ := cmd.StderrPipe()
	if err := cmd.Start(); err != nil {
		return err
	}
	maskAndPrint := func(r io.Reader) {
		scanner := bufio.NewScanner(r)
		for scanner.Scan() {
			line := maskSecureEnvVars(scanner.Text())
			fmt.Println(line)
		}
	}
	go maskAndPrint(stdout)
	go maskAndPrint(stderr)
	return cmd.Wait()
}

func EnvAction(step Step) error {
	if len(step.Env) == 0 {
		return fmt.Errorf("env action requires an 'env' map field")
	}
	maps.Copy(EnvVars, step.Env)
	return nil
}

func EnvSecureAction(step Step) error {
	if len(step.Env) == 0 {
		return fmt.Errorf("env-secure action requires an 'env' map field")
	}
	for k, v := range step.Env {
		EnvVars[k] = v
		SecureEnvVars[k] = struct{}{}
	}
	return nil
}

func EnvUnsetAction(step Step) error {
	if len(step.Files) == 0 {
		return fmt.Errorf("env-unset action requires 'env' array with env var names to unset")
	}
	for _, k := range step.Files {
		delete(EnvVars, k)
		delete(SecureEnvVars, k)
	}
	return nil
}

func PrintEnvAction(step Step) error {
	masked := map[string]string{}
	for k, v := range EnvVars {
		if _, isSecure := SecureEnvVars[k]; isSecure {
			masked[k] = "***"
		} else {
			masked[k] = v
		}
	}
	b, err := json.MarshalIndent(masked, "", "  ")
	if err != nil {
		return err
	}
	fmt.Println(string(b))
	return nil
}

func UploadArtifactAction(step Step) error {
	if len(step.Files) == 0 {
		return fmt.Errorf("upload-artifact action requires a 'files' array field")
	}
	uploadDir := filepath.Join(".actioneer-tmp", "upload")
	if err := os.MkdirAll(uploadDir, 0o755); err != nil {
		return fmt.Errorf("failed to create upload dir: %w", err)
	}
	workdir := filepath.Join(".actioneer-tmp", "checkout")
	for _, file := range step.Files {
		src := filepath.Join(workdir, file)
		dst := filepath.Join(uploadDir, filepath.Base(file))
		srcFile, err := os.Open(src)
		if err != nil {
			return fmt.Errorf("failed to read artifact %s: %w", file, err)
		}
		defer srcFile.Close()
		dstFile, err := os.Create(dst)
		if err != nil {
			return fmt.Errorf("failed to write artifact %s: %w", file, err)
		}
		defer dstFile.Close()
		if _, err := io.Copy(dstFile, srcFile); err != nil {
			return fmt.Errorf("failed to copy artifact %s: %w", file, err)
		}
	}
	return nil
}

func DownloadArtifactAction(step Step) error {
	if len(step.Files) == 0 {
		return fmt.Errorf("download-artifact action requires a 'files' array field")
	}
	uploadDir := filepath.Join(".actioneer-tmp", "upload")
	workdir := filepath.Join(".actioneer-tmp", "checkout")
	if err := os.MkdirAll(workdir, 0o755); err != nil {
		return fmt.Errorf("failed to create checkout dir: %w", err)
	}
	for _, file := range step.Files {
		src := filepath.Join(uploadDir, filepath.Base(file))
		dst := filepath.Join(workdir, filepath.Base(file))
		srcFile, err := os.Open(src)
		if err != nil {
			return fmt.Errorf("failed to read artifact %s: %w", file, err)
		}
		defer srcFile.Close()
		dstFile, err := os.Create(dst)
		if err != nil {
			return fmt.Errorf("failed to write artifact %s: %w", file, err)
		}
		defer dstFile.Close()
		if _, err := io.Copy(dstFile, srcFile); err != nil {
			return fmt.Errorf("failed to copy artifact %s: %w", file, err)
		}
	}
	return nil
}
