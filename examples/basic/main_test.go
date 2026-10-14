package main

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/containeroo/appkit/httpserver"
)

func TestExampleHandlers(t *testing.T) {
	cfg, err := parse([]string{"--route-prefix", "/example", "--listen-address", "127.0.0.1:9100", "--management-listen-address", "127.0.0.1:9101"})
	if err != nil {
		t.Fatal(err)
	}
	if cfg.Server.ListenAddress != "127.0.0.1:9100" || cfg.Server.ManagementListenAddress != "127.0.0.1:9101" {
		t.Fatalf("unexpected addresses: %#v", cfg.Server)
	}
	handlers := httpserver.New(newServerConfig(cfg))
	for _, tc := range []struct {
		name    string
		handler http.Handler
		path    string
		status  int
	}{
		{name: "application", handler: handlers.Application, path: "/example/", status: http.StatusOK},
		{name: "liveness", handler: handlers.Management, path: "/healthz", status: http.StatusNoContent},
		{name: "readiness", handler: handlers.Management, path: "/readyz", status: http.StatusNoContent},
	} {
		t.Run(tc.name, func(t *testing.T) {
			response := httptest.NewRecorder()
			tc.handler.ServeHTTP(response, httptest.NewRequest(http.MethodGet, tc.path, nil))
			if response.Code != tc.status {
				t.Fatalf("status = %d, want %d", response.Code, tc.status)
			}
		})
	}
}
