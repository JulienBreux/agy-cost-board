package setup_test

import (
	"context"
	"errors"
	"strings"
	"testing"

	"github.com/julienbreux/agy-cost-board/internal/domain"
	"github.com/julienbreux/agy-cost-board/internal/setup"
)

func TestADCChecker(t *testing.T) {
	t.Run("returns warning when ProjectID is empty", func(t *testing.T) {
		t.Setenv("GOOGLE_CLOUD_PROJECT", "")
		t.Setenv("GCP_PROJECT", "")
		cfg := setup.Config{
			ProjectID: "",
		}
		res := setup.CheckADC(t.Context(), cfg)
		if res.Status != domain.StatusWarning {
			t.Errorf("expected StatusWarning, got %s", res.Status)
		}
		if !strings.Contains(res.Message, "No GCP Project ID") {
			t.Errorf("expected missing project message, got: %s", res.Message)
		}
		if res.RemediationCommand == "" {
			t.Errorf("expected remediation command for missing project")
		}
	})

	t.Run("resolves project ID from GOOGLE_CLOUD_PROJECT", func(t *testing.T) {
		t.Setenv("GOOGLE_CLOUD_PROJECT", "env-project-1")
		t.Setenv("GCP_PROJECT", "")
		t.Setenv("GOOGLE_APPLICATION_CREDENTIALS", "")
		cfg := setup.Config{ProjectID: "", Demo: false}
		res := setup.CheckADC(t.Context(), cfg)
		if res.Status != domain.StatusOK {
			t.Errorf("expected StatusOK, got %s", res.Status)
		}
	})

	t.Run("resolves project ID from GCP_PROJECT", func(t *testing.T) {
		t.Setenv("GOOGLE_CLOUD_PROJECT", "")
		t.Setenv("GCP_PROJECT", "env-project-2")
		t.Setenv("GOOGLE_APPLICATION_CREDENTIALS", "")
		cfg := setup.Config{ProjectID: "", Demo: false}
		res := setup.CheckADC(t.Context(), cfg)
		if res.Status != domain.StatusOK {
			t.Errorf("expected StatusOK, got %s", res.Status)
		}
	})
	t.Run("returns OK with project ID in live mode", func(t *testing.T) {
		t.Setenv("GOOGLE_APPLICATION_CREDENTIALS", "")
		cfg := setup.Config{
			ProjectID: "my-live-project",
			Demo:      false,
		}
		res := setup.CheckADC(t.Context(), cfg)
		if res.Status != domain.StatusOK {
			t.Errorf("expected StatusOK, got %s", res.Status)
		}
	})

	t.Run("returns error when GOOGLE_APPLICATION_CREDENTIALS file is missing", func(t *testing.T) {
		t.Setenv("GOOGLE_APPLICATION_CREDENTIALS", "/non/existent/creds.json")
		cfg := setup.Config{
			ProjectID: "my-live-project",
			Demo:      false,
		}
		res := setup.CheckADC(t.Context(), cfg)
		if res.Status != domain.StatusError {
			t.Errorf("expected StatusError, got %s", res.Status)
		}
	})
}

func TestCheckIAM(t *testing.T) {
	t.Run("returns warning when ProjectID is empty", func(t *testing.T) {
		cfg := setup.Config{ProjectID: "", Demo: false}
		res := setup.CheckIAM(t.Context(), cfg)
		if res.Status != domain.StatusWarning {
			t.Errorf("expected StatusWarning, got %s", res.Status)
		}
	})

	t.Run("returns OK in demo mode", func(t *testing.T) {
		cfg := setup.Config{ProjectID: "demo-proj", Demo: true}
		res := setup.CheckIAM(t.Context(), cfg)
		if res.Status != domain.StatusOK {
			t.Errorf("expected StatusOK, got %s", res.Status)
		}
	})

	t.Run("returns OK in live mode with ProjectID", func(t *testing.T) {
		cfg := setup.Config{ProjectID: "live-proj", Demo: false}
		res := setup.CheckIAM(t.Context(), cfg)
		if res.Status != domain.StatusOK {
			t.Errorf("expected StatusOK, got %s", res.Status)
		}
	})
}

