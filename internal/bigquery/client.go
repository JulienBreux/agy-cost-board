package bigquery

import (
	"cmp"
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"cloud.google.com/go/bigquery"
	"github.com/julienbreux/agy-cost-board/internal/domain"
	"google.golang.org/api/iterator"
	"google.golang.org/api/option"
)

// ClientConfig holds configuration parameters for the Google BigQuery client.
type ClientConfig struct {
	ProjectID      string
	TelemetryTable string
	BillingTable   string
	SeatQuota      int
	CredentialsFile string
}

// Validate checks that required configuration fields are provided.
func (c ClientConfig) Validate() error {
	if strings.TrimSpace(c.ProjectID) == "" {
		return errors.New("project ID is required")
	}
	if strings.TrimSpace(c.TelemetryTable) == "" {
		return errors.New("telemetry table is required")
	}
	if strings.TrimSpace(c.BillingTable) == "" {
		return errors.New("billing table is required")
	}
	return nil
}

// BuildTelemetryQuery generates SQL to extract Antigravity inference logs with default payload field.
func BuildTelemetryQuery(table string, days int) string {
	return BuildTelemetryQueryForField(table, days, "jsonPayload")
}

// BuildTelemetryQueryForField generates SQL to extract Antigravity inference logs with a specified payload column.
func BuildTelemetryQueryForField(table string, days int, payloadField string) string {
	if days <= 0 {
		days = 30
	}
	payloadField = cmp.Or(payloadField, "jsonPayload")
	return fmt.Sprintf(`SELECT
  timestamp,
  IFNULL(resource.labels.project_id, '') AS project_id,
  labels.user_id AS user_id,
  labels.model AS model,
  IFNULL(CAST(%s.metadata.totalTokenCount AS INT64), 0) AS total_tokens,
  IFNULL(CAST(%s.metadata.promptTokenCount AS INT64), 0) AS prompt_tokens,
  IFNULL(CAST(%s.metadata.cachedContentTokenCount AS INT64), 0) AS cached_tokens,
  IFNULL(CAST(%s.metadata.candidatesTokenCount AS INT64), 0) AS completion_tokens,
  IFNULL(CAST(%s.metadata.thoughtsTokenCount AS INT64), 0) AS thoughts_tokens,
  IFNULL(%s.metadata.quotaType, 'REGULAR_QUOTA') AS quota_type,
  IFNULL(%s.experience, '') AS experience,
  IFNULL(labels.client_name, 'antigravity') AS client_name,
  IFNULL(labels.client_version, '') AS client_version
FROM
  `+"`%s`"+`
WHERE
  timestamp >= TIMESTAMP_SUB(CURRENT_TIMESTAMP(), INTERVAL %d DAY)
  AND %s.metadata.totalTokenCount IS NOT NULL
ORDER BY
  timestamp DESC`, payloadField, payloadField, payloadField, payloadField, payloadField, payloadField, payloadField, table, days, payloadField)
}

// BuildBillingQuery generates SQL to extract net costs for Gemini & Antigravity SKUs.
func BuildBillingQuery(table string, days int) string {
	if days <= 0 {
		days = 30
	}
	return fmt.Sprintf(`SELECT
  FORMAT_DATE('%%Y-%%m-%%d', usage_start_time) AS usage_date,
  IFNULL(project.id, '') AS project_id,
  sku.description AS sku_description,
  SUM(cost + IFNULL((SELECT SUM(c.amount) FROM UNNEST(credits) c), 0)) AS net_cost,
  currency
FROM
  `+"`%s`"+`
WHERE
  usage_start_time >= TIMESTAMP_SUB(CURRENT_TIMESTAMP(), INTERVAL %d DAY)
  AND (
    REGEXP_CONTAINS(sku.description, r'(?i)gemini|antigravity')
    OR REGEXP_CONTAINS(service.description, r'(?i)gemini|antigravity')
  )
GROUP BY
  usage_date, project_id, sku_description, currency
ORDER BY
  usage_date DESC`, table, days)
}

// BigQueryClient queries Google Cloud BigQuery for telemetry and billing exports.
type BigQueryClient struct {
	client *bigquery.Client
	config ClientConfig
}

