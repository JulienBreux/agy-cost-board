package domain_test

import (
	"testing"
	"time"

	"github.com/julienbreux/agy-cost-board/internal/domain"
)

func TestDiagnosticReport_Evaluation(t *testing.T) {
	t.Run("empty report has status OK", func(t *testing.T) {
		report := domain.NewDiagnosticReport("test-project")
		if report.OverallStatus != domain.StatusOK {
			t.Errorf("expected StatusOK for empty report, got %s", report.OverallStatus)
		}
		if report.ProjectID != "test-project" {
			t.Errorf("expected test-project, got %s", report.ProjectID)
		}
	})

	t.Run("all OK checks result in overall OK", func(t *testing.T) {
		report := domain.NewDiagnosticReport("test-project")
		report.AddCheck(domain.CheckResult{
			ID:          "gcp_adc",
			Name:        "GCP Credentials",
			Status:      domain.StatusOK,
			Message:     "Valid ADC credentials found",
			Duration:    12 * time.Millisecond,
		})
		report.AddCheck(domain.CheckResult{
			ID:          "bq_dataset",
			Name:        "BigQuery Dataset",
			Status:      domain.StatusOK,
			Message:     "Dataset exists",
			Duration:    45 * time.Millisecond,
		})

		if report.OverallStatus != domain.StatusOK {
			t.Errorf("expected overall OK, got %s", report.OverallStatus)
		}
		if report.PassedCount != 2 || report.WarningCount != 0 || report.ErrorCount != 0 {
			t.Errorf("counts mismatch: passed=%d, warn=%d, err=%d", report.PassedCount, report.WarningCount, report.ErrorCount)
		}
	})

	t.Run("warning checks escalate status to warning", func(t *testing.T) {
		report := domain.NewDiagnosticReport("test-project")
		report.AddCheck(domain.CheckResult{
			ID:       "c1",
			Name:     "Check 1",
			Status:   domain.StatusOK,
			Duration: 5 * time.Millisecond,
		})
		report.AddCheck(domain.CheckResult{
			ID:                  "c2",
			Name:                "Check 2",
			Status:              domain.StatusWarning,
			Message:             "No logs ingested in last 7 days",
			RemediationCommand: "gcloud logging sinks describe ...",
			Duration:            10 * time.Millisecond,
		})

		if report.OverallStatus != domain.StatusWarning {
			t.Errorf("expected overall WARNING, got %s", report.OverallStatus)
		}
		if report.WarningCount != 1 || report.PassedCount != 1 {
			t.Errorf("unexpected counts: %v", report)
		}
	})

	t.Run("error checks escalate status to error over warning", func(t *testing.T) {
		report := domain.NewDiagnosticReport("test-project")
		report.AddCheck(domain.CheckResult{
			ID:       "c1",
			Name:     "Check 1",
			Status:   domain.StatusOK,
			Duration: 5 * time.Millisecond,
		})
		report.AddCheck(domain.CheckResult{
			ID:       "c2",
			Name:     "Check 2",
			Status:   domain.StatusWarning,
			Duration: 5 * time.Millisecond,
		})
		report.AddCheck(domain.CheckResult{
			ID:                  "c3",
			Name:                "Check 3",
			Status:              domain.StatusError,
			Message:             "Logging sink not found",
			RemediationCommand: "gcloud logging sinks create ...",
			Duration:            15 * time.Millisecond,
		})

		if report.OverallStatus != domain.StatusError {
			t.Errorf("expected overall ERROR, got %s", report.OverallStatus)
		}
		if report.ErrorCount != 1 || report.WarningCount != 1 || report.PassedCount != 1 {
			t.Errorf("unexpected counts: %v", report)
		}
	})

	t.Run("ToTableRows renders headers and rows correctly", func(t *testing.T) {
		report := domain.NewDiagnosticReport("my-project")
		report.AddCheck(domain.CheckResult{
			ID:       "adc",
			Name:     "GCP ADC",
			Status:   domain.StatusOK,
			Message:  "ADC active",
			Duration: 20 * time.Millisecond,
		})
		report.AddCheck(domain.CheckResult{
			ID:                 "sink",
			Name:               "Telemetry Sink",
			Status:             domain.StatusError,
			Message:            "Sink missing",
			RemediationCommand: "gcloud logging sinks create ...",
			Duration:           30 * time.Millisecond,
		})

		headers := domain.DiagnosticTableHeaders()
		if len(headers) != 4 {
			t.Errorf("expected 4 headers, got %d", len(headers))
		}

		rows := report.ToTableRows()
		if len(rows) != 2 {
			t.Fatalf("expected 2 table rows, got %d", len(rows))
		}
		if rows[0][0] != "OK" || rows[0][1] != "GCP ADC" {
			t.Errorf("unexpected first row: %v", rows[0])
		}
		if rows[1][0] != "ERROR" || rows[1][1] != "Telemetry Sink" {
			t.Errorf("unexpected second row: %v", rows[1])
		}
	})
}
