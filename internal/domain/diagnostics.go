package domain

import (
	"fmt"
	"time"
)

// CheckStatus represents the status of a specific diagnostic check.
type CheckStatus string

const (
	StatusOK      CheckStatus = "OK"
	StatusWarning CheckStatus = "WARNING"
	StatusError   CheckStatus = "ERROR"
)

// CheckResult contains the result and metadata for a single diagnostic step.
type CheckResult struct {
	ID                 string         `json:"id"`
	Name               string         `json:"name"`
	Description        string         `json:"description,omitempty"`
	Status             CheckStatus    `json:"status"`
	Message            string         `json:"message"`
	Duration           time.Duration  `json:"duration"`
	RemediationCommand string         `json:"remediation_command,omitempty"`
	Details            map[string]any `json:"details,omitempty"`
}

// DiagnosticReport aggregates all diagnostic check results for a GCP project.
type DiagnosticReport struct {
	ProjectID      string        `json:"project_id"`
	Timestamp      time.Time     `json:"timestamp"`
	OverallStatus  CheckStatus   `json:"overall_status"`
	Checks         []CheckResult `json:"checks"`
	PassedCount    int           `json:"passed_count"`
	WarningCount   int           `json:"warning_count"`
	ErrorCount     int           `json:"error_count"`
}

// NewDiagnosticReport initializes an empty diagnostic report for the given project.
func NewDiagnosticReport(projectID string) *DiagnosticReport {
	return &DiagnosticReport{
		ProjectID:     projectID,
		Timestamp:     time.Now().UTC(),
		OverallStatus: StatusOK,
		Checks:        make([]CheckResult, 0),
	}
}

// AddCheck appends a check result and recalculates overall counts and status.
func (r *DiagnosticReport) AddCheck(c CheckResult) {
	r.Checks = append(r.Checks, c)
	switch c.Status {
	case StatusOK:
		r.PassedCount++
	case StatusWarning:
		r.WarningCount++
	case StatusError:
		r.ErrorCount++
	}

	r.EvaluateOverallStatus()
}

// EvaluateOverallStatus evaluates the worst status among all checks.
func (r *DiagnosticReport) EvaluateOverallStatus() {
	if r.ErrorCount > 0 {
		r.OverallStatus = StatusError
	} else if r.WarningCount > 0 {
		r.OverallStatus = StatusWarning
	} else {
		r.OverallStatus = StatusOK
	}
}

// DiagnosticTableHeaders returns table column headers for terminal output.
func DiagnosticTableHeaders() []string {
	return []string{"STATUS", "CHECK", "MESSAGE", "LATENCY"}
}

// ToTableRows returns tabular string rows for rendering via FormatTable.
func (r *DiagnosticReport) ToTableRows() [][]string {
	rows := make([][]string, 0, len(r.Checks))
	for _, c := range r.Checks {
		rows = append(rows, []string{
			string(c.Status),
			c.Name,
			c.Message,
			fmt.Sprintf("%dms", c.Duration.Milliseconds()),
		})
	}
	return rows
}
