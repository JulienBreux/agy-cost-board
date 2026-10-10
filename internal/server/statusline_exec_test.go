package server_test

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func TestStatuslineBashSyntax(t *testing.T) {
	router := setupTestServer(t)
	req := httptest.NewRequest(http.MethodGet, "/statusline.sh", nil)
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200 OK, got %d", rec.Code)
	}

	tmpDir := t.TempDir()
	scriptPath := filepath.Join(tmpDir, "statusline.sh")
	if err := os.WriteFile(scriptPath, rec.Body.Bytes(), 0755); err != nil {
		t.Fatalf("failed to write script: %v", err)
	}

	cmd := exec.Command("bash", "-n", scriptPath)
	output, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("bash -n syntax validation failed: %v, output: %s", err, string(output))
	}
}

func TestStatuslineScriptExecution(t *testing.T) {
	// Mock backend API serving /api/v1/users/{id}
	mockServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if strings.HasPrefix(r.URL.Path, "/api/v1/users/") {
			w.Header().Set("Content-Type", "application/json")
			_ = json.NewEncoder(w).Encode(map[string]interface{}{
				"user_id":    "alice@example.com",
				"total_cost": 123.45,
				"currency":   "USD",
			})
			return
		}
		http.NotFound(w, r)
	}))
	defer mockServer.Close()

	router := setupTestServer(t)
	req := httptest.NewRequest(http.MethodGet, "/statusline.sh?user=alice@example.com", nil)
	req.Host = strings.TrimPrefix(mockServer.URL, "http://")
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200 OK, got %d", rec.Code)
	}

	tmpDir := t.TempDir()
	scriptPath := filepath.Join(tmpDir, "statusline.sh")
	if err := os.WriteFile(scriptPath, rec.Body.Bytes(), 0755); err != nil {
		t.Fatalf("failed to write script: %v", err)
	}

	stdinJSON := `{
		"agent_state": "idle",
		"context_window": { "used_percentage": 25.5 },
		"vcs": { "branch": "main", "dirty": false },
		"sandbox": { "enabled": false },
		"artifact_count": 3,
		"subagents": [],
		"task_count": 2,
		"model": { "display_name": "gemini-1.5-pro" },
		"terminal_width": 100
	}`

	t.Run("Script outputs cost when cache exists", func(t *testing.T) {
		cacheFile := "/tmp/agy_cost_alice_example_com.cache"
		_ = os.WriteFile(cacheFile, []byte("123\n"), 0644)
		defer os.Remove(cacheFile)

		cmd := exec.Command("bash", scriptPath)
		cmd.Stdin = strings.NewReader(stdinJSON)
		cmd.Env = append(os.Environ(), "AGY_COST_BOARD_URL="+mockServer.URL)

		var stdout, stderr bytes.Buffer
		cmd.Stdout = &stdout
		cmd.Stderr = &stderr

		if err := cmd.Run(); err != nil {
			t.Fatalf("script execution failed: %v, stderr: %s", err, stderr.String())
		}

		outStr := stdout.String()
		if !strings.Contains(outStr, "cost") || !strings.Contains(outStr, "$123") {
			t.Errorf("expected output to contain 'cost' and '$123', got: %q", outStr)
		}
	})

	t.Run("Script respects AGY_COST_DISABLED=true", func(t *testing.T) {
		cacheFile := "/tmp/agy_cost_alice_example_com.cache"
		_ = os.WriteFile(cacheFile, []byte("123\n"), 0644)
		defer os.Remove(cacheFile)

		cmd := exec.Command("bash", scriptPath)
		cmd.Stdin = strings.NewReader(stdinJSON)
		cmd.Env = append(os.Environ(), "AGY_COST_BOARD_URL="+mockServer.URL, "AGY_COST_DISABLED=true")

		var stdout, stderr bytes.Buffer
		cmd.Stdout = &stdout
		cmd.Stderr = &stderr

		if err := cmd.Run(); err != nil {
			t.Fatalf("script execution failed: %v, stderr: %s", err, stderr.String())
		}

		outStr := stdout.String()
		if strings.Contains(outStr, "cost") {
			t.Errorf("expected cost to be hidden when AGY_COST_DISABLED=true, got: %q", outStr)
		}
	})

	t.Run("Script triggers background fetch when cache is missing", func(t *testing.T) {
		cacheFile := "/tmp/agy_cost_alice_example_com.cache"
		lockFile := "/tmp/agy_cost_alice_example_com.lock"
		_ = os.Remove(cacheFile)
		_ = os.Remove(lockFile)
		defer os.Remove(cacheFile)
		defer os.Remove(lockFile)

		cmd := exec.Command("bash", scriptPath)
		cmd.Stdin = strings.NewReader(stdinJSON)
		cmd.Env = append(os.Environ(), "AGY_COST_BOARD_URL="+mockServer.URL)

		var stdout, stderr bytes.Buffer
		cmd.Stdout = &stdout
		cmd.Stderr = &stderr

		if err := cmd.Run(); err != nil {
			t.Fatalf("script execution failed: %v, stderr: %s", err, stderr.String())
		}

		// Wait briefly for background subshell to complete fetch and write cache
		var cachePopulated bool
		for i := 0; i < 20; i++ {
			time.Sleep(100 * time.Millisecond)
			if data, err := os.ReadFile(cacheFile); err == nil && len(bytes.TrimSpace(data)) > 0 {
				val := string(bytes.TrimSpace(data))
				if val == "123" {
					cachePopulated = true
					break
				}
			}
		}

		if !cachePopulated {
			t.Errorf("expected background fetch to populate cache file %s with '123'", cacheFile)
		}
	})

	t.Run("Script respects AGY_COST_USER override over DEFAULT_USER", func(t *testing.T) {
		cacheBob := "/tmp/agy_cost_bob_example_com.cache"
		_ = os.WriteFile(cacheBob, []byte("456\n"), 0644)
		defer os.Remove(cacheBob)

		cmd := exec.Command("bash", scriptPath)
		cmd.Stdin = strings.NewReader(stdinJSON)
		cmd.Env = append(os.Environ(),
			"AGY_COST_BOARD_URL="+mockServer.URL,
			"AGY_COST_USER=bob@example.com",
		)

		var stdout bytes.Buffer
		cmd.Stdout = &stdout

		if err := cmd.Run(); err != nil {
			t.Fatalf("script execution failed: %v", err)
		}

		outStr := stdout.String()
		if !strings.Contains(outStr, "$456") {
			t.Errorf("expected output to contain '$456' for AGY_COST_USER, got: %q", outStr)
		}
	})

	t.Run("Script attaches IAM Bearer token from gcloud auth print-identity-token", func(t *testing.T) {
		var authHeaderReceived string
		authServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			authHeaderReceived = r.Header.Get("Authorization")
			w.Header().Set("Content-Type", "application/json")
			_ = json.NewEncoder(w).Encode(map[string]interface{}{
				"user_id":    "alice@example.com",
				"total_cost": 50.0,
				"currency":   "USD",
			})
		}))
		defer authServer.Close()

		mockBinDir := t.TempDir()
		mockGcloud := filepath.Join(mockBinDir, "gcloud")
		_ = os.WriteFile(mockGcloud, []byte("#!/bin/sh\nif [ \"$1\" = \"auth\" ]; then echo \"mock-iam-token-xyz\"; fi\n"), 0755)

		cacheFile := "/tmp/agy_cost_alice_example_com.cache"
		lockFile := "/tmp/agy_cost_alice_example_com.lock"
		_ = os.Remove(cacheFile)
		_ = os.Remove(lockFile)
		defer os.Remove(cacheFile)
		defer os.Remove(lockFile)

		cmd := exec.Command("bash", scriptPath)
		cmd.Stdin = strings.NewReader(stdinJSON)
		cmd.Env = append(os.Environ(),
			"AGY_COST_BOARD_URL="+authServer.URL,
			"PATH="+mockBinDir+":"+os.Getenv("PATH"),
		)

		if err := cmd.Run(); err != nil {
			t.Fatalf("script execution failed: %v", err)
		}

		// Wait for subshell to make the authenticated request
		for i := 0; i < 20; i++ {
			time.Sleep(100 * time.Millisecond)
			if authHeaderReceived != "" {
				break
			}
		}

		if authHeaderReceived != "Bearer mock-iam-token-xyz" {
			t.Errorf("expected Authorization 'Bearer mock-iam-token-xyz', got %q", authHeaderReceived)
		}
	})

	t.Run("Script handles server error gracefully and fails open", func(t *testing.T) {
		errorServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			http.Error(w, "not found", http.StatusNotFound)
		}))
		defer errorServer.Close()

		cacheFile := "/tmp/agy_cost_unknown_example_com.cache"
		lockFile := "/tmp/agy_cost_unknown_example_com.lock"
		_ = os.Remove(cacheFile)
		_ = os.Remove(lockFile)
		defer os.Remove(cacheFile)
		defer os.Remove(lockFile)

		cmd := exec.Command("bash", scriptPath)
		cmd.Stdin = strings.NewReader(stdinJSON)
		cmd.Env = append(os.Environ(),
			"AGY_COST_BOARD_URL="+errorServer.URL,
			"AGY_COST_USER=unknown@example.com",
		)

		var stdout bytes.Buffer
		cmd.Stdout = &stdout

		if err := cmd.Run(); err != nil {
			t.Fatalf("script should fail-open without error, got: %v", err)
		}
	})
}
