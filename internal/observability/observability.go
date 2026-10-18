package observability

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"io"
	"log/slog"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"
)

type traceCtxKey struct{}

// TraceInfo holds distributed tracing and request correlation metadata carried in context.Context.
type TraceInfo struct {
	TraceID      string
	SpanID       string
	ParentSpanID string
	Sampled      bool
	RequestID    string
	EntryPoint   string
	ProjectID    string
}

// Config configures structured logging and distributed tracing behavior.
type Config struct {
	Level     string
	Format    string
	Service   string
	Version   string
	ProjectID string
	Output    io.Writer
}

var (
	globalMu        sync.RWMutex
	globalProjectID string
)

// InitLogger initializes the global default slog.Logger with Google Cloud Logging compatible keys
// when JSON formatting is enabled, or human-readable text formatting for local CLI usage.
func InitLogger(cfg Config) *slog.Logger {
	out := cfg.Output
	if out == nil {
		out = os.Stderr
	}

	level := ParseLevel(cfg.Level)
	format := strings.ToLower(strings.TrimSpace(cfg.Format))
	if format == "" || format == "auto" {
		if os.Getenv("K_SERVICE") != "" || os.Getenv("CLOUD_RUN_JOB") != "" {
			format = "json"
		} else {
			format = "text"
		}
	}

	SetDefaultProjectID(cfg.ProjectID)

	opts := &slog.HandlerOptions{
		Level: level,
		ReplaceAttr: func(groups []string, a slog.Attr) slog.Attr {
			if len(groups) > 0 {
				return a
			}
			// Map standard slog keys to Google Cloud Logging structured JSON conventions
			if format == "json" {
				switch a.Key {
				case slog.LevelKey:
					lvl, ok := a.Value.Any().(slog.Level)
					if ok {
						return slog.String("severity", cloudSeverity(lvl))
					}
					return slog.String("severity", a.Value.String())
				case slog.MessageKey:
					return slog.Attr{Key: "message", Value: a.Value}
				case slog.TimeKey:
					if t, ok := a.Value.Any().(time.Time); ok {
						return slog.String("timestamp", t.UTC().Format(time.RFC3339Nano))
					}
				}
			}
			return a
		},
	}

	var baseHandler slog.Handler
	if format == "json" {
		baseHandler = slog.NewJSONHandler(out, opts)
	} else {
		baseHandler = slog.NewTextHandler(out, opts)
	}

	if cfg.Service != "" && format == "json" {
		baseHandler = baseHandler.WithAttrs([]slog.Attr{
			slog.Group("serviceContext",
				slog.String("service", cfg.Service),
				slog.String("version", cfg.Version),
			),
		})
	}

	logger := slog.New(&contextHandler{next: baseHandler})
	slog.SetDefault(logger)
	return logger
}

// SetDefaultProjectID sets the default GCP project ID used to format Cloud Trace resource paths.
func SetDefaultProjectID(projectID string) {
	globalMu.Lock()
	defer globalMu.Unlock()
	globalProjectID = strings.TrimSpace(projectID)
}

// DefaultProjectID returns the configured default GCP project ID.
func DefaultProjectID() string {
	globalMu.RLock()
	defer globalMu.RUnlock()
	return globalProjectID
}

// ParseLevel converts a level string ("debug", "info", "warn", "error") into slog.Level.
func ParseLevel(raw string) slog.Level {
	switch strings.ToLower(strings.TrimSpace(raw)) {
	case "debug":
		return slog.LevelDebug
	case "warn", "warning":
		return slog.LevelWarn
	case "error":
		return slog.LevelError
	default:
		return slog.LevelInfo
	}
}

func cloudSeverity(l slog.Level) string {
	switch {
	case l <= slog.LevelDebug:
		return "DEBUG"
	case l < slog.LevelWarn:
		return "INFO"
	case l < slog.LevelError:
		return "WARNING"
	default:
		return "ERROR"
	}
}

type contextHandler struct {
	next slog.Handler
}

func (h *contextHandler) Enabled(ctx context.Context, level slog.Level) bool {
	return h.next.Enabled(ctx, level)
}

