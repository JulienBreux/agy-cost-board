package server

import (
	"encoding/json"
	"errors"
	"io/fs"
	"net/http"
	"net/url"
	"strconv"
	"strings"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"
	"github.com/julienbreux/agy-cost-board/internal/attribution"
	"github.com/julienbreux/agy-cost-board/internal/domain"
	"github.com/julienbreux/agy-cost-board/internal/setup"
)

// Server encapsulates the HTTP router, middleware, and dependency engine.
type Server struct {
	engine      *attribution.Engine
	staticFS    fs.FS
	router      *chi.Mux
	setupCfg    setup.Config
	setupRunner setup.Runner
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
	}

	s.setupRoutes()
	return s
}

// Router returns the underlying Chi router.
func (s *Server) Router() http.Handler {
	return s.router
}

func (s *Server) setupRoutes() {
	r := s.router

	// Standard middlewares
	r.Use(middleware.RequestID)
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)

	// Cross-Origin Resource Sharing
	r.Use(cors.Handler(cors.Options{
		AllowedOrigins:   []string{"*"},
		AllowedMethods:   []string{"GET", "POST", "OPTIONS"},
		AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type"},
		ExposedHeaders:   []string{"Link"},
		AllowCredentials: true,
		MaxAge:           300,
	}))

	// Health check for Cloud Run and Kubernetes probes
	r.Get("/healthz", func(w http.ResponseWriter, r *http.Request) {
		respondJSON(w, http.StatusOK, map[string]string{
			"status":  "ok",
			"service": "agy-cost-board",
		})
	})

	// API v1 routes
	r.Route("/api/v1", func(api chi.Router) {
		api.Get("/me", s.handleCurrentUser)
		api.Get("/metrics/overview", s.handleOverviewMetrics)
		api.Get("/costs/users", s.handleAttributedCosts)
		api.Get("/licenses/status", s.handleLicenseGovernance)
		api.Get("/users/{id}", s.handleUserSummary)
		api.Get("/users/{id}/activity", s.handleUserActivity)
		api.Get("/users/{id}/dashboard", s.handleUserDashboard)
		api.Get("/setup/status", s.handleSetupStatus)
	})

	// Antigravity CLI statusline dynamic script
	r.Get("/statusline.sh", s.handleStatuslineScript)

	// Embedded SPA static file serving
	if s.staticFS != nil {
		s.setupStaticSPA(r)
	}
}

func (s *Server) handleOverviewMetrics(w http.ResponseWriter, r *http.Request) {
	days := parseIntQuery(r, "days", 30)
	overview, err := s.engine.GetOverviewMetrics(r.Context(), days)
	if err != nil {
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}
	respondJSON(w, http.StatusOK, overview)
}

func (s *Server) handleAttributedCosts(w http.ResponseWriter, r *http.Request) {
	days := parseIntQuery(r, "days", 30)
	modelFilter := r.URL.Query().Get("model")

	costs, err := s.engine.GetAttributedCosts(r.Context(), days, modelFilter)
	if err != nil {
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}
	respondJSON(w, http.StatusOK, costs)
}

func (s *Server) handleLicenseGovernance(w http.ResponseWriter, r *http.Request) {
	days := parseIntQuery(r, "days", 30)
	gov, err := s.engine.GetLicenseGovernance(r.Context(), days)
	if err != nil {
		respondError(w, http.StatusInternalServerError, err.Error())
		return
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
		respondError(w, http.StatusBadRequest, "missing user id parameter")
		return
	}
	days := parseIntQuery(r, "days", 30)

	summary, err := s.engine.GetUserSummary(r.Context(), userID, days)
	if err != nil {
		if errors.Is(err, domain.ErrUserNotFound) {
			respondError(w, http.StatusNotFound, "user not found in telemetry window")
			return
		}
		respondError(w, http.StatusInternalServerError, err.Error())
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
		respondError(w, http.StatusBadRequest, "missing user id parameter")
		return
	}
	days := parseIntQuery(r, "days", 30)
	limit := parseIntQuery(r, "limit", 50)

	activity, err := s.engine.GetUserActivity(r.Context(), userID, days, limit)
	if err != nil {
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}
	respondJSON(w, http.StatusOK, activity)
}

func (s *Server) handleUserDashboard(w http.ResponseWriter, r *http.Request) {
	userID := extractUserID(r)
	if userID == "" {
		respondError(w, http.StatusBadRequest, "missing user id parameter")
		return
	}
	days := parseIntQuery(r, "days", 30)
	budget := parseFloatQuery(r, "budget", 0)
	if budget <= 0 {
		budget = parseFloatQuery(r, "monthlyBudget", 150.0)
	}

	driving, err := s.engine.GetUserConsumptionDriving(r.Context(), userID, days, budget)
	if err != nil {
		if errors.Is(err, domain.ErrUserNotFound) {
			respondError(w, http.StatusNotFound, "user not found in telemetry window")
			return
		}
		respondError(w, http.StatusInternalServerError, err.Error())
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

func respondError(w http.ResponseWriter, status int, message string) {
	respondJSON(w, status, map[string]string{
		"error": message,
	})
}
