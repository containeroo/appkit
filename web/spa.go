// Package web serves embedded SPA assets and falls back to index HTML only for navigation requests.
// The host application remains responsible for registering API and authentication routes.
package web

import (
	"bytes"
	"io/fs"
	"mime"
	"net/http"
	"path"
	"strings"
	"time"
)

// SPA configures an embedded single-page application handler.
type SPA struct {
	// Assets contains the frontend build files rooted at its distribution directory.
	Assets fs.FS
	// Index is the entry point, typically index.html.
	Index string
	// BasePath is the externally mounted prefix ending in a slash (or / at root).
	BasePath string
}

// Handler serves assets with correct content types and an index fallback for browser navigation.
// Mount it under a prefix using httpkit/prefix; API routes must be more specific.
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
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet && r.Method != http.MethodHead {
			w.Header().Set("Allow", "GET, HEAD")
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}
		name := strings.TrimPrefix(r.URL.Path, "/")
		if name == "" {
			name = index
		}
		if !fs.ValidPath(name) {
			http.NotFound(w, r)
			return
		}
		data, err := fs.ReadFile(s.Assets, name)
		if err != nil {
			// Returning HTML for a missing JS/CSS asset would mask broken frontend builds.
			if path.Ext(name) != "" || !strings.Contains(r.Header.Get("Accept"), "text/html") {
				http.NotFound(w, r)
				return
			}
			name = index
			data, err = fs.ReadFile(s.Assets, index)
			if err != nil {
				http.NotFound(w, r)
				return
			}
		}
		if name == index {
			data = bytes.ReplaceAll(data, []byte("{{.BaseHref}}"), []byte(base))
			w.Header().Set("Cache-Control", "no-cache")
			w.Header().Set("Content-Type", "text/html; charset=utf-8")
		} else {
			if contentType := mime.TypeByExtension(path.Ext(name)); contentType != "" {
				w.Header().Set("Content-Type", contentType)
			}
			w.Header().Set("Cache-Control", "public, max-age=3600")
		}
		http.ServeContent(w, r, name, time.Time{}, bytes.NewReader(data))
	})
}
