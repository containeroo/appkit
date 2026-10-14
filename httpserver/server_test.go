package httpserver

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestNewSeparatesApplicationPrefixAndManagement(t *testing.T) {
	application := http.NewServeMux()
	application.HandleFunc("GET /", func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write([]byte("application"))
	})
	management := http.NewServeMux()
	management.HandleFunc("GET /healthz", func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write([]byte("management"))
	})

	handlers := New(Config{
		Application: application,
		Management:  management,
		RoutePrefix: "/example",
	})

	cases := []struct {
		name    string
		handler http.Handler
		path    string
		status  int
		body    string
	}{
		{name: "application", handler: handlers.Application, path: "/example/", status: http.StatusOK, body: "application"},
		{name: "management", handler: handlers.Management, path: "/healthz", status: http.StatusOK, body: "management"},
		{name: "prefixed management", handler: handlers.Management, path: "/example/healthz", status: http.StatusNotFound},
	}
	for _, test := range cases {
		t.Run(test.name, func(t *testing.T) {
			response := httptest.NewRecorder()
			test.handler.ServeHTTP(response, httptest.NewRequest(http.MethodGet, test.path, nil))
			if response.Code != test.status {
				t.Fatalf("status = %d, want %d", response.Code, test.status)
			}
			if test.body != "" && response.Body.String() != test.body {
				t.Fatalf("body = %q, want %q", response.Body.String(), test.body)
			}
		})
	}
}

func TestNewSupportsOptionalManagement(t *testing.T) {
	handlers := New(Config{Application: http.NewServeMux()})
	if handlers.Application == nil || handlers.Management != nil {
		t.Fatalf("unexpected handlers: %#v", handlers)
	}
}

func TestRunRejectsInvalidListeners(t *testing.T) {
	app := http.HandlerFunc(func(http.ResponseWriter, *http.Request) {})
	for _, test := range []struct {
		name       string
		appAddress string
		mgmtAddr   string
		handlers   Handlers
	}{
		{name: "missing address", handlers: Handlers{Application: app}},
		{name: "missing application", appAddress: "127.0.0.1:8080"},
		{name: "missing management handler", appAddress: "127.0.0.1:8080", mgmtAddr: "127.0.0.1:8081", handlers: Handlers{Application: app}},
	} {
		t.Run(test.name, func(t *testing.T) {
			if err := Run(context.Background(), test.appAddress, test.mgmtAddr, test.handlers, nil); err == nil {
				t.Fatal("expected invalid listener configuration to fail")
			}
		})
	}
}
