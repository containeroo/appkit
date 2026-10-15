// Command basic runs an HTTP application with a separate operational listener.
package main

import (
	"context"
	"fmt"
	"net/http"
	"os"

	"github.com/containeroo/appkit/config"
	"github.com/containeroo/appkit/httpserver"
	"github.com/containeroo/appkit/logging"
	server "github.com/containeroo/httpkit/grace"
)

// applicationConfig holds the reusable deployment configuration for this example.
type applicationConfig struct {
	// Server contains application and management listen addresses and the route prefix.
	Server config.Server
	// Logging contains the logging format and diagnostic settings.
	Logging config.Logging
}

// parse registers the shared flags and resolves the final application configuration.
func parse(args []string) (applicationConfig, error) {
	var cfg applicationConfig

	flags := config.NewFlagSet("basic", "BASIC", "dev")
	flags.SectionOrder(config.SectionServer, config.SectionLogging)

	finishServer := config.RegisterServer(flags, &cfg.Server, config.ServerOptions{})
	finishLogging := config.RegisterLogging(flags, &cfg.Logging)

	if err := flags.Parse(args); err != nil {
		return applicationConfig{}, err
	}

	finishServer()
	finishLogging()

	return cfg, nil
}

// newServerConfig wires the application and operational routes.
func newServerConfig(cfg applicationConfig) httpserver.Config {
	application := http.NewServeMux()
	application.HandleFunc("GET /", func(w http.ResponseWriter, _ *http.Request) {
		_, _ = fmt.Fprintln(w, "Hello from appkit!")
	})

	management := httpserver.NewManagement(httpserver.ManagementRoutes{
		Health: http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
			w.WriteHeader(http.StatusNoContent)
		}),
		Readiness: http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
			w.WriteHeader(http.StatusNoContent)
		}),
	})

	return httpserver.Config{
		Application: application,
		Management:  management,
		RoutePrefix: cfg.Server.RoutePrefix,
	}
}

// run starts both listeners and returns when either exits or the context is canceled.
func run(ctx context.Context, args []string) error {
	cfg, err := parse(args)
	if err != nil {
		return err
	}

	logger := logging.Setup(cfg.Logging.Format, cfg.Logging.Debug, os.Stdout)
	serverConfig := newServerConfig(cfg)

	// Bind the process lifetime to operating-system shutdown signals.
	ctx, stop := server.SignalContext(ctx)
	defer stop()

	handlers := httpserver.New(serverConfig)
	if err := httpserver.Run(
		ctx,
		cfg.Server.ListenAddress,
		cfg.Server.ManagementListenAddress,
		handlers,
		logger.With("server", "application"),
		logger.With("server", "management"),
	); err != nil {
		logger.Error("run server", "event", "server_run_failed", "error", err)
		return err
	}

	return nil
}

// main runs the server until interrupted and reports startup or runtime failures.
func main() {
	if err := run(context.Background(), os.Args[1:]); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
