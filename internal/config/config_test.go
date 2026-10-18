package config_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/julienbreux/agy-cost-board/internal/config"
)

func TestConfigLoadAndSave(t *testing.T) {
	tmpDir := t.TempDir()
	cfgPath := filepath.Join(tmpDir, ".agy-cost-board.yaml")

	t.Run("loading non-existent file returns default empty config without error", func(t *testing.T) {
		cfg, err := config.Load(filepath.Join(tmpDir, "missing.yaml"))
		if err != nil {
			t.Fatalf("unexpected error loading missing config: %v", err)
		}
		if cfg == nil {
			t.Fatalf("expected non-nil config")
		}
		if cfg.ProjectID != "" {
			t.Errorf("expected empty project ID, got: %s", cfg.ProjectID)
		}
	})

	t.Run("save and load roundtrip preserves values", func(t *testing.T) {
		initial := &config.FileConfig{
			ProjectID:      "test-persist-project",
			TelemetryTable: "test-persist-project.antigravity.inference_logs",
			BillingTable:   "test-persist-project.billing.gcp_billing_export_v1_000",
			SinkName:       "my-custom-sink",
			SeatQuota:      25,
		}

		err := config.Save(cfgPath, initial)
		if err != nil {
			t.Fatalf("failed to save config: %v", err)
		}

		loaded, err := config.Load(cfgPath)
		if err != nil {
			t.Fatalf("failed to load saved config: %v", err)
		}

		if loaded.ProjectID != initial.ProjectID {
			t.Errorf("expected %s, got %s", initial.ProjectID, loaded.ProjectID)
		}
		if loaded.TelemetryTable != initial.TelemetryTable {
			t.Errorf("expected %s, got %s", initial.TelemetryTable, loaded.TelemetryTable)
		}
		if loaded.BillingTable != initial.BillingTable {
			t.Errorf("expected %s, got %s", initial.BillingTable, loaded.BillingTable)
		}
		if loaded.SinkName != initial.SinkName {
			t.Errorf("expected %s, got %s", initial.SinkName, loaded.SinkName)
		}
		if loaded.SeatQuota != initial.SeatQuota {
			t.Errorf("expected %d, got %d", initial.SeatQuota, loaded.SeatQuota)
		}
	})

	t.Run("loading corrupted config returns error", func(t *testing.T) {
		corruptPath := filepath.Join(tmpDir, "corrupt.yaml")
		_ = os.WriteFile(corruptPath, []byte("invalid: [yaml: broken"), 0644)

		_, err := config.Load(corruptPath)
		if err == nil {
			t.Fatalf("expected error loading corrupted yaml, got nil")
		}
	})

	t.Run("save with empty path uses default filename", func(t *testing.T) {
		initial := &config.FileConfig{ProjectID: "default-path-proj"}
		// Test Load with empty path
		cfg, err := config.Load("")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if cfg == nil {
			t.Fatalf("expected non-nil config")
		}

		// Save error on invalid path
		err = config.Save("/non_existent_dir_xyz/config.yaml", initial)
		if err == nil {
			t.Errorf("expected error saving to invalid directory, got nil")
		}
	})
}

func TestResolveSetting(t *testing.T) {
	t.Run("flag takes highest precedence", func(t *testing.T) {
		t.Setenv("TEST_ENV_VAR", "env_value")
		val := config.ResolveSetting("flag_value", "TEST_ENV_VAR", "config_value", "default_value")
		if val != "flag_value" {
			t.Errorf("expected flag_value, got %s", val)
		}
	})

	t.Run("env takes precedence when flag is empty", func(t *testing.T) {
		t.Setenv("TEST_ENV_VAR", "env_value")
		val := config.ResolveSetting("", "TEST_ENV_VAR", "config_value", "default_value")
		if val != "env_value" {
			t.Errorf("expected env_value, got %s", val)
		}
	})

	t.Run("config value takes precedence when flag and env are empty", func(t *testing.T) {
		t.Setenv("TEST_ENV_VAR", "")
		val := config.ResolveSetting("", "TEST_ENV_VAR", "config_value", "default_value")
		if val != "config_value" {
			t.Errorf("expected config_value, got %s", val)
		}
	})

	t.Run("default value is used when flag, env, and config are empty", func(t *testing.T) {
		t.Setenv("TEST_ENV_VAR", "")
		val := config.ResolveSetting("", "TEST_ENV_VAR", "", "default_value")
		if val != "default_value" {
			t.Errorf("expected default_value, got %s", val)
		}
	})
}

func TestResolveIntSetting(t *testing.T) {
	t.Run("flag takes precedence if non-zero", func(t *testing.T) {
		val := config.ResolveIntSetting(50, 20, 10)
		if val != 50 {
			t.Errorf("expected 50, got %d", val)
		}
	})

	t.Run("config value takes precedence if flag is zero", func(t *testing.T) {
		val := config.ResolveIntSetting(0, 20, 10)
		if val != 20 {
			t.Errorf("expected 20, got %d", val)
		}
	})

	t.Run("default value is used if flag and config are zero", func(t *testing.T) {
		val := config.ResolveIntSetting(0, 0, 10)
		if val != 10 {
			t.Errorf("expected 10, got %d", val)
		}
	})
}