// NewBigQueryClient constructs a new BigQuery client with ADC or service account credentials.
func NewBigQueryClient(ctx context.Context, cfg ClientConfig) (*BigQueryClient, error) {
	if err := cfg.Validate(); err != nil {
		return nil, err
	}

	var opts []option.ClientOption
	if cfg.CredentialsFile != "" {
		opts = append(opts, option.WithAuthCredentialsFile(option.ServiceAccount, cfg.CredentialsFile))
	}

	bqClient, err := bigquery.NewClient(ctx, cfg.ProjectID, opts...)
	if err != nil {
		return nil, fmt.Errorf("connect to bigquery: %w", err)
	}

	return &BigQueryClient{
		client: bqClient,
		config: cfg,
	}, nil
}

// Close closes the underlying BigQuery client connection.
func (c *BigQueryClient) Close() error {
	if c != nil && c.client != nil {
		return c.client.Close()
	}
	return nil
}

// TelemetryRow represents a raw record read from BigQuery inference logs.
type TelemetryRow struct {
	Timestamp        time.Time `bigquery:"timestamp"`
	ProjectID        string    `bigquery:"project_id"`
	UserID           string    `bigquery:"user_id"`
	Model            string    `bigquery:"model"`
	TotalTokens      int64     `bigquery:"total_tokens"`
	PromptTokens     int64     `bigquery:"prompt_tokens"`
	CachedTokens     int64     `bigquery:"cached_tokens"`
	CompletionTokens int64     `bigquery:"completion_tokens"`
	ThoughtsTokens   int64     `bigquery:"thoughts_tokens"`
	QuotaType        string    `bigquery:"quota_type"`
	Experience       string    `bigquery:"experience"`
	ClientName       string    `bigquery:"client_name"`
	ClientVersion    string    `bigquery:"client_version"`
}

// BillingRow represents a raw record read from BigQuery billing exports.
type BillingRow struct {
	UsageDate      string  `bigquery:"usage_date"`
	ProjectID      string  `bigquery:"project_id"`
	SKUDescription string  `bigquery:"sku_description"`
	NetCost        float64 `bigquery:"net_cost"`
	Currency       string  `bigquery:"currency"`
}

// MapTelemetryRow transforms a raw BigQuery telemetry row into a domain TelemetryLog.
func MapTelemetryRow(projectID string, row TelemetryRow) domain.TelemetryLog {
	effort := domain.ExtractEffortLevel(row.Experience)
	if effort == "default" {
		effort = domain.ExtractEffortLevel(row.Model)
	}
	quotaType := cmp.Or(strings.TrimSpace(row.QuotaType), "REGULAR_QUOTA")
	clientName := cmp.Or(strings.TrimSpace(row.ClientName), "antigravity")
	resolvedProject := cmp.Or(strings.TrimSpace(row.ProjectID), projectID)

	return domain.TelemetryLog{
		Timestamp:        row.Timestamp,
		ProjectID:        resolvedProject,
		UserID:           row.UserID,
		Model:            row.Model,
		TotalTokens:      row.TotalTokens,
		PromptTokens:     row.PromptTokens,
		CachedTokens:     row.CachedTokens,
		CompletionTokens: row.CompletionTokens,
		ThoughtsTokens:   row.ThoughtsTokens,
		QuotaType:        quotaType,
		EffortLevel:      effort,
		ClientName:       clientName,
		ClientVersion:    strings.TrimSpace(row.ClientVersion),
	}
}

// MapBillingRow transforms a raw BigQuery billing row into a domain BilledCost.
func MapBillingRow(row BillingRow) domain.BilledCost {
	return domain.BilledCost{
		UsageDate:      row.UsageDate,
		ProjectID:      strings.TrimSpace(row.ProjectID),
		SKUDescription: row.SKUDescription,
		NetCost:        row.NetCost,
		Currency:       row.Currency,
	}
}

