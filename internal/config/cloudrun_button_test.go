package config_test

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// CloudRunAppSpec represents the schema of app.json for Google Cloud Run Button.
// Specification: https://github.com/GoogleCloudPlatform/cloud-run-button
type CloudRunAppSpec struct {
	Name    string                     `json:"name"`
	Env     map[string]CloudRunEnvSpec `json:"env"`
	Options CloudRunOptionsSpec        `json:"options"`
}

type CloudRunEnvSpec struct {
	Description string `json:"description"`
	Value       string `json:"value,omitempty"`
	Required    bool   `json:"required,omitempty"`
}

type CloudRunOptionsSpec struct {
	AllowUnauthenticated bool   `json:"allow-unauthenticated"`
	Port                 int    `json:"port"`
	Memory               string `json:"memory"`
	CPU                  string `json:"cpu"`
}

func findRepoRoot(t *testing.T) string {
	t.Helper()
	dir, err := os.Getwd()
	if err != nil {
		t.Fatalf("failed to get working dir: %v", err)
	}

	for {
		if _, err := os.Stat(filepath.Join(dir, "go.mod")); err == nil {
			return dir
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			t.Fatalf("could not locate repo root containing go.mod from %s", dir)
		}
		dir = parent
	}
}

func TestCloudRunButtonAppJsonSpecification(t *testing.T) {
	rootDir := findRepoRoot(t)
	appJSONPath := filepath.Join(rootDir, "app.json")

	// 1. File existence
	data, err := os.ReadFile(appJSONPath)
	if err != nil {
		t.Fatalf("app.json does not exist at repo root: %v", err)
	}

	// 2. Valid JSON with DisallowUnknownFields() matching Cloud Run Button backend
	dec := json.NewDecoder(bytes.NewReader(data))
	dec.DisallowUnknownFields()
	var spec CloudRunAppSpec
	if err := dec.Decode(&spec); err != nil {
		t.Fatalf("failed to parse app.json with DisallowUnknownFields: %v", err)
	}

	// 3. Name
	if spec.Name != "agy-cost-board" {
		t.Errorf("expected spec.Name to be 'agy-cost-board', got: '%s'", spec.Name)
	}

	// 4. Options
	if spec.Options.AllowUnauthenticated {
		t.Errorf("expected allow-unauthenticated to be false for IAM security, got true")
	}
	if spec.Options.Port != 8080 {
		t.Errorf("expected port to be 8080, got %d", spec.Options.Port)
	}
	if spec.Options.Memory != "512Mi" {
		t.Errorf("expected memory to be '512Mi', got '%s'", spec.Options.Memory)
	}
	if spec.Options.CPU != "1" {
		t.Errorf("expected cpu to be '1', got '%s'", spec.Options.CPU)
	}

	// 5. Environment Variables
	expectedEnvKeys := []string{
		"PROJECT_ID",
		"AGY_COST_BOARD_DEMO",
		"AGY_COST_BOARD_TELEMETRY_TABLE",
		"AGY_COST_BOARD_BILLING_TABLE",
	}

	for _, key := range expectedEnvKeys {
		envItem, exists := spec.Env[key]
		if !exists {
			t.Errorf("expected env item '%s' to be defined in app.json", key)
			continue
		}
		if envItem.Description == "" {
			t.Errorf("expected env item '%s' to have non-empty description", key)
		}
	}

	// Check demo default value is false
	if demoEnv, ok := spec.Env["AGY_COST_BOARD_DEMO"]; ok {
		if demoEnv.Value != "false" {
			t.Errorf("expected AGY_COST_BOARD_DEMO default value to be 'false', got '%s'", demoEnv.Value)
		}
	}
}

func TestReadmeCloudRunButtonIntegration(t *testing.T) {
	rootDir := findRepoRoot(t)
	readmePath := filepath.Join(rootDir, "README.md")

	data, err := os.ReadFile(readmePath)
	if err != nil {
		t.Fatalf("failed to read README.md: %v", err)
	}
	content := string(data)

	// 1. Badge row contains official Cloud Run Button badge and link
	expectedBadgeMarkdown := "[![Run on Google Cloud](https://deploy.cloud.run/button.svg)](https://deploy.cloud.run)"
	if !strings.Contains(content, expectedBadgeMarkdown) {
		t.Errorf("expected README.md to contain Cloud Run Button badge: %s", expectedBadgeMarkdown)
	}

	// 2. Dedicated Cloud Run One-Click Deployment section
	if !strings.Contains(content, "Deploy to Cloud Run in One Click") {
		t.Errorf("expected README.md to contain 'Deploy to Cloud Run in One Click' section")
	}

	// 3. Mentions IAM authentication security and proxy invocation
	if !strings.Contains(content, "gcloud run services proxy") {
		t.Errorf("expected README.md to document IAM authenticated access via 'gcloud run services proxy'")
	}
}

