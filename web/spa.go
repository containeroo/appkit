// Package web serves embedded single-page applications with optional Go-template index data.
package web

import (
	"bytes"
	"io/fs"
	"net/http"
	"path"
	"strings"
	"text/template"
	"time"
)

// SPA describes a frontend build and its externally mounted base path.
type SPA struct {
	// Assets is rooted at the web distribution directory.
	Assets fs.FS
	// Index defaults to index.html.
	Index string
	// BasePath is the URL mount prefix, for example /motus/.
	BasePath string
	// TemplateData provides optional extra values for index.html.
	TemplateData map[string]any
}

// Handler serves index for browser navigations and rejects missing static assets.
func (s SPA) Handler() http.Handler {
	index := s.Index
	if index == "" {
		index = "index.html"
	}
	base := s.BasePath
	if base == "" {
		base = "/"
	}
	if !strings.HasSuffix(base, "/") {
		base += "/"
	}
	routePrefix := strings.TrimSuffix(base, "/")
	values := map[string]any{"BaseHref": base, "RoutePrefix": routePrefix}
	for key, val := range s.TemplateData {
		values[key] = val
	}
	data, err := fs.ReadFile(s.Assets, index)
	if err != nil {
		return http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
			http.Error(w, "app not built (run frontend build)", http.StatusServiceUnavailable)
		})
	}
	tmpl, err := template.New(index).Parse(string(data))
	if err != nil {
		return http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
			http.Error(w, "invalid SPA index template", http.StatusInternalServerError)
		})
	}
	var rendered bytes.Buffer
	if err := tmpl.Execute(&rendered, values); err != nil {
		return http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
			http.Error(w, "unable to render SPA index", http.StatusInternalServerError)
		})
	}
	info, _ := fs.Stat(s.Assets, index)
	modified := time.Time{}
	if info != nil {
		modified = info.ModTime()
	}
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet && r.Method != http.MethodHead {
			w.Header().Set("Allow", "GET, HEAD")
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}
		name := strings.TrimPrefix(r.URL.Path, "/")
		if name != "" && name != index && (path.Ext(name) != "" || strings.Contains(name, "..")) {
			http.NotFound(w, r)
			return
		}
		accept := r.Header.Get("Accept")
		if name != "" && name != index && accept != "" && !strings.Contains(accept, "text/html") {
			http.NotFound(w, r)
			return
		}
		w.Header().Set("Cache-Control", "no-cache")
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		http.ServeContent(w, r, index, modified, bytes.NewReader(rendered.Bytes()))
	})
}