func (h *contextHandler) Handle(ctx context.Context, r slog.Record) error {
	if ctx != nil {
		if info, ok := TraceFromContext(ctx); ok {
			if info.RequestID != "" {
				r.AddAttrs(slog.String("request_id", info.RequestID))
			}
			if info.TraceID != "" {
				r.AddAttrs(slog.String("trace_id", info.TraceID))
				proj := info.ProjectID
				if proj == "" {
					proj = DefaultProjectID()
				}
				if proj != "" {
					r.AddAttrs(slog.String("logging.googleapis.com/trace", fmt.Sprintf("projects/%s/traces/%s", proj, info.TraceID)))
				}
			}
			if info.SpanID != "" {
				r.AddAttrs(slog.String("span_id", info.SpanID))
				r.AddAttrs(slog.String("logging.googleapis.com/spanId", info.SpanID))
			}
			if info.ParentSpanID != "" {
				r.AddAttrs(slog.String("parent_span_id", info.ParentSpanID))
			}
			if info.Sampled {
				r.AddAttrs(slog.Bool("logging.googleapis.com/trace_sampled", true))
			}
			if info.EntryPoint != "" {
				r.AddAttrs(slog.String("entry_point", info.EntryPoint))
			}
		}
	}
	return h.next.Handle(ctx, r)
}

func (h *contextHandler) WithAttrs(attrs []slog.Attr) slog.Handler {
	return &contextHandler{next: h.next.WithAttrs(attrs)}
}

func (h *contextHandler) WithGroup(name string) slog.Handler {
	return &contextHandler{next: h.next.WithGroup(name)}
}

// WithTrace returns a new context carrying the provided TraceInfo.
func WithTrace(ctx context.Context, info TraceInfo) context.Context {
	if ctx == nil {
		ctx = context.Background()
	}
	return context.WithValue(ctx, traceCtxKey{}, info)
}

// TraceFromContext extracts TraceInfo from context if present.
func TraceFromContext(ctx context.Context) (TraceInfo, bool) {
	if ctx == nil {
		return TraceInfo{}, false
	}
	info, ok := ctx.Value(traceCtxKey{}).(TraceInfo)
	return info, ok
}

// Logger returns the default structured logger. Because the default handler is context-aware,
// callers using slog.InfoContext / Logger(ctx).InfoContext(ctx, ...) automatically include trace and request fields.
func Logger(_ context.Context) *slog.Logger {
	return slog.Default()
}

// ParseCloudTraceHeader parses Google Cloud's `X-Cloud-Trace-Context` header:
// Format: `TRACE_ID/SPAN_ID;o=TRACE_TRUE`
func ParseCloudTraceHeader(header string) (traceID string, spanID string, sampled bool) {
	header = strings.TrimSpace(header)
	if header == "" {
		return "", "", false
	}

	slashIdx := strings.IndexByte(header, '/')
	if slashIdx <= 0 {
		if isValidHex(header, 32) {
			return strings.ToLower(header), "", false
		}
		return "", "", false
	}

	rawTraceID := header[:slashIdx]
	if !isValidHex(rawTraceID, 32) {
		return "", "", false
	}
	traceID = strings.ToLower(rawTraceID)

	rest := header[slashIdx+1:]
	semiIdx := strings.IndexByte(rest, ';')
	rawSpan := rest
	rawOpts := ""
	if semiIdx >= 0 {
		rawSpan = rest[:semiIdx]
		rawOpts = rest[semiIdx+1:]
	}

	if rawSpan != "" {
		// In X-Cloud-Trace-Context, SPAN_ID is often a decimal uint64 or 16-char hex string.
		if u, err := strconv.ParseUint(rawSpan, 10, 64); err == nil && u > 0 {
			spanID = fmt.Sprintf("%016x", u)
		} else if isValidHex(rawSpan, 16) {
			spanID = strings.ToLower(rawSpan)
		}
	}

	if strings.Contains(rawOpts, "o=1") {
		sampled = true
	}

	return traceID, spanID, sampled
}

