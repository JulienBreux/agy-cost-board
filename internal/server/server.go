package server

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/fs"
	"log/slog"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"
	"github.com/julienbreux/agy-cost-board/internal/attribution"
	"github.com/julienbreux/agy-cost-board/internal/domain"
	"github.com/julienbreux/agy-cost-board/internal/mcp"
	"github.com/julienbreux/agy-cost-board/internal/observability"
	"github.com/julienbreux/agy-cost-board/internal/setup"
)

// VersionInfo contains build and release metadata.
type VersionInfo struct {
	Version   string `json:"version"`
	Commit    string `json:"commit,omitempty"`
	BuildDate string `json:"build_date,omitempty"`
}

// Server encapsulates the HTTP router, middleware, and dependency engine.
type Server struct {
	engine      *attribution.Engine
	staticFS    fs.FS
	router      *chi.Mux
	setupCfg    setup.Config
	setupRunner setup.Runner
	versionInfo VersionInfo
	mcpServer   *mcp.Server
	mcpEnabled  bool
}

// NewServer initializes the Chi router with middleware, API routes, and embedded SPA handler.
func NewServer(engine *attribution.Engine, staticFS fs.FS) *Server {
	return NewServerWithSetup(engine, staticFS, setup.Config{
		ProjectID:      "demo-project",
		TelemetryTable: "demo-project.antigravity_telemetry.inference_logs",
		BillingTable:   "demo-project.billing_export.gcp_billing_export_v1_000",
		Demo:           true,
	}, setup.NewRunner())
}

// NewServerWithSetup initializes the Chi router with customized setup configuration and diagnostics runner.
func NewServerWithSetup(engine *attribution.Engine, staticFS fs.FS, cfg setup.Config, runner setup.Runner) *Server {
	if runner == nil {
		runner = setup.NewRunner()
	}
	s := &Server{
		engine:      engine,
		staticFS:    staticFS,
		router:      chi.NewRouter(),
		setupCfg:    cfg,
		setupRunner: runner,
		mcpServer:   mcp.NewServer(engine, "dev"),
		mcpEnabled:  true,
	}

	s.setupRoutes()
	return s
}

// SetVersionInfo sets the application build version metadata.
func (s *Server) SetVersionInfo(v VersionInfo) {
	s.versionInfo = v
	if s.engine != nil {
		s.mcpServer = mcp.NewServer(s.engine, v.Version)
	}
}

// VersionInfo returns the application build version metadata.
func (s *Server) VersionInfo() VersionInfo {
	return s.versionInfo
}

// SetMCPEnabled enables or disables the /mcp endpoint at runtime.
func (s *Server) SetMCPEnabled(enabled bool) {
	s.mcpEnabled = enabled
}

// MCPEnabled reports whether the /mcp endpoint is enabled.
func (s *Server) MCPEnabled() bool {
	return s.mcpEnabled
}

// Router returns the underlying Chi router.
func (s *Server) Router() http.Handler {
	return s.router
}

func (s *Server) setupRoutes() {
	r := s.router

	// Standard middlewares
	r.Use(middleware.RequestID)
	r.Use(s.traceAndLogMiddleware)
	r.Use(middleware.Recoverer)

	// Cross-Origin Resource Sharing
	r.Use(cors.Handler(cors.Options{
		AllowedOrigins:   []string{"*"},
		AllowedMethods:   []string{"GET", "POST", "OPTIONS"},
		AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type", "X-Request-Id", "X-Cloud-Trace-Context", "traceparent"},
		ExposedHeaders:   []string{"Link", "X-Request-Id", "X-Trace-Id"},
		AllowCredentials: true,
		MaxAge:           300,
	}))

	// Health check for Cloud Run and Kubernetes probes
	r.Get("/healthz", func(w http.ResponseWriter, r *http.Request) {
		res := map[string]string{
			"status":  "ok",
			"service": "agy-cost-board",
		}
		if s.versionInfo.Version != "" {
			res["version"] = s.versionInfo.Version
		}
		respondJSON(w, http.StatusOK, res)
	})

	// API v1 routes
	r.Route("/api/v1", func(api chi.Router) {
		api.Get("/version", func(w http.ResponseWriter, r *http.Request) {
			respondJSON(w, http.StatusOK, s.versionInfo)
		})
		api.Get("/mcp/status", s.handleMCPStatus)
		api.Get("/me", s.handleCurrentUser)
		api.Get("/metrics/overview", s.handleOverviewMetrics)
		api.Get("/leaderboard", s.handleLeaderboard)
		api.Get("/costs/users", s.handleAttributedCosts)
		api.Get("/licenses/status", s.handleLicenseGovernance)
		api.Get("/users/{id}", s.handleUserSummary)
		api.Get("/users/{id}/activity", s.handleUserActivity)
		api.Get("/users/{id}/dashboard", s.handleUserDashboard)
		api.Get("/setup/status", s.handleSetupStatus)
	})

	// Model Context Protocol (MCP) Streamable HTTP endpoint
	r.Post("/mcp", s.handleMCP)
	r.Get("/mcp", s.handleMCP)

	// Antigravity CLI statusline dynamic script
	r.Get("/statusline.sh", s.handleStatuslineScript)

	// Embedded SPA static file serving
	if s.staticFS != nil {
		s.setupStaticSPA(r)
	}
}