func TestTelemetrySinkChecker(t *testing.T) {
	t.Run("returns error with custom sink and dataset name fallback", func(t *testing.T) {
		cfg := setup.Config{
			ProjectID:      "custom-proj",
			TelemetryTable: "",
			SinkName:       "my-custom-sink",
			DatasetName:    "my_custom_dataset",
		}
		res := setup.CheckTelemetrySink(t.Context(), cfg)
		if res.Status != domain.StatusError {
			t.Errorf("expected StatusError, got %s", res.Status)
		}
		if !strings.Contains(res.RemediationCommand, "my-custom-sink") || !strings.Contains(res.RemediationCommand, "my_custom_dataset") {
			t.Errorf("expected custom sink/dataset in remediation command, got: %s", res.RemediationCommand)
		}
	})

	t.Run("returns error when telemetry table has invalid format", func(t *testing.T) {
		cfg := setup.Config{
			ProjectID:      "test-project",
			TelemetryTable: "invalid_table_format",
			Demo:           false,
		}
		res := setup.CheckTelemetrySink(t.Context(), cfg)
		if res.Status != domain.StatusError {
			t.Errorf("expected StatusError for invalid table format, got %s", res.Status)
		}
	})

	t.Run("returns OK when in demo mode with telemetry table", func(t *testing.T) {
		cfg := setup.Config{
			ProjectID:      "test-project",
			TelemetryTable: "test-project.antigravity.inference_logs",
			Demo:           true,
		}
		res := setup.CheckTelemetrySink(t.Context(), cfg)
		if res.Status != domain.StatusOK {
			t.Errorf("expected StatusOK in demo mode, got %s", res.Status)
		}
	})
}

func TestBillingExportChecker(t *testing.T) {
	t.Run("returns error when billing table is missing", func(t *testing.T) {
		cfg := setup.Config{
			ProjectID:    "test-project",
			BillingTable: "",
		}
		res := setup.CheckBillingExport(t.Context(), cfg)
		if res.Status != domain.StatusError {
			t.Errorf("expected StatusError for missing billing table, got %s", res.Status)
		}
		if !strings.Contains(res.Message, "Billing export table is not configured") {
			t.Errorf("unexpected message: %s", res.Message)
		}
	})

	t.Run("returns error when billing table has invalid format", func(t *testing.T) {
		cfg := setup.Config{
			ProjectID:    "test-project",
			BillingTable: "invalid_format",
			Demo:         false,
		}
		res := setup.CheckBillingExport(t.Context(), cfg)
		if res.Status != domain.StatusError {
			t.Errorf("expected StatusError for invalid format, got %s", res.Status)
		}
	})

	t.Run("returns OK when in demo mode with billing table", func(t *testing.T) {
		cfg := setup.Config{
			ProjectID:    "test-project",
			BillingTable: "test-project.billing.gcp_billing_export_v1_000",
			Demo:         true,
		}
		res := setup.CheckBillingExport(t.Context(), cfg)
		if res.Status != domain.StatusOK {
			t.Errorf("expected StatusOK in demo mode, got %s", res.Status)
		}
	})
}

func TestTelemetryPipelineChecker(t *testing.T) {
	t.Run("returns warning when telemetry table is not configured", func(t *testing.T) {
		cfg := setup.Config{
			ProjectID:      "test-project",
			TelemetryTable: "",
		}
		res := setup.CheckTelemetryPipeline(t.Context(), cfg)
		if res.Status != domain.StatusWarning {
			t.Errorf("expected StatusWarning, got %s", res.Status)
		}
	})

	t.Run("returns warning when telemetry table has invalid format", func(t *testing.T) {
		cfg := setup.Config{
			ProjectID:      "test-project",
			TelemetryTable: "single_part",
			Demo:           false,
		}
		res := setup.CheckTelemetryPipeline(t.Context(), cfg)
		if res.Status != domain.StatusWarning {
			t.Errorf("expected StatusWarning for invalid table, got %s", res.Status)
		}
	})

	t.Run("returns OK in demo mode", func(t *testing.T) {
		cfg := setup.Config{
			ProjectID:      "test-project",
			TelemetryTable: "test-project.antigravity.logs",
			Demo:           true,
		}
		res := setup.CheckTelemetryPipeline(t.Context(), cfg)
		if res.Status != domain.StatusOK {
			t.Errorf("expected StatusOK in demo mode, got %s", res.Status)
		}
		if !strings.Contains(res.Message, "InferenceResponseLog records verified") {
			t.Errorf("unexpected message: %s", res.Message)
		}
	})
}

func TestDiagnosticRunner(t *testing.T) {
	t.Run("RunAll aggregates all checks into DiagnosticReport", func(t *testing.T) {
		runner := setup.NewRunner()
		cfg := setup.Config{
			ProjectID:      "demo-project",
			TelemetryTable: "demo-project.telemetry.inference_logs",
			BillingTable:   "demo-project.billing.gcp_billing_export_v1_123",
			Demo:           true,
		}

		report := runner.RunAll(t.Context(), cfg)
		if report == nil {
			t.Fatalf("expected non-nil report")
		}
		if report.OverallStatus != domain.StatusOK {
			t.Errorf("expected overall OK in demo mode, got %s", report.OverallStatus)
		}
		if len(report.Checks) != 5 {
			t.Errorf("expected 5 checks in report, got %d", len(report.Checks))
		}
		if report.PassedCount != 5 {
			t.Errorf("expected 5 passed checks, got %d", report.PassedCount)
		}
	})
}

type mockInspector struct {
	checkTableFn func(ctx context.Context, projectID, datasetID, tableID string) (int64, error)
	queryCountFn func(ctx context.Context, projectID, query string) (int64, error)
}