func TestViperConfigFeatures(t *testing.T) {
	t.Run("loading json configuration file", func(t *testing.T) {
		tmpDir := t.TempDir()
		jsonPath := filepath.Join(tmpDir, ".agy-cost-board.json")
		jsonContent := `{
  "project_id": "json-project",
  "telemetry_table": "json-project.telemetry.inference_logs",
  "billing_table": "json-project.billing.gcp_billing_export_v1_000",
  "sink_name": "json-sink",
  "dataset": "json_telemetry",
  "seat_quota": 48,
  "demo": false,
  "port": 9090,
  "host": "127.0.0.1"
}`
		if err := os.WriteFile(jsonPath, []byte(jsonContent), 0644); err != nil {
			t.Fatalf("failed to write json config: %v", err)
		}

		cfg, err := config.Load(jsonPath)
		if err != nil {
			t.Fatalf("failed to load json config: %v", err)
		}

		if cfg.ProjectID != "json-project" {
			t.Errorf("expected json-project, got %s", cfg.ProjectID)
		}
		if cfg.SeatQuota != 48 {
			t.Errorf("expected seat quota 48, got %d", cfg.SeatQuota)
		}
		if cfg.Port != 9090 {
			t.Errorf("expected port 9090, got %d", cfg.Port)
		}
		if cfg.Host != "127.0.0.1" {
			t.Errorf("expected host 127.0.0.1, got %s", cfg.Host)
		}
	})

	t.Run("environment variable override with AGY_COST_BOARD_ prefix", func(t *testing.T) {
		t.Setenv("AGY_COST_BOARD_PROJECT_ID", "env-prefixed-project")
		t.Setenv("AGY_COST_BOARD_TELEMETRY_TABLE", "prefixed.ds.telemetry")
		t.Setenv("AGY_COST_BOARD_BILLING_TABLE", "prefixed.billing.export")
		t.Setenv("AGY_COST_BOARD_SEAT_QUOTA", "77")
		t.Setenv("AGY_COST_BOARD_PORT", "9999")

		cfg, err := config.Load(filepath.Join(t.TempDir(), "nonexistent.yaml"))
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if cfg.ProjectID != "env-prefixed-project" {
			t.Errorf("expected env-prefixed-project, got %s", cfg.ProjectID)
		}
		if cfg.TelemetryTable != "prefixed.ds.telemetry" {
			t.Errorf("expected prefixed.ds.telemetry, got %s", cfg.TelemetryTable)
		}
		if cfg.BillingTable != "prefixed.billing.export" {
			t.Errorf("expected prefixed.billing.export, got %s", cfg.BillingTable)
		}
		if cfg.SeatQuota != 77 {
			t.Errorf("expected seat quota 77, got %d", cfg.SeatQuota)
		}
		if cfg.Port != 9999 {
			t.Errorf("expected port 9999, got %d", cfg.Port)
		}
	})

	t.Run("unprefixed legacy environment variables are ignored", func(t *testing.T) {
		t.Setenv("GCP_PROJECT", "legacy-gcp-project")
		t.Setenv("PROJECT_ID", "legacy-project-id")
		t.Setenv("TELEMETRY_TABLE", "legacy.ds.telemetry")
		t.Setenv("BILLING_TABLE", "legacy.billing.export")
		t.Setenv("SEAT_QUOTA", "52")
		t.Setenv("PORT", "7777")
		t.Setenv("_MCP_ENABLED", "false")
		t.Setenv("MCP_ENABLED", "false")

		cfg, err := config.Load(filepath.Join(t.TempDir(), "nonexistent.yaml"))
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if cfg.ProjectID != "" {
			t.Errorf("expected empty ProjectID when only unprefixed vars set, got %s", cfg.ProjectID)
		}
		if cfg.TelemetryTable != "" {
			t.Errorf("expected empty TelemetryTable when only unprefixed vars set, got %s", cfg.TelemetryTable)
		}
		if cfg.BillingTable != "" {
			t.Errorf("expected empty BillingTable when only unprefixed vars set, got %s", cfg.BillingTable)
		}
		if cfg.SeatQuota != 10 {
			t.Errorf("expected default seat quota 10, got %d", cfg.SeatQuota)
		}
		if cfg.Port != 8080 {
			t.Errorf("expected default port 8080, got %d", cfg.Port)
		}
		if !cfg.MCPEnabled {
			t.Errorf("expected MCPEnabled to remain true when only unprefixed _MCP_ENABLED/MCP_ENABLED are set")
		}
	})

	t.Run("mcp_enabled defaults to true and respects AGY_COST_BOARD_MCP_ENABLED override", func(t *testing.T) {
		cfgDefault, err := config.Load(filepath.Join(t.TempDir(), "nonexistent.yaml"))
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if !cfgDefault.MCPEnabled {
			t.Errorf("expected MCPEnabled to default to true, got false")
		}

		t.Run("disabled via AGY_COST_BOARD_MCP_ENABLED=false", func(t *testing.T) {
			t.Setenv("AGY_COST_BOARD_MCP_ENABLED", "false")
			cfg, err := config.Load(filepath.Join(t.TempDir(), "nonexistent.yaml"))
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if cfg.MCPEnabled {
				t.Errorf("expected MCPEnabled to be false when AGY_COST_BOARD_MCP_ENABLED=false")
			}
		})
	})

	t.Run("monitored_projects parses comma-separated env var and deduplicates", func(t *testing.T) {
		t.Setenv("AGY_COST_BOARD_MONITORED_PROJECTS", "proj-a, proj-b,proj-a, proj-c ")
		cfg, err := config.Load(filepath.Join(t.TempDir(), "nonexistent.yaml"))
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(cfg.MonitoredProjects) != 3 {
			t.Fatalf("expected 3 monitored projects, got %v", cfg.MonitoredProjects)
		}
		if cfg.MonitoredProjects[0] != "proj-a" || cfg.MonitoredProjects[1] != "proj-b" || cfg.MonitoredProjects[2] != "proj-c" {
			t.Errorf("unexpected monitored projects: %v", cfg.MonitoredProjects)
		}
	})
}
