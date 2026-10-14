// Package httpserver composes application and management HTTP handlers and serves them independently.
package httpserver

import (
	"context"
	"errors"
	"log/slog"
	"net/http"

	"github.com/containeroo/httpkit/grace"
	"github.com/containeroo/httpkit/prefix"
	kitserver "github.com/containeroo/httpkit/server"
)

// Config supplies public and operational routes and the application deployment prefix.
type Config struct {
	// Application serves application and API traffic.
	Application http.Handler
	// Management serves operational traffic such as health, readiness, and metrics.
	Management http.Handler
	// RoutePrefix mounts public routes beneath the deployment prefix; management routes remain unprefixed.
	RoutePrefix string
}

// Handlers contains the assembled application and management handlers.
type Handlers struct {
	// Application is served on the public listener.
	Application http.Handler
	// Management is served on the operational listener when configured.
	Management http.Handler
}

// New constructs independent HTTP handlers; only application routes use the route prefix.
func New(config Config) Handlers {
	var handlers Handlers
	if config.Application != nil {
		handlers.Application = prefix.MountUnderPrefixWithOptions(
			config.Application, config.RoutePrefix, prefix.WithRedirectRewriting(),
		)
	}
	handlers.Management = config.Management
	return handlers
}

// Run serves application and optionally management traffic until shutdown or a listener fails.
func Run(
	ctx context.Context,
	applicationAddress string,
	managementAddress string,
	handlers Handlers,
	appLogger *slog.Logger,
	mgmtLogger *slog.Logger,
) error {
	if applicationAddress == "" {
		return errors.New("application listen address is required")
	}
	if handlers.Application == nil {
		return errors.New("application handler is required")
	}
	if appLogger == nil {
		return errors.New("application logger is required")
	}

	listeners := []kitserver.Listener{{
		Address: applicationAddress,
		Handler: handlers.Application,
		Logger:  appLogger,
		Options: []grace.Option{grace.WithMaxHeaderValueCount(100)},
	}}
	if managementAddress != "" {
		if handlers.Management == nil {
			return errors.New("management handler is required when management listener is configured")
		}
		if mgmtLogger == nil {
			return errors.New("management logger is required when management listener is configured")
		}
		listeners = append(listeners, kitserver.Listener{
			Address: managementAddress,
			Handler: handlers.Management,
			Logger:  mgmtLogger,
			Options: []grace.Option{grace.WithMaxHeaderValueCount(100)},
		})
	}
	return kitserver.Run(ctx, listeners...)
}
