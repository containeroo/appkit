# appkit

A composable Go framework for HTTP applications, including configuration, authentication, sessions, logging, encryption, and web serving.

## Install

```sh
go get github.com/containeroo/appkit
```

## Example

Define your application and management routes, then serve them independently:

```go
app := http.NewServeMux()
app.HandleFunc("GET /", home)

management := http.NewServeMux()
management.HandleFunc("GET /healthz", health)
management.HandleFunc("GET /readyz", readiness)

serverConfig := httpserver.Config{
    Application: app,
    Management:  management,
    RoutePrefix: cfg.Server.RoutePrefix,
}
handlers := httpserver.New(serverConfig)
if err := httpserver.Run(
    ctx,
    cfg.Server.ListenAddress,
    cfg.Server.ManagementListenAddress,
    handlers,
    logger.With("component", "server"),
); err != nil {
    logger.Error("run server", "event", "server_run_failed", "error", err)
    return err
}
```

See [`examples/basic`](examples/basic/main.go) for a complete runnable example with configuration flags, logging, graceful shutdown, and both listeners.

```sh
go run ./examples/basic --route-prefix /app

curl http://127.0.0.1:8080/app/
curl -i http://127.0.0.1:8081/healthz
```