func (m *mockInspector) CheckTable(ctx context.Context, projectID, datasetID, tableID string) (int64, error) {
	if m.checkTableFn != nil {
		return m.checkTableFn(ctx, projectID, datasetID, tableID)
	}
	return 100, nil
}

func (m *mockInspector) QueryCount(ctx context.Context, projectID, query string) (int64, error) {
	if m.queryCountFn != nil {
		return m.queryCountFn(ctx, projectID, query)
	}
	return 50, nil
}

func TestTelemetrySinkChecker_Inspector(t *testing.T) {
	t.Run("returns OK when inspector succeeds", func(t *testing.T) {
		cfg := setup.Config{
			ProjectID:      "test-proj",
			TelemetryTable: "test-proj.telemetry.logs",
			Demo:           false,
			Inspector: &mockInspector{
				checkTableFn: func(ctx context.Context, p, d, table string) (int64, error) {
					return 1500, nil
				},
			},
		}
		res := setup.CheckTelemetrySink(t.Context(), cfg)
		if res.Status != domain.StatusOK {
			t.Errorf("expected StatusOK, got %s", res.Status)
		}
	})

	t.Run("returns warning when inspector returns error", func(t *testing.T) {
		cfg := setup.Config{
			ProjectID:      "test-proj",
			TelemetryTable: "test-proj.telemetry.logs",
			Demo:           false,
			Inspector: &mockInspector{
				checkTableFn: func(ctx context.Context, p, d, table string) (int64, error) {
					return 0, errors.New("dataset not found")
				},
			},
		}
		res := setup.CheckTelemetrySink(t.Context(), cfg)
		if res.Status != domain.StatusWarning {
			t.Errorf("expected StatusWarning, got %s", res.Status)
		}
	})
}

func TestBillingExportChecker_Inspector(t *testing.T) {
	t.Run("returns OK when inspector succeeds", func(t *testing.T) {
		cfg := setup.Config{
			ProjectID:    "test-proj",
			BillingTable: "test-proj.billing.gcp_billing_export_v1_123",
			Demo:         false,
			Inspector: &mockInspector{
				checkTableFn: func(ctx context.Context, p, d, table string) (int64, error) {
					return 400, nil
				},
			},
		}
		res := setup.CheckBillingExport(t.Context(), cfg)
		if res.Status != domain.StatusOK {
			t.Errorf("expected StatusOK, got %s", res.Status)
		}
	})

	t.Run("returns warning when inspector returns error", func(t *testing.T) {
		cfg := setup.Config{
			ProjectID:    "test-proj",
			BillingTable: "test-proj.billing.gcp_billing_export_v1_123",
			Demo:         false,
			Inspector: &mockInspector{
				checkTableFn: func(ctx context.Context, p, d, table string) (int64, error) {
					return 0, errors.New("billing dataset permission denied")
				},
			},
		}
		res := setup.CheckBillingExport(t.Context(), cfg)
		if res.Status != domain.StatusWarning {
			t.Errorf("expected StatusWarning, got %s", res.Status)
		}
	})
}

func TestTelemetryPipelineChecker_Inspector(t *testing.T) {
	t.Run("returns OK when logs are found", func(t *testing.T) {
		cfg := setup.Config{
			ProjectID:      "test-proj",
			TelemetryTable: "test-proj.telemetry.logs",
			Demo:           false,
			Inspector: &mockInspector{
				queryCountFn: func(ctx context.Context, p, q string) (int64, error) {
					return 250, nil
				},
			},
		}
		res := setup.CheckTelemetryPipeline(t.Context(), cfg)
		if res.Status != domain.StatusOK {
			t.Errorf("expected StatusOK, got %s", res.Status)
		}
	})

	t.Run("returns warning when 0 logs are found", func(t *testing.T) {
		cfg := setup.Config{
			ProjectID:      "test-proj",
			TelemetryTable: "test-proj.telemetry.logs",
			Demo:           false,
			Inspector: &mockInspector{
				queryCountFn: func(ctx context.Context, p, q string) (int64, error) {
					return 0, nil
				},
			},
		}
		res := setup.CheckTelemetryPipeline(t.Context(), cfg)
		if res.Status != domain.StatusWarning {
			t.Errorf("expected StatusWarning for 0 logs, got %s", res.Status)
		}
	})

	t.Run("returns warning when query fails", func(t *testing.T) {
		cfg := setup.Config{
			ProjectID:      "test-proj",
			TelemetryTable: "test-proj.telemetry.logs",
			Demo:           false,
			Inspector: &mockInspector{
				queryCountFn: func(ctx context.Context, p, q string) (int64, error) {
					return 0, errors.New("query syntax error")
				},
			},
		}
		res := setup.CheckTelemetryPipeline(t.Context(), cfg)
		if res.Status != domain.StatusWarning {
			t.Errorf("expected StatusWarning on query failure, got %s", res.Status)
		}
	})
}
