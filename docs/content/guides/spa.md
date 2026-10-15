# Single-page applications

`web.SPA` serves an embedded frontend build while keeping browser navigation safe. It serves static files when they exist and falls back to the configured index page only for HTML navigation requests. Missing assets return a normal 404 instead of the SPA index.

```go
//go:embed dist/*
var assets embed.FS

frontend := web.SPA{
	Assets:   assets,
	Index:    "dist/index.html",
	BasePath: "/app",
	TemplateData: map[string]any{
		"APIBaseURL": "/app/api",
	},
}

mux.Handle("/app/", http.StripPrefix("/app", frontend.Handler()))
```

The index is parsed as a Go `text/template`, enabling a small amount of server-provided configuration at render time. Keep template data non-sensitive: rendered values are sent to every browser that loads the application.

`Handler` supports `GET` and `HEAD` only. It rejects path traversal, returns a service-unavailable response when the frontend was not built, and adds a modification time when the embedded filesystem provides one. Put API handlers on their own routes so an API 404 is never converted into the SPA index page.
