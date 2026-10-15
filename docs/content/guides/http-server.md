# HTTP server

Keep browser or API traffic separate from operational traffic. `httpserver.New` applies the application route prefix only to the application handler; management endpoints remain on their own listener.

```go
app := http.NewServeMux()
app.HandleFunc("GET /", home)

management := http.NewServeMux()
management.HandleFunc("GET /healthz", health)
management.HandleFunc("GET /readyz", readiness)

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
	logger.With("component", "server"),
); err != nil {
	return err
}
```

The runnable [`examples/basic`](https://github.com/containeroo/appkit/tree/main/examples/basic) program adds common flags, structured logging, graceful shutdown, and both listeners:

```sh
go run ./examples/basic --route-prefix /app
curl http://127.0.0.1:8080/app/
curl -i http://127.0.0.1:8081/healthz
```