func (s *Server) traceAndLogMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()

		traceID, parentSpanID, sampled := observability.ParseCloudTraceHeader(r.Header.Get("X-Cloud-Trace-Context"))
		if traceID == "" {
			traceID, parentSpanID, sampled = observability.ParseTraceparentHeader(r.Header.Get("traceparent"))
		}
		if traceID == "" {
			traceID = observability.NewTraceID()
		}
		spanID := observability.NewSpanID()

		reqID := strings.TrimSpace(r.Header.Get("X-Request-Id"))
		if reqID == "" {
			reqID = middleware.GetReqID(r.Context())
		}
		if reqID == "" {
			reqID = spanID
		}

		entryPoint := "http"
		if strings.HasPrefix(r.URL.Path, "/mcp") {
			entryPoint = "mcp-http"
		}

		info := observability.TraceInfo{
			TraceID:      traceID,
			SpanID:       spanID,
			ParentSpanID: parentSpanID,
			Sampled:      sampled,
			RequestID:    reqID,
			EntryPoint:   entryPoint,
			ProjectID:    s.setupCfg.ProjectID,
		}

		w.Header().Set("X-Request-Id", reqID)
		w.Header().Set("X-Trace-Id", traceID)

		ctx := observability.WithTrace(r.Context(), info)
		ww := middleware.NewWrapResponseWriter(w, r.ProtoMajor)
		next.ServeHTTP(ww, r.WithContext(ctx))

		status := ww.Status()
		if status == 0 {
			status = http.StatusOK
		}
		duration := time.Since(start)
		durationMs := float64(duration.Microseconds()) / 1000.0

		level := slog.LevelInfo
		switch {
		case status >= 500:
			level = slog.LevelError
		case status >= 400:
			level = slog.LevelWarn
		case r.URL.Path == "/healthz" || (!strings.HasPrefix(r.URL.Path, "/api/") && !strings.HasPrefix(r.URL.Path, "/mcp")):
			level = slog.LevelDebug
		}

		slog.LogAttrs(ctx, level, fmt.Sprintf("%s %s %d", r.Method, r.URL.Path, status),
			slog.String("event", "http_request"),
			slog.String("method", r.Method),
			slog.String("path", r.URL.Path),
			slog.String("query", r.URL.RawQuery),
			slog.Int("status", status),
			slog.Int("bytes", ww.BytesWritten()),
			slog.Float64("duration_ms", durationMs),
			slog.Group("httpRequest",
				slog.String("requestMethod", r.Method),
				slog.String("requestUrl", r.URL.RequestURI()),
				slog.Int("status", status),
				slog.Int("responseSize", ww.BytesWritten()),
				slog.String("userAgent", r.UserAgent()),
				slog.String("remoteIp", r.RemoteAddr),
				slog.String("latency", fmt.Sprintf("%.6fs", duration.Seconds())),
			),
		)
	})
}

func (s *Server) handleMCP(w http.ResponseWriter, r *http.Request) {
	if !s.mcpEnabled || s.mcpServer == nil {
		http.NotFound(w, r)
		return
	}
	s.mcpServer.ServeHTTP(w, r)
}

func (s *Server) handleMCPStatus(w http.ResponseWriter, r *http.Request) {
	var tools []mcp.Tool
	if s.mcpServer != nil {
		tools = s.mcpServer.Tools()
	}
	respondJSON(w, http.StatusOK, map[string]any{
		"mcp_enabled":      s.mcpEnabled,
		"protocol_version": mcp.ProtocolVersion,
		"endpoint":         "/mcp",
		"transports":       []string{"http", "stdio"},
		"tools":            tools,
	})
}

func (s *Server) scopedEngine(r *http.Request) *attribution.Engine {
	if s.engine == nil {
		return nil
	}
	return s.engine.FilterByProject(r.URL.Query().Get("project"))
}

