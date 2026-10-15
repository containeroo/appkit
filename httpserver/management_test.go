package httpserver

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestNewManagementRegistersOperationalRoutes(t *testing.T) {
	t.Parallel()

	handler := NewManagement(ManagementRoutes{
		Health:    managementResponse("healthy"),
		Readiness: managementResponse("ready"),
		Metrics:   managementResponse("metric 1"),
	})

	t.Run("health", func(t *testing.T) {
		t.Parallel()
		assertManagementResponse(t, handler, http.MethodGet, "/healthz", http.StatusOK, "healthy")
	})
	t.Run("readiness", func(t *testing.T) {
		t.Parallel()
		assertManagementResponse(t, handler, http.MethodGet, "/readyz", http.StatusOK, "ready")
	})
	t.Run("metrics", func(t *testing.T) {
		t.Parallel()
		assertManagementResponse(t, handler, http.MethodGet, "/metrics", http.StatusOK, "metric 1")
	})
	t.Run("unsupported method", func(t *testing.T) {
		t.Parallel()
		assertManagementResponse(t, handler, http.MethodPost, "/healthz", http.StatusMethodNotAllowed, "Method Not Allowed\n")
	})
}

func TestNewManagementOmitsOptionalRoutes(t *testing.T) {
	t.Parallel()

	handler := NewManagement(ManagementRoutes{Health: managementResponse("healthy")})

	assertManagementResponse(t, handler, http.MethodGet, "/healthz", http.StatusOK, "healthy")
	assertManagementResponse(t, handler, http.MethodGet, "/readyz", http.StatusNotFound, "404 page not found\n")
	assertManagementResponse(t, handler, http.MethodGet, "/metrics", http.StatusNotFound, "404 page not found\n")
}

func managementResponse(body string) http.Handler {
	return http.HandlerFunc(func(response http.ResponseWriter, _ *http.Request) {
		_, _ = response.Write([]byte(body))
	})
}

func assertManagementResponse(t *testing.T, handler http.Handler, method, path string, status int, body string) {
	t.Helper()
	response := httptest.NewRecorder()
	handler.ServeHTTP(response, httptest.NewRequest(method, path, nil))
	if response.Code != status {
		t.Fatalf("status = %d, want %d", response.Code, status)
	}
	if response.Body.String() != body {
		t.Fatalf("body = %q, want %q", response.Body.String(), body)
	}
}
