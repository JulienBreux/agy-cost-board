package server_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"testing/fstest"
	"time"

	"github.com/julienbreux/agy-ge-board/internal/attribution"
	"github.com/julienbreux/agy-ge-board/internal/bigquery"
	"github.com/julienbreux/agy-ge-board/internal/domain"
	"github.com/julienbreux/agy-ge-board/internal/server"
)

func setupTestServer() http.Handler {
	provider := bigquery.NewDemoDataProvider()
	engine := attribution.NewEngine(provider, 5*time.Minute)

	mockFS := fstest.MapFS{
		"index.html": &fstest.MapFile{
			Data: []byte("<!DOCTYPE html><html><body><h1>AGY GE Board</h1></body></html>"),
		},
		"assets/app.js": &fstest.MapFile{
			Data: []byte("console.log('loaded');"),
		},
	}

	srv := server.NewServer(engine, mockFS)
	return srv.Router()
}

func TestAPIEndpoints(t *testing.T) {
	router := setupTestServer()

	t.Run("GET /healthz returns status 200 OK", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/healthz", nil)
		rec := httptest.NewRecorder()
		router.ServeHTTP(rec, req)

		if rec.Code != http.StatusOK {
			t.Fatalf("expected 200 OK, got %d", rec.Code)
		}
		var res map[string]string
		if err := json.NewDecoder(rec.Body).Decode(&res); err != nil {
			t.Fatalf("invalid json response: %v", err)
		}
		if res["status"] != "ok" {
			t.Errorf("expected status: ok, got %v", res["status"])
		}
	})

	t.Run("GET /api/v1/metrics/overview returns high-level metrics", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/api/v1/metrics/overview?days=30", nil)
		rec := httptest.NewRecorder()
		router.ServeHTTP(rec, req)

		if rec.Code != http.StatusOK {
			t.Fatalf("expected 200 OK, got %d", rec.Code)
		}
		var metrics domain.OverviewMetrics
		if err := json.NewDecoder(rec.Body).Decode(&metrics); err != nil {
			t.Fatalf("failed to decode overview metrics: %v", err)
		}
		if metrics.TotalBilledCost <= 0 {
			t.Errorf("expected positive total billed cost, got %f", metrics.TotalBilledCost)
		}
		if metrics.TotalTokens <= 0 {
			t.Errorf("expected positive total tokens, got %d", metrics.TotalTokens)
		}
		if len(metrics.DailyTrends) == 0 {
			t.Errorf("expected non-empty daily spend trend")
		}
	})

	t.Run("GET /api/v1/costs/users returns list of attributed user costs", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/api/v1/costs/users?days=14", nil)
		rec := httptest.NewRecorder()
		router.ServeHTTP(rec, req)

		if rec.Code != http.StatusOK {
			t.Fatalf("expected 200 OK, got %d", rec.Code)
		}
		var costs []domain.AllocatedUserCost
		if err := json.NewDecoder(rec.Body).Decode(&costs); err != nil {
			t.Fatalf("failed to decode costs: %v", err)
		}
		if len(costs) == 0 {
			t.Errorf("expected non-empty costs slice")
		}
		if costs[0].UserID == "" || costs[0].AllocatedCost < 0 {
			t.Errorf("invalid cost record: %+v", costs[0])
		}
	})

	t.Run("GET /api/v1/costs/users with model filter", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/api/v1/costs/users?days=14&model=gemini-1.5-pro", nil)
		rec := httptest.NewRecorder()
		router.ServeHTTP(rec, req)

		if rec.Code != http.StatusOK {
			t.Fatalf("expected 200 OK, got %d", rec.Code)
		}
		var costs []domain.AllocatedUserCost
		if err := json.NewDecoder(rec.Body).Decode(&costs); err != nil {
			t.Fatalf("failed to decode filtered costs: %v", err)
		}
		for _, c := range costs {
			if c.Model != "gemini-1.5-pro" {
				t.Errorf("expected model gemini-1.5-pro, got %s", c.Model)
			}
		}
	})

	t.Run("GET /api/v1/licenses/status returns license governance metrics", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/api/v1/licenses/status?days=30", nil)
		rec := httptest.NewRecorder()
		router.ServeHTTP(rec, req)

		if rec.Code != http.StatusOK {
			t.Fatalf("expected 200 OK, got %d", rec.Code)
		}
		var gov domain.LicenseGovernanceSummary
		if err := json.NewDecoder(rec.Body).Decode(&gov); err != nil {
			t.Fatalf("failed to decode governance: %v", err)
		}
		if gov.SeatQuota <= 0 {
			t.Errorf("expected positive seat quota, got %d", gov.SeatQuota)
		}
		if gov.DormantSeats < 0 {
			t.Errorf("unexpected dormant seats: %d", gov.DormantSeats)
		}
	})

	t.Run("GET /api/v1/users/{id} returns user breakdown summary", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/api/v1/users/alex.turner@example.com?days=30", nil)
		rec := httptest.NewRecorder()
		router.ServeHTTP(rec, req)

		if rec.Code != http.StatusOK {
			t.Fatalf("expected 200 OK, got %d", rec.Code)
		}
		var user domain.UserCostSummary
		if err := json.NewDecoder(rec.Body).Decode(&user); err != nil {
			t.Fatalf("failed to decode user summary: %v", err)
		}
		if user.UserID != "alex.turner@example.com" {
			t.Errorf("expected alex.turner@example.com, got %s", user.UserID)
		}
		if user.TotalCost <= 0 {
			t.Errorf("expected positive total cost, got %f", user.TotalCost)
		}
	})

	t.Run("GET /api/v1/users/unknown@example.com returns 404", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/api/v1/users/unknown.nobody@example.com?days=30", nil)
		rec := httptest.NewRecorder()
		router.ServeHTTP(rec, req)

		if rec.Code != http.StatusNotFound {
			t.Errorf("expected 404 Not Found, got %d", rec.Code)
		}
	})

	t.Run("Static asset serving and SPA fallback to index.html", func(t *testing.T) {
		// Exact static asset
		req1 := httptest.NewRequest(http.MethodGet, "/assets/app.js", nil)
		rec1 := httptest.NewRecorder()
		router.ServeHTTP(rec1, req1)
		if rec1.Code != http.StatusOK {
			t.Errorf("expected 200 for /assets/app.js, got %d", rec1.Code)
		}

		// Root path
		req2 := httptest.NewRequest(http.MethodGet, "/", nil)
		rec2 := httptest.NewRecorder()
		router.ServeHTTP(rec2, req2)
		if rec2.Code != http.StatusOK {
			t.Errorf("expected 200 for /, got %d", rec2.Code)
		}

		// SPA route (e.g. /licenses) falls back to index.html
		req3 := httptest.NewRequest(http.MethodGet, "/licenses", nil)
		rec3 := httptest.NewRecorder()
		router.ServeHTTP(rec3, req3)
		if rec3.Code != http.StatusOK {
			t.Errorf("expected 200 for SPA fallback /licenses, got %d", rec3.Code)
		}
	})
}
