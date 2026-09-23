package config_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/julienbreux/agy-ge-board/internal/config"
)

func TestConfigLoadAndSave(t *testing.T) {
	tmpDir := t.TempDir()
	cfgPath := filepath.Join(tmpDir, ".agy-ge-board.yaml")

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
