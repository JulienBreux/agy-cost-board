package server

import (
	"encoding/json"
	"errors"
	"io/fs"
	"net/http"
	"strconv"
	"strings"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"
	"github.com/julienbreux/agy-ge-board/internal/attribution"
	"github.com/julienbreux/agy-ge-board/internal/domain"
)

// Server encapsulates the HTTP router, middleware, and dependency engine.
type Server struct {
	engine   *attribution.Engine
	staticFS fs.FS
	router   *chi.Mux
}

// NewServer initializes the Chi router with middleware, API routes, and embedded SPA handler.
func NewServer(engine *attribution.Engine, staticFS fs.FS) *Server {
	s := &Server{
		engine:   engine,
		staticFS: staticFS,
		router:   chi.NewRouter(),
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
	r.Use(middleware.RealIP)
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
			"service": "agy-ge-board",
		})
	})

	// API v1 routes
	r.Route("/api/v1", func(api chi.Router) {
		api.Get("/metrics/overview", s.handleOverviewMetrics)
		api.Get("/costs/users", s.handleAttributedCosts)
		api.Get("/licenses/status", s.handleLicenseGovernance)
		api.Get("/users/{id}", s.handleUserSummary)
	})

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

func (s *Server) handleUserSummary(w http.ResponseWriter, r *http.Request) {
	userID := chi.URLParam(r, "id")
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
