package observability

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"log/slog"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestParseCloudTraceHeader(t *testing.T) {
	traceID, spanID, sampled := ParseCloudTraceHeader("105445aa7843bc8bf206b12000100000/1;o=1")
	assert.Equal(t, "105445aa7843bc8bf206b12000100000", traceID)
	assert.Equal(t, "0000000000000001", spanID)
	assert.True(t, sampled)

	traceID, spanID, sampled = ParseCloudTraceHeader("105445aa7843bc8bf206b12000100000/abcdef0123456789;o=0")
	assert.Equal(t, "105445aa7843bc8bf206b12000100000", traceID)
	assert.Equal(t, "abcdef0123456789", spanID)
	assert.False(t, sampled)

	traceID, spanID, sampled = ParseCloudTraceHeader("invalid-header")
	assert.Empty(t, traceID)
	assert.Empty(t, spanID)
	assert.False(t, sampled)
}

func TestParseTraceparentHeader(t *testing.T) {
	traceID, spanID, sampled := ParseTraceparentHeader("00-4bf92f3577b34da6a3ce929d0e0e4736-00f067aa0ba902b7-01")
	assert.Equal(t, "4bf92f3577b34da6a3ce929d0e0e4736", traceID)
	assert.Equal(t, "00f067aa0ba902b7", spanID)
	assert.True(t, sampled)

	traceID, spanID, sampled = ParseTraceparentHeader("00-4bf92f3577b34da6a3ce929d0e0e4736-00f067aa0ba902b7-00")
	assert.Equal(t, "4bf92f3577b34da6a3ce929d0e0e4736", traceID)
	assert.Equal(t, "00f067aa0ba902b7", spanID)
	assert.False(t, sampled)

	traceID, _, _ = ParseTraceparentHeader("bad")
	assert.Empty(t, traceID)
}

func TestJSONLoggerAndSpanCorrelation(t *testing.T) {
	var buf bytes.Buffer
	InitLogger(Config{
		Level:     "debug",
		Format:    "json",
		ProjectID: "julienbreux-run",
		Output:    &buf,
	})

	rootCtx := WithTrace(context.Background(), TraceInfo{
		TraceID:    "105445aa7843bc8bf206b12000100000",
		SpanID:     "0000000000000001",
		Sampled:    true,
		RequestID:  "req-abc-123",
		EntryPoint: "http",
		ProjectID:  "julienbreux-run",
	})

	spanCtx, span := StartSpan(rootCtx, "bigquery.fetch_telemetry", slog.Int("days", 7))
	span.AddAttrs(slog.Int("row_count", 42))
	span.End(nil)

	_, errSpan := StartSpan(spanCtx, "bigquery.fetch_billing", slog.Int("days", 7))
	errSpan.End(errors.New("table not found"))

	lines := strings.Split(strings.TrimSpace(buf.String()), "\n")
	require.Len(t, lines, 2)

	var first map[string]any
	require.NoError(t, json.Unmarshal([]byte(lines[0]), &first))
	assert.Equal(t, "DEBUG", first["severity"])
	assert.Equal(t, "span_completed", first["event"])
	assert.Equal(t, "bigquery.fetch_telemetry", first["span_name"])
	assert.Equal(t, "ok", first["status"])
	assert.Equal(t, "req-abc-123", first["request_id"])
	assert.Equal(t, "http", first["entry_point"])
	assert.Equal(t, "projects/julienbreux-run/traces/105445aa7843bc8bf206b12000100000", first["logging.googleapis.com/trace"])
	assert.Equal(t, "0000000000000001", first["parent_span_id"])
	assert.Equal(t, true, first["logging.googleapis.com/trace_sampled"])
	assert.Equal(t, float64(42), first["row_count"])

	var second map[string]any
	require.NoError(t, json.Unmarshal([]byte(lines[1]), &second))
	assert.Equal(t, "ERROR", second["severity"])
	assert.Equal(t, "bigquery.fetch_billing", second["span_name"])
	assert.Equal(t, "error", second["status"])
	assert.Equal(t, "table not found", second["error"])
}
