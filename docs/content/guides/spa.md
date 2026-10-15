# Single-page applications

`web.SPA` serves an embedded frontend build while keeping browser navigation safe. It serves public files from the embedded filesystem—including root-level icons, manifests, and files in `assets/`—before trying SPA navigation. Missing assets return 404 instead of the SPA index. There is no need to register individual public files or `/assets/` routes in the application.

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

If `Index` is `dist/index.html`, public files are resolved relative to `dist/`: `/favicon.ico` serves `dist/favicon.ico` from the embedded filesystem. With an already-rooted filesystem, omit `Index` or set it to `index.html`. The index is parsed as a Go `text/template`, enabling a small amount of server-provided configuration at render time. Keep template data non-sensitive: rendered values are sent to every browser that loads the application.

`Handler` supports `GET` and `HEAD` only and never serves directory listings or dotfiles. HTML navigation returns the rendered index with `Cache-Control: no-cache`; root-level public files also use `no-cache`. Build assets under `assets/` receive `public, max-age=31536000, immutable`: **only place fingerprinted files in that directory**, because clients may cache them for one year. All missing asset paths, including missing entries under `assets/`, return 404. Missing index files return 503. Put API handlers on their own routes so an API 404 is never converted into the SPA index page.