func (s *Server) handleOverviewMetrics(w http.ResponseWriter, r *http.Request) {
	days := parseIntQuery(r, "days", 30)
	engine := s.scopedEngine(r)
	overview, err := engine.GetOverviewMetrics(r.Context(), days)
	if err != nil {
		respondError(w, r, http.StatusInternalServerError, err.Error())
		return
	}
	poolMode := r.URL.Query().Get("pool_mode")
	seatBasis := r.URL.Query().Get("seat_basis")
	if (poolMode != "" || seatBasis != "") && overview != nil {
		ovCopy := *overview
		ovCopy.WeeklyPool = attribution.ApplyPoolWhatIf(overview.WeeklyPool, poolMode, seatBasis)
		ovCopy.Forecast = attribution.RecomputeForecastForPool(overview.Forecast, ovCopy.WeeklyPool)
		overview = &ovCopy
	}
	respondJSON(w, http.StatusOK, overview)
}

func (s *Server) handleLeaderboard(w http.ResponseWriter, r *http.Request) {
	days := parseIntQuery(r, "days", 30)
	engine := s.scopedEngine(r)
	entries, err := engine.GetLeaderboard(r.Context(), days)
	if err != nil {
		respondError(w, r, http.StatusInternalServerError, err.Error())
		return
	}
	if paygMode := r.URL.Query().Get("payg_mode"); paygMode != "" {
		entries = attribution.ApplyPAYGModeToLeaderboard(entries, paygMode)
	}
	respondJSON(w, http.StatusOK, entries)
}

func (s *Server) handleAttributedCosts(w http.ResponseWriter, r *http.Request) {
	days := parseIntQuery(r, "days", 30)
	modelFilter := r.URL.Query().Get("model")
	engine := s.scopedEngine(r)

	costs, err := engine.GetAttributedCosts(r.Context(), days, modelFilter)
	if err != nil {
		respondError(w, r, http.StatusInternalServerError, err.Error())
		return
	}
	if paygMode := r.URL.Query().Get("payg_mode"); paygMode != "" {
		costs = attribution.ApplyPAYGModeToCosts(costs, paygMode)
	}
	respondJSON(w, http.StatusOK, costs)
}

func (s *Server) handleLicenseGovernance(w http.ResponseWriter, r *http.Request) {
	days := parseIntQuery(r, "days", 30)
	gov, err := s.scopedEngine(r).GetLicenseGovernance(r.Context(), days)
	if err != nil {
		respondError(w, r, http.StatusInternalServerError, err.Error())
		return
	}
	poolMode := r.URL.Query().Get("pool_mode")
	seatBasis := r.URL.Query().Get("seat_basis")
	if (poolMode != "" || seatBasis != "") && gov != nil {
		govCopy := *gov
		govCopy.WeeklyPool = attribution.ApplyPoolWhatIf(gov.WeeklyPool, poolMode, seatBasis)
		gov = &govCopy
	}
	respondJSON(w, http.StatusOK, gov)
}

func extractUserID(r *http.Request) string {
	raw := chi.URLParam(r, "id")
	if unescaped, err := url.PathUnescape(raw); err == nil && unescaped != "" {
		return unescaped
	}
	return raw
}

func (s *Server) handleUserSummary(w http.ResponseWriter, r *http.Request) {
	userID := extractUserID(r)
	if userID == "" {
		respondError(w, r, http.StatusBadRequest, "missing user id parameter")
		return
	}
	days := parseIntQuery(r, "days", 30)

	summary, err := s.scopedEngine(r).GetUserSummary(r.Context(), userID, days)
	if err != nil {
		if errors.Is(err, domain.ErrUserNotFound) {
			respondError(w, r, http.StatusNotFound, "user not found in telemetry window")
			return
		}
		respondError(w, r, http.StatusInternalServerError, err.Error())
		return
	}
	respondJSON(w, http.StatusOK, summary)
}

func (s *Server) handleCurrentUser(w http.ResponseWriter, r *http.Request) {
	userHeader := r.Header.Get("X-Goog-Authenticated-User-Email")
	if userHeader == "" {
		userHeader = r.Header.Get("X-Forwarded-Email")
	}

	identity := domain.CurrentUserIdentity{
		Authenticated: false,
		Source:        "none",
	}

	if userHeader != "" {
		cleanUser := strings.TrimPrefix(userHeader, "accounts.google.com:")
		identity.UserID = cleanUser
		identity.Email = cleanUser
		identity.Authenticated = true
		identity.Source = "iap"
		respondJSON(w, http.StatusOK, identity)
		return
	}

	if s.setupCfg.Demo {
		identity.UserID = "alex.turner@google.com"
		identity.Email = "alex.turner@google.com"
		identity.Authenticated = false
		identity.Source = "demo"
		respondJSON(w, http.StatusOK, identity)
		return
	}

	respondJSON(w, http.StatusOK, identity)
}

