// Package web serves embedded single-page applications with optional Go-template index data.
package web

import (
	"bytes"
	"errors"
	"io/fs"
	"net/http"
	"path"
	"strings"
	"text/template"
	"time"
)

const (
	// Static files emitted under assets/ are fingerprinted build outputs.
	immutableAssetCache = "public, max-age=31536000, immutable"
	indexCache          = "no-cache"
)

// SPA describes a frontend build and its externally mounted base path.
type SPA struct {
	// Assets is rooted at the web distribution directory, unless Index names a subdirectory.
	Assets fs.FS
	// Index defaults to index.html; its directory also locates public static files.
	Index string
	// BasePath is the URL mount prefix, for example /motus/.
	BasePath string
	// TemplateData provides optional extra values for index.html.
	TemplateData map[string]any
}

// Handler serves public static files and falls back to index for browser navigation.
// Missing files and unsafe paths never fall back to index.html.
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
		return spaFailure("app not built (run frontend build)", http.StatusServiceUnavailable)
	}
	tmpl, err := template.New(index).Parse(string(data))
	if err != nil {
		return spaFailure("invalid SPA index template", http.StatusInternalServerError)
	}
	var rendered bytes.Buffer
	if err := tmpl.Execute(&rendered, values); err != nil {
		return spaFailure("unable to render SPA index", http.StatusInternalServerError)
	}

	// When Index is e.g. dist/index.html, the public URLs still start at /.
	assets, err := fs.Sub(s.Assets, path.Dir(index))
	if err != nil {
		return spaFailure("invalid SPA asset directory", http.StatusInternalServerError)
	}
	files := http.FileServer(http.FS(assets))
	indexName := path.Base(index)
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
		trailingSlash := strings.HasSuffix(name, "/")
		lookupName := strings.TrimSuffix(name, "/")
		if (name != "" && lookupName == "") || !safeSPAPath(lookupName) {
			http.NotFound(w, r)
			return
		}
		if name != "" && name != indexName {
			// Serve only actual regular files; never expose directory listings.
			fileInfo, err := fs.Stat(assets, lookupName)
			if err == nil {
				if trailingSlash || !fileInfo.Mode().IsRegular() {
					http.NotFound(w, r)
					return
				}
				w.Header().Set("Cache-Control", spaStaticCache(name))
				files.ServeHTTP(w, r)
				return
			}
			if !errors.Is(err, fs.ErrNotExist) {
				http.Error(w, "unable to read SPA assets", http.StatusInternalServerError)
				return
			}
			// A missing static asset is never an HTML navigation.
			if path.Ext(lookupName) != "" || lookupName == "assets" || strings.HasPrefix(lookupName, "assets/") {
				http.NotFound(w, r)
				return
			}
			accept := r.Header.Get("Accept")
			if accept != "" && !strings.Contains(accept, "text/html") {
				http.NotFound(w, r)
				return
			}
		}
		w.Header().Set("Cache-Control", indexCache)
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		http.ServeContent(w, r, indexName, modified, bytes.NewReader(rendered.Bytes()))
	})
}

// spaFailure constructs the handler returned when SPA startup configuration is invalid.
func spaFailure(message string, status int) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		http.Error(w, message, status)
	})
}

// safeSPAPath rejects dot segments, hidden files, and ambiguous or invalid paths.
func safeSPAPath(name string) bool {
	if name == "" {
		return true
	}
	if !fs.ValidPath(name) || strings.ContainsRune(name, '\\') {
		return false
	}
	for _, segment := range strings.Split(name, "/") {
		if strings.HasPrefix(segment, ".") {
			return false
		}
	}
	return true
}

// spaStaticCache retains immutable caching for Vite assets and revalidates other files.
func spaStaticCache(name string) string {
	if strings.HasPrefix(name, "assets/") {
		return immutableAssetCache
	}
	return indexCache
}
