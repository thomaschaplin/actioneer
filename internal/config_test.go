package internal

import (
	"path/filepath"
	"testing"
)

func TestLoadConfig(t *testing.T) {
	tmpFile := filepath.Join("testdata", "pipeline.yaml")

	cfg, err := LoadConfig(tmpFile)
	if err != nil {
		t.Fatalf("LoadConfig failed: %v", err)
	}
	if cfg.Pipeline.Name != "test-pipeline" {
		t.Errorf("expected pipeline name 'test-pipeline', got '%s'", cfg.Pipeline.Name)
	}
	if len(cfg.Pipeline.Steps) != 2 {
		t.Errorf("expected 2 steps, got %d", len(cfg.Pipeline.Steps))
	}
}

func TestLoadConfig_InvalidFile(t *testing.T) {
	tmpFile := filepath.Join("testdata", "doesnotexist.yaml")
	_, err := LoadConfig(tmpFile)
	if err == nil {
		t.Error("expected error for missing config file")
	}
}

func TestLoadConfig_InvalidYAML(t *testing.T) {
	tmpFile := filepath.Join("testdata", "invalid.yaml")
	_, err := LoadConfig(tmpFile)
	if err == nil {
		t.Error("expected error for invalid YAML")
	}
}
