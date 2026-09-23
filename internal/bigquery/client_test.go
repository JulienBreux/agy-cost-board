package bigquery_test

import (
	"context"
	"strings"
	"testing"
	"time"

	"github.com/julienbreux/agy-ge-board/internal/bigquery"
	"github.com/julienbreux/agy-ge-board/internal/domain"
)

func TestQueryBuilders(t *testing.T) {
	t.Run("BuildTelemetryQuery generates valid SQL with date interval", func(t *testing.T) {
		table := "my-project.antigravity_logs.inference_logs"
		query := bigquery.BuildTelemetryQuery(table, 14)

		if !strings.Contains(query, "`my-project.antigravity_logs.inference_logs`") {
			t.Errorf("query does not contain backtick-escaped table name: %s", query)
		}
		if !strings.Contains(query, "INTERVAL 14 DAY") {
			t.Errorf("query does not contain correct interval: %s", query)
		}
		if !strings.Contains(query, "labels.user_id") {
			t.Errorf("query missing user_id extraction: %s", query)
		}

		// Test default days <= 0 fallback
		defaultQuery := bigquery.BuildTelemetryQuery(table, 0)
		if !strings.Contains(defaultQuery, "INTERVAL 30 DAY") {
			t.Errorf("query does not default to 30 days: %s", defaultQuery)
		}
	})

	t.Run("BuildBillingQuery generates valid SQL for Gemini and Antigravity SKUs", func(t *testing.T) {
		table := "my-project.billing.gcp_billing_export_v1_000000"
		query := bigquery.BuildBillingQuery(table, 30)

		if !strings.Contains(query, "`my-project.billing.gcp_billing_export_v1_000000`") {
			t.Errorf("query does not contain billing table name: %s", query)
		}
		if !strings.Contains(query, "INTERVAL 30 DAY") {
			t.Errorf("query does not contain correct interval: %s", query)
		}
		if !strings.Contains(query, "REGEXP_CONTAINS") {
			t.Errorf("query missing regex filter for SKUs: %s", query)
		}
		if !strings.Contains(query, "GROUP BY") {
			t.Errorf("query missing GROUP BY: %s", query)
		}

		// Test default days <= 0 fallback
		defaultQuery := bigquery.BuildBillingQuery(table, -5)
		if !strings.Contains(defaultQuery, "INTERVAL 30 DAY") {
			t.Errorf("query does not default to 30 days: %s", defaultQuery)
		}
	})
}

func TestClientConfigValidation(t *testing.T) {
	tests := []struct {
		name        string
		cfg         bigquery.ClientConfig
		expectError bool
	}{
		{
			name: "Valid configuration",
			cfg: bigquery.ClientConfig{
				ProjectID:      "test-project",
				TelemetryTable: "test-project.logs.inference",
				BillingTable:   "test-project.billing.export",
				SeatQuota:      10,
			},
			expectError: false,
		},
		{
			name: "Missing project ID",
			cfg: bigquery.ClientConfig{
				TelemetryTable: "test-project.logs.inference",
				BillingTable:   "test-project.billing.export",
			},
			expectError: true,
		},
		{
			name: "Missing telemetry table",
			cfg: bigquery.ClientConfig{
				ProjectID:    "test-project",
				BillingTable: "test-project.billing.export",
			},
			expectError: true,
		},
		{
			name: "Missing billing table",
			cfg: bigquery.ClientConfig{
				ProjectID:      "test-project",
				TelemetryTable: "test-project.logs.inference",
			},
			expectError: true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			err := tc.cfg.Validate()
			if tc.expectError && err == nil {
				t.Errorf("expected validation error but got nil")
			}
			if !tc.expectError && err != nil {
				t.Errorf("unexpected validation error: %v", err)
			}
		})
	}
}

func TestNewBigQueryClientInvalid(t *testing.T) {
	ctx := context.Background()
	cfg := bigquery.ClientConfig{} // missing required fields

	client, err := bigquery.NewBigQueryClient(ctx, cfg)
	if err == nil {
		t.Errorf("expected error when initializing client with empty config, got nil")
	}
	if client != nil {
		t.Errorf("expected nil client on error")
	}

	// Test Close on nil client
	var emptyClient *bigquery.BigQueryClient
	if err := emptyClient.Close(); err != nil {
		t.Errorf("unexpected error on nil client close: %v", err)
	}
}

func TestRowMappers(t *testing.T) {
	now := time.Now().UTC()

	t.Run("MapTelemetryRow maps fields accurately", func(t *testing.T) {
		raw := bigquery.TelemetryRow{
			Timestamp:        now,
			UserID:           "dev@example.com",
			Model:            "gemini-1.5-pro",
			TotalTokens:      1500,
			PromptTokens:     1000,
			CompletionTokens: 500,
		}

		mapped := bigquery.MapTelemetryRow("my-prj", raw)
		if mapped.ProjectID != "my-prj" {
			t.Errorf("expected project ID my-prj, got %s", mapped.ProjectID)
		}
		if mapped.UserID != "dev@example.com" {
			t.Errorf("expected dev@example.com, got %s", mapped.UserID)
		}
		if mapped.TotalTokens != 1500 {
			t.Errorf("expected 1500 total tokens, got %d", mapped.TotalTokens)
		}
	})

	t.Run("MapBillingRow maps fields accurately", func(t *testing.T) {
		raw := bigquery.BillingRow{
			UsageDate:      "2026-09-22",
			SKUDescription: "Gemini 1.5 Pro",
			NetCost:        12.34,
			Currency:       "USD",
		}

		mapped := bigquery.MapBillingRow(raw)
		if mapped.UsageDate != "2026-09-22" {
			t.Errorf("expected date 2026-09-22, got %s", mapped.UsageDate)
		}
		if mapped.NetCost != 12.34 {
			t.Errorf("expected net cost 12.34, got %f", mapped.NetCost)
		}
	})
}

func TestDeriveLicenseSeats(t *testing.T) {
	now := time.Now().UTC()
	logs := []domain.TelemetryLog{
		{
			Timestamp:   now.Add(-2 * time.Hour),
			UserID:      "active@example.com",
			TotalTokens: 5000,
		},
		{
			Timestamp:   now.Add(-10 * time.Hour),
			UserID:      "active@example.com",
			TotalTokens: 2000,
		},
		{
			Timestamp:   now.AddDate(0, 0, -5),
			UserID:      "atrisk@example.com",
			TotalTokens: 100,
		},
	}

	seats, quota := bigquery.DeriveLicenseSeats(logs, 30, 10)
	if quota != 10 {
		t.Errorf("expected quota 10, got %d", quota)
	}
	if len(seats) != 2 {
		t.Errorf("expected 2 derived seats, got %d", len(seats))
	}

	// Test default quota when <= 0
	_, fallbackQuota := bigquery.DeriveLicenseSeats(logs, 30, 0)
	if fallbackQuota != 2 {
		t.Errorf("expected fallback quota to equal len(seats)=2, got %d", fallbackQuota)
	}
}