func (s *Server) handleUserActivity(w http.ResponseWriter, r *http.Request) {
	userID := extractUserID(r)
	if userID == "" {
		respondError(w, r, http.StatusBadRequest, "missing user id parameter")
		return
	}
	days := parseIntQuery(r, "days", 30)
	limit := parseIntQuery(r, "limit", 50)

	activity, err := s.scopedEngine(r).GetUserActivity(r.Context(), userID, days, limit)
	if err != nil {
		respondError(w, r, http.StatusInternalServerError, err.Error())
		return
	}
	respondJSON(w, http.StatusOK, activity)
}

func (s *Server) handleUserDashboard(w http.ResponseWriter, r *http.Request) {
	userID := extractUserID(r)
	if userID == "" {
		respondError(w, r, http.StatusBadRequest, "missing user id parameter")
		return
	}
	days := parseIntQuery(r, "days", 30)
	budget := parseFloatQuery(r, "budget", 0)
	if budget <= 0 {
		budget = parseFloatQuery(r, "monthlyBudget", 150.0)
	}

	driving, err := s.scopedEngine(r).GetUserConsumptionDriving(r.Context(), userID, days, budget)
	if err != nil {
		if errors.Is(err, domain.ErrUserNotFound) {
			respondError(w, r, http.StatusNotFound, "user not found in telemetry window")
			return
		}
		respondError(w, r, http.StatusInternalServerError, err.Error())
		return
	}
	respondJSON(w, http.StatusOK, driving)
}

func (s *Server) handleSetupStatus(w http.ResponseWriter, r *http.Request) {
	runner := s.setupRunner
	if runner == nil {
		runner = setup.NewRunner()
	}
	report := runner.RunAll(r.Context(), s.setupCfg)
	respondJSON(w, http.StatusOK, report)
}

func (s *Server) setupStaticSPA(r *chi.Mux) {
	fileServer := http.FileServer(http.FS(s.staticFS))

	r.Get("/*", func(w http.ResponseWriter, r *http.Request) {
		path := strings.TrimPrefix(r.URL.Path, "/")
		if path == "" {
			path = "index.html"
		}

		// Try opening requested file
		f, err := s.staticFS.Open(path)
		if err == nil {
			_ = f.Close()
			fileServer.ServeHTTP(w, r)
			return
		}

		// Fallback to index.html for React SPA client-side routing
		indexFile, err := s.staticFS.Open("index.html")
		if err == nil {
			_ = indexFile.Close()
			r.URL.Path = "/"
			fileServer.ServeHTTP(w, r)
			return
		}

		http.NotFound(w, r)
	})
}

func parseIntQuery(r *http.Request, key string, defaultValue int) int {
	valStr := r.URL.Query().Get(key)
	if valStr == "" {
		return defaultValue
	}
	val, err := strconv.Atoi(valStr)
	if err != nil || val <= 0 {
		return defaultValue
	}
	return val
}

func parseFloatQuery(r *http.Request, key string, defaultValue float64) float64 {
	valStr := r.URL.Query().Get(key)
	if valStr == "" {
		return defaultValue
	}
	val, err := strconv.ParseFloat(valStr, 64)
	if err != nil || val <= 0 {
		return defaultValue
	}
	return val
}

func respondJSON(w http.ResponseWriter, status int, data any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(data)
}

func respondError(w http.ResponseWriter, r *http.Request, status int, message string) {
	res := map[string]string{
		"error": message,
	}
	if r != nil {
		ctx := r.Context()
		if info, ok := observability.TraceFromContext(ctx); ok {
			if info.RequestID != "" {
				res["request_id"] = info.RequestID
			}
			if info.TraceID != "" {
				res["trace_id"] = info.TraceID
			}
		}
		level := slog.LevelWarn
		if status >= 500 {
			level = slog.LevelError
		}
		slog.LogAttrs(ctx, level, "HTTP handler error",
			slog.String("event", "http_handler_error"),
			slog.String("method", r.Method),
			slog.String("path", r.URL.Path),
			slog.Int("status", status),
			slog.String("error", message),
		)
	}
	respondJSON(w, status, res)
}
