package bigquery

import (
	"context"

	"github.com/julienbreux/agy-ge-board/internal/domain"
)

// DataProvider defines the interface for retrieving telemetry, billing, and license data.
type DataProvider interface {
	FetchTelemetryLogs(ctx context.Context, days int) ([]domain.TelemetryLog, error)
	FetchBilledCosts(ctx context.Context, days int) ([]domain.BilledCost, error)
	FetchLicenseSeats(ctx context.Context, windowDays int) ([]domain.LicenseSeat, int, error)
}
