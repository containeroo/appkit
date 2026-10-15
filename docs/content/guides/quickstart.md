# Quickstart

Appkit is designed to fit around an application rather than dictate its architecture. The application provides its own handlers, user model, persistence, and lifecycle; appkit supplies reusable deployment and HTTP primitives.

## 1. Register deployment flags

Create a `tinyflags` set and register only the groups your application uses. Each registration function fills a small struct and, where necessary, returns a finalizer that you call after parsing.

```go
flags := config.NewFlagSet("example", "EXAMPLE", version)

var server config.Server
finishServer := config.RegisterServer(flags, &server, config.ServerOptions{})

var loggingConfig config.Logging
finishLogging := config.RegisterLogging(flags, &loggingConfig)

if err := flags.Parse(os.Args[1:]); err != nil {
	return err
}
finishServer()
finishLogging()
```

This pattern preserves whether an optional CLI or environment override was omitted or explicitly set. Your application can then combine deployment overrides with its own persisted settings.

## 2. Build handlers

Keep public routes and operational routes separate. The management listener is the right place for health and readiness endpoints because it can be firewalled independently from application traffic.

```go
app := http.NewServeMux()
app.HandleFunc("GET /", home)

management := http.NewServeMux()
management.HandleFunc("GET /healthz", health)
management.HandleFunc("GET /readyz", readiness)
```

## 3. Run the server

Use the resolved server configuration and a structured logger. `httpserver.Run` blocks until context cancellation or a listener failure.

```go
logger := logging.Setup(loggingConfig.Format, loggingConfig.Debug, os.Stderr)
handlers := httpserver.New(httpserver.Config{
	Application: app,
	Management:  management,
	RoutePrefix:  server.RoutePrefix,
})

return httpserver.Run(
	ctx,
	server.ListenAddress,
	server.ManagementListenAddress,
	handlers,
	logger.With("server", "application"),
	logger.With("server", "management"),
)
```

For a runnable end-to-end implementation, see the repository's [`examples/basic`](https://github.com/containeroo/appkit/tree/main/examples/basic) package.
