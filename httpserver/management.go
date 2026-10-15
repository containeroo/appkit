package httpserver

import "net/http"

// ManagementRoutes contains the operational endpoints served outside application middleware.
type ManagementRoutes struct {
	// Health serves process liveness checks at /healthz.
	Health http.Handler
	// Readiness serves dependency readiness checks at /readyz.
	Readiness http.Handler
	// Metrics optionally serves metrics exposition at /metrics.
	Metrics http.Handler
}

// NewManagement constructs the standard operational route handler.
func NewManagement(routes ManagementRoutes) http.Handler {
	mux := http.NewServeMux()
	if routes.Health != nil {
		mux.Handle("GET /healthz", routes.Health)
	}
	if routes.Readiness != nil {
		mux.Handle("GET /readyz", routes.Readiness)
	}
	if routes.Metrics != nil {
		mux.Handle("GET /metrics", routes.Metrics)
	}
	return mux
}
