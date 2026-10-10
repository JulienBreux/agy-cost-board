package server_test

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestStatuslineEndpoint(t *testing.T) {
	router := setupTestServer(t)

	t.Run("GET /statusline.sh returns 200 OK and text/x-shellscript content type", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/statusline.sh", nil)
		rec := httptest.NewRecorder()
		router.ServeHTTP(rec, req)

		if rec.Code != http.StatusOK {
			t.Fatalf("expected 200 OK, got %d", rec.Code)
		}

		contentType := rec.Header().Get("Content-Type")
		if !strings.HasPrefix(contentType, "text/x-shellscript") {
			t.Errorf("expected Content-Type text/x-shellscript, got %s", contentType)
		}

		body := rec.Body.String()
		if !strings.HasPrefix(body, "#!/bin/bash") {
			t.Errorf("expected script to start with #!/bin/bash, got: %s", body[:min(30, len(body))])
		}
	})

	t.Run("GET /statusline.sh substitutes server base URL from request host and headers", func(t *testing.T) {
		// Case 1: Standard Host
		req1 := httptest.NewRequest(http.MethodGet, "/statusline.sh", nil)
		req1.Host = "localhost:8080"
		rec1 := httptest.NewRecorder()
		router.ServeHTTP(rec1, req1)

		if rec1.Code != http.StatusOK {
			t.Fatalf("expected 200 OK, got %d", rec1.Code)
		}
		body1 := rec1.Body.String()
		expectedURL1 := `DEFAULT_BOARD_URL="http://localhost:8080"`
		if !strings.Contains(body1, expectedURL1) {
			t.Errorf("expected script to contain %q, but was not found", expectedURL1)
		}

		// Case 2: X-Forwarded-Proto and X-Forwarded-Host
		req2 := httptest.NewRequest(http.MethodGet, "/statusline.sh", nil)
		req2.Host = "internal-ip"
		req2.Header.Set("X-Forwarded-Proto", "https")
		req2.Header.Set("X-Forwarded-Host", "cost.example.com")
		rec2 := httptest.NewRecorder()
		router.ServeHTTP(rec2, req2)

		if rec2.Code != http.StatusOK {
			t.Fatalf("expected 200 OK, got %d", rec2.Code)
		}
		body2 := rec2.Body.String()
		expectedURL2 := `DEFAULT_BOARD_URL="https://cost.example.com"`
		if !strings.Contains(body2, expectedURL2) {
			t.Errorf("expected script to contain %q, but was not found", expectedURL2)
		}
	})

	t.Run("GET /statusline.sh substitutes query parameters ?user=, ?days=, and ?ttl=", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/statusline.sh?user=alice@example.com&days=60&ttl=120", nil)
		rec := httptest.NewRecorder()
		router.ServeHTTP(rec, req)

		if rec.Code != http.StatusOK {
			t.Fatalf("expected 200 OK, got %d", rec.Code)
		}

		body := rec.Body.String()
		expectedUser := `DEFAULT_USER="alice@example.com"`
		if !strings.Contains(body, expectedUser) {
			t.Errorf("expected script to contain %q, but was not found", expectedUser)
		}

		expectedDays := `DEFAULT_DAYS="60"`
		if !strings.Contains(body, expectedDays) {
			t.Errorf("expected script to contain %q, but was not found", expectedDays)
		}

		expectedTTL := `DEFAULT_TTL="120"`
		if !strings.Contains(body, expectedTTL) {
			t.Errorf("expected script to contain %q, but was not found", expectedTTL)
		}
	})

	t.Run("GET /statusline.sh uses sensible defaults when query params are omitted", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/statusline.sh", nil)
		rec := httptest.NewRecorder()
		router.ServeHTTP(rec, req)

		if rec.Code != http.StatusOK {
			t.Fatalf("expected 200 OK, got %d", rec.Code)
		}

		body := rec.Body.String()
		expectedUser := `DEFAULT_USER=""`
		if !strings.Contains(body, expectedUser) {
			t.Errorf("expected script to contain %q, but was not found", expectedUser)
		}

		expectedDays := `DEFAULT_DAYS="30"`
		if !strings.Contains(body, expectedDays) {
			t.Errorf("expected script to contain %q, but was not found", expectedDays)
		}

		expectedTTL := `DEFAULT_TTL="300"`
		if !strings.Contains(body, expectedTTL) {
			t.Errorf("expected script to contain %q, but was not found", expectedTTL)
		}
	})
}