// DeriveLicenseSeats calculates assigned seats, active usage, and dormant flags from logs.
func DeriveLicenseSeats(logs []domain.TelemetryLog, windowDays int, configuredQuota int) ([]domain.LicenseSeat, int) {
	userMap := make(map[string]*domain.LicenseSeat)
	defaultTier, defaultUSD, defaultWTU := domain.ResolveSeatAllowance("")
	for _, l := range logs {
		seat, exists := userMap[l.UserID]
		if !exists {
			seat = &domain.LicenseSeat{
				ProjectID:          l.ProjectID,
				UserID:             l.UserID,
				Assigned:           true,
				LastActivity:       l.Timestamp,
				Tier:               defaultTier,
				WeeklyAllowanceUSD: defaultUSD,
				WeeklyAllowanceWTU: defaultWTU,
				Location:           "global",
			}
			userMap[l.UserID] = seat
		} else if seat.ProjectID == "" && l.ProjectID != "" {
			seat.ProjectID = l.ProjectID
		}

		seat.TotalTokensInWindow += l.TotalTokens
		if l.Timestamp.After(seat.LastActivity) {
			seat.LastActivity = l.Timestamp
		}
	}

	seats := make([]domain.LicenseSeat, 0, len(userMap))
	for _, seat := range userMap {
		seat.Status = domain.ClassifySeatStatus(seat.TotalTokensInWindow, seat.LastActivity, windowDays)
		seats = append(seats, *seat)
	}

	quota := configuredQuota
	if quota <= 0 {
		quota = len(seats)
	}

	return seats, quota
}

// DetectPayloadField inspects the telemetry table schema to determine the root payload column name.
// Cloud Logging writes typed protobuf logs as jsonpayload_v1_inferenceresponselog,
// while untyped JSON logs are stored under jsonPayload.
func (c *BigQueryClient) DetectPayloadField(ctx context.Context) string {
	if c == nil || c.client == nil {
		return "jsonPayload"
	}
	project, rest, ok1 := strings.Cut(c.config.TelemetryTable, ".")
	if ok1 {
		dataset, table, ok2 := strings.Cut(rest, ".")
		if ok2 {
			md, err := c.client.DatasetInProject(project, dataset).Table(table).Metadata(ctx)
			if err == nil && md != nil {
				for _, f := range md.Schema {
					if strings.EqualFold(f.Name, "jsonpayload_v1_inferenceresponselog") {
						return "jsonpayload_v1_inferenceresponselog"
					}
				}
			}
		}
	}
	return "jsonPayload"
}

// FetchTelemetryLogs executes the telemetry query and returns individual events.
func (c *BigQueryClient) FetchTelemetryLogs(ctx context.Context, days int) ([]domain.TelemetryLog, error) {
	field := c.DetectPayloadField(ctx)
	sql := BuildTelemetryQueryForField(c.config.TelemetryTable, days, field)
	q := c.client.Query(sql)

	it, err := q.Read(ctx)
	if err != nil {
		return nil, fmt.Errorf("execute telemetry query: %w", err)
	}

	var logs []domain.TelemetryLog
	for {
		var row TelemetryRow
		err := it.Next(&row)
		if errors.Is(err, iterator.Done) {
			break
		}
		if err != nil {
			return nil, fmt.Errorf("read telemetry row: %w", err)
		}

		logs = append(logs, MapTelemetryRow(c.config.ProjectID, row))
	}

	return logs, nil
}

// FetchBilledCosts executes the billing query and returns net billed cost records.
func (c *BigQueryClient) FetchBilledCosts(ctx context.Context, days int) ([]domain.BilledCost, error) {
	sql := BuildBillingQuery(c.config.BillingTable, days)
	q := c.client.Query(sql)

	it, err := q.Read(ctx)
	if err != nil {
		return nil, fmt.Errorf("execute billing query: %w", err)
	}

	var costs []domain.BilledCost
	for {
		var row BillingRow
		err := it.Next(&row)
		if errors.Is(err, iterator.Done) {
			break
		}
		if err != nil {
			return nil, fmt.Errorf("read billing row: %w", err)
		}

		costs = append(costs, MapBillingRow(row))
	}

	return costs, nil
}

// FetchLicenseSeats derives assigned seat usage and dormant status from BigQuery telemetry.
func (c *BigQueryClient) FetchLicenseSeats(ctx context.Context, windowDays int) ([]domain.LicenseSeat, int, error) {
	logs, err := c.FetchTelemetryLogs(ctx, windowDays)
	if err != nil {
		return nil, 0, err
	}

	seats, quota := DeriveLicenseSeats(logs, windowDays, c.config.SeatQuota)
	return seats, quota, nil
}
