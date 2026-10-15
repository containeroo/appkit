# HTTP server

Keep browser or API traffic separate from operational traffic. `httpserver.New` applies the application route prefix only to the application handler; management endpoints remain on their own listener.

```go
app := http.NewServeMux()
app.HandleFunc("GET /", home)

management := httpserver.NewManagement(httpserver.ManagementRoutes{
	Health:    http.HandlerFunc(health),
	Readiness: http.HandlerFunc(readiness),
	Metrics:   prometheusHandler,
})

handlers := httpserver.New(httpserver.Config{
	Application: app,
	Management:  management,
	RoutePrefix:  cfg.Server.RoutePrefix,
})
if err := httpserver.Run(
	ctx,
	cfg.Server.ListenAddress,
	cfg.Server.ManagementListenAddress,
	handlers,
	logger.With("server", "application"),
	logger.With("server", "management"),
); err != nil {
	return err
}
```

`NewManagement` registers `GET /healthz`, `GET /readyz`, and, when supplied, `GET /metrics`. Go's HTTP router also accepts `HEAD` for these routes and returns `405 Method Not Allowed` for other methods. Omitted handlers are not registered.

The runnable [`examples/basic`](https://github.com/containeroo/appkit/tree/main/examples/basic) program adds common flags, structured logging, graceful shutdown, and both listeners:

```sh
go run ./examples/basic --route-prefix /app
curl http://127.0.0.1:8080/app/
curl -i http://127.0.0.1:8081/healthz
```
