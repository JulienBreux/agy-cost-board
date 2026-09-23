package setup

import (
	"context"

	"github.com/julienbreux/agy-ge-board/internal/domain"
)

// Runner coordinates execution of diagnostic setup checks.
type Runner interface {
	RunAll(ctx context.Context, cfg Config) *domain.DiagnosticReport
}

type defaultRunner struct{}

// NewRunner creates a new diagnostic check runner.
func NewRunner() Runner {
	return &defaultRunner{}
}

// RunAll executes the complete series of 5 setup checks in order and returns an aggregated report.
func (r *defaultRunner) RunAll(ctx context.Context, cfg Config) *domain.DiagnosticReport {
	report := domain.NewDiagnosticReport(cfg.ProjectID)

	// Check 1: ADC & Project
	report.AddCheck(CheckADC(ctx, cfg))

	// Check 2: IAM Permissions
	report.AddCheck(CheckIAM(ctx, cfg))

	// Check 3: Telemetry Sink & Dataset
	report.AddCheck(CheckTelemetrySink(ctx, cfg))

	// Check 4: Cloud Billing Export
	report.AddCheck(CheckBillingExport(ctx, cfg))

	// Check 5: Pipeline & Ingestion probe
	report.AddCheck(CheckTelemetryPipeline(ctx, cfg))

	return report
}