// ParseTraceparentHeader parses W3C `traceparent` header:
// Format: `00-<32-hex-trace-id>-<16-hex-span-id>-<2-hex-flags>`
func ParseTraceparentHeader(header string) (traceID string, spanID string, sampled bool) {
	parts := strings.Split(strings.TrimSpace(header), "-")
	if len(parts) != 4 {
		return "", "", false
	}
	if !isValidHex(parts[1], 32) || !isValidHex(parts[2], 16) {
		return "", "", false
	}
	traceID = strings.ToLower(parts[1])
	spanID = strings.ToLower(parts[2])
	if flags, err := strconv.ParseUint(parts[3], 16, 8); err == nil {
		sampled = (flags & 0x01) == 0x01
	}
	return traceID, spanID, sampled
}

// NewTraceID generates a random 16-byte (32 hex char) trace ID.
func NewTraceID() string {
	var b [16]byte
	if _, err := rand.Read(b[:]); err != nil {
		return fmt.Sprintf("%032x", time.Now().UnixNano())
	}
	return hex.EncodeToString(b[:])
}

// NewSpanID generates a random 8-byte (16 hex char) span ID.
func NewSpanID() string {
	var b [8]byte
	if _, err := rand.Read(b[:]); err != nil {
		return fmt.Sprintf("%016x", time.Now().UnixNano())
	}
	return hex.EncodeToString(b[:])
}

func isValidHex(s string, expectedLen int) bool {
	if len(s) != expectedLen {
		return false
	}
	allZero := true
	for i := range len(s) {
		c := s[i]
		isDigit := c >= '0' && c <= '9'
		isLowerHex := c >= 'a' && c <= 'f'
		isUpperHex := c >= 'A' && c <= 'F'
		if !isDigit && !isLowerHex && !isUpperHex {
			return false
		}
		if c != '0' {
			allZero = false
		}
	}
	return !allZero
}

// Span represents an active in-flight operation span for timing and structured trace logging.
type Span struct {
	ctx       context.Context
	name      string
	startTime time.Time
	attrs     []slog.Attr
	info      TraceInfo
}

// StartSpan creates a child span in ctx (initializing a root TraceInfo if none exists)
// and returns the updated context and Span handle. Callers must invoke `span.End(err)` when done.
func StartSpan(ctx context.Context, name string, attrs ...slog.Attr) (context.Context, *Span) {
	if ctx == nil {
		ctx = context.Background()
	}
	parent, ok := TraceFromContext(ctx)
	child := parent
	if !ok || child.TraceID == "" {
		child.TraceID = NewTraceID()
	}
	if child.RequestID == "" {
		child.RequestID = NewSpanID()
	}
	if child.ProjectID == "" {
		child.ProjectID = DefaultProjectID()
	}
	child.ParentSpanID = parent.SpanID
	child.SpanID = NewSpanID()

	spanCtx := WithTrace(ctx, child)
	return spanCtx, &Span{
		ctx:       spanCtx,
		name:      name,
		startTime: time.Now(),
		attrs:     attrs,
		info:      child,
	}
}

// AddAttrs appends structured attributes to the span before it completes.
func (s *Span) AddAttrs(attrs ...slog.Attr) {
	if s == nil {
		return
	}
	s.attrs = append(s.attrs, attrs...)
}

// End completes the span, computing elapsed duration and emitting a structured span log record.
// Failed spans log at ERROR level; slow spans (>2s) log at WARN level; normal spans log at DEBUG level.
func (s *Span) End(err error) {
	if s == nil {
		return
	}
	duration := time.Since(s.startTime)
	durationMs := float64(duration.Microseconds()) / 1000.0

	logAttrs := make([]slog.Attr, 0, len(s.attrs)+4)
	logAttrs = append(logAttrs,
		slog.String("event", "span_completed"),
		slog.String("span_name", s.name),
		slog.Float64("duration_ms", durationMs),
	)
	logAttrs = append(logAttrs, s.attrs...)

	if err != nil {
		logAttrs = append(logAttrs,
			slog.String("status", "error"),
			slog.String("error", err.Error()),
		)
		slog.LogAttrs(s.ctx, slog.LevelError, fmt.Sprintf("span %s failed: %v", s.name, err), logAttrs...)
		return
	}

	logAttrs = append(logAttrs, slog.String("status", "ok"))
	level := slog.LevelDebug
	if duration >= 2*time.Second {
		level = slog.LevelWarn
	}
	slog.LogAttrs(s.ctx, level, fmt.Sprintf("span %s completed", s.name), logAttrs...)
}
