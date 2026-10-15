package web

import (
	"io/fs"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"testing/fstest"
)

// TestSPAIndexAndMissingAssets verifies navigation fallback without swallowing missing files.
func TestSPAIndexAndMissingAssets(t *testing.T) {
	t.Parallel()
	handler := (SPA{Assets: testSPAAssets(), BasePath: "/motus"}).Handler()
	for _, tc := range []struct {
		name   string
		path   string
		accept string
		want   int
		body   string
		cache  string
	}{
		{name: "root", path: "/", accept: "text/html", want: 200, body: `<base href="/motus/"><meta name="route" content="/motus">`, cache: "no-cache"},
		{name: "index file", path: "/index.html", accept: "text/html", want: 200, body: `<base href="/motus/">`, cache: "no-cache"},
		{name: "navigation", path: "/training", accept: "text/html", want: 200, body: `<base href="/motus/">`, cache: "no-cache"},
		{name: "navigation trailing slash", path: "/training/", accept: "text/html", want: 200, body: `<base href="/motus/">`, cache: "no-cache"},
		{name: "navigation without accept", path: "/settings", want: 200, body: `<base href="/motus/">`},
		{name: "svg at root", path: "/app.svg", want: 200, body: "<svg>logo</svg>", cache: "no-cache"},
		{name: "favicon at root", path: "/favicon.ico", want: 200, body: "ICO", cache: "no-cache"},
		{name: "manifest at root", path: "/manifest.webmanifest", want: 200, body: "{}"},
		{name: "versioned bundle", path: "/assets/main-12345.js", want: 200, body: "console.log(1)", cache: "public, max-age=31536000, immutable"},
		{name: "missing svg", path: "/missing.svg", accept: "text/html", want: 404},
		{name: "asset with slash suffix", path: "/app.svg/", accept: "text/html", want: 404},
		{name: "double slash", path: "//", accept: "text/html", want: 404},
		{name: "missing hashed bundle", path: "/assets/missing.js", accept: "text/html", want: 404},
		{name: "missing bundle without extension", path: "/assets/missing", accept: "text/html", want: 404},
		{name: "API without HTML accept", path: "/api/missing", accept: "application/json", want: 404},
		{name: "existing directory", path: "/assets", accept: "text/html", want: 404},
		{name: "hidden file", path: "/.secret", accept: "text/html", want: 404},
		{name: "traversal", path: "/../secret.txt", accept: "text/html", want: 404},
		{name: "nested traversal", path: "/assets/../app.svg", accept: "text/html", want: 404},
		{name: "encoded backslash", path: "/assets%5cmain.js", want: 404},
	} {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			request := httptest.NewRequest(http.MethodGet, tc.path, nil)
			if tc.accept != "" {
				request.Header.Set("Accept", tc.accept)
			}
			response := httptest.NewRecorder()
			handler.ServeHTTP(response, request)
			if response.Code != tc.want {
				t.Fatalf("%s: status = %d, want %d (%s)", tc.path, response.Code, tc.want, response.Body.String())
			}
			if tc.body != "" && !strings.Contains(response.Body.String(), tc.body) {
				t.Fatalf("%s: body %q does not contain %q", tc.path, response.Body.String(), tc.body)
			}
			if tc.cache != "" && response.Header().Get("Cache-Control") != tc.cache {
				t.Fatalf("%s: Cache-Control = %q, want %q", tc.path, response.Header().Get("Cache-Control"), tc.cache)
			}
		})
	}
}

// TestSPAHeadAndMethods checks HEAD handling and HTTP method rejection.
func TestSPAHeadAndMethods(t *testing.T) {
	t.Parallel()
	handler := (SPA{Assets: testSPAAssets()}).Handler()
	for _, tc := range []struct {
		method, path string
		want         int
	}{
		{http.MethodHead, "/app.svg", http.StatusOK},
		{http.MethodHead, "/", http.StatusOK},
		{http.MethodPost, "/app.svg", http.StatusMethodNotAllowed},
		{http.MethodPost, "/", http.StatusMethodNotAllowed},
	} {
		response := httptest.NewRecorder()
		handler.ServeHTTP(response, httptest.NewRequest(tc.method, tc.path, nil))
		if response.Code != tc.want {
			t.Errorf("%s %s: status = %d, want %d", tc.method, tc.path, response.Code, tc.want)
		}
		if tc.method == http.MethodHead && response.Body.Len() != 0 {
			t.Errorf("HEAD %s unexpectedly wrote %q", tc.path, response.Body.String())
		}
	}
}

// TestSPANestedIndex serves static files relative to the embedded index directory.
func TestSPANestedIndex(t *testing.T) {
	t.Parallel()
	handler := (SPA{Assets: fstest.MapFS{
		"dist/index.html": {Data: []byte("nested index")},
		"dist/icon.svg":   {Data: []byte("nested icon")},
		"icon.svg":        {Data: []byte("not public")},
	}, Index: "dist/index.html"}).Handler()
	response := httptest.NewRecorder()
	handler.ServeHTTP(response, httptest.NewRequest(http.MethodGet, "/icon.svg", nil))
	if response.Code != http.StatusOK || response.Body.String() != "nested icon" {
		t.Fatalf("response = %d %q", response.Code, response.Body.String())
	}
}

// TestSPAMissingIndex returns an unavailable response when frontend build is missing.
func TestSPAMissingIndex(t *testing.T) {
	t.Parallel()
	handler := (SPA{Assets: fstest.MapFS{}}).Handler()
	response := httptest.NewRecorder()
	handler.ServeHTTP(response, httptest.NewRequest(http.MethodGet, "/", nil))
	if response.Code != http.StatusServiceUnavailable {
		t.Fatalf("status = %d, want %d", response.Code, http.StatusServiceUnavailable)
	}
}

// TestSPAErrorsOnUnreadableFiles distinguishes filesystem failures from not-found.
func TestSPAErrorsOnUnreadableFiles(t *testing.T) {
	t.Parallel()
	fsWithError := fstest.MapFS{
		"index.html": {Data: []byte("index")},
		"denied.svg": {Mode: fs.ModeIrregular},
	}
	handler := (SPA{Assets: fsWithError}).Handler()
	response := httptest.NewRecorder()
	handler.ServeHTTP(response, httptest.NewRequest(http.MethodGet, "/denied.svg", nil))
	if response.Code != http.StatusNotFound {
		t.Fatalf("status = %d, want %d", response.Code, http.StatusNotFound)
	}
}

// testSPAAssets provides static assets, hidden files, and a templated index for tests.
func testSPAAssets() fstest.MapFS {
	return fstest.MapFS{
		"index.html":           {Data: []byte(`<base href="{{ .BaseHref }}"><meta name="route" content="{{ .RoutePrefix }}">`)},
		"app.svg":              {Data: []byte("<svg>logo</svg>")},
		"favicon.ico":          {Data: []byte("ICO")},
		"manifest.webmanifest": {Data: []byte("{}")},
		"assets/main-12345.js": {Data: []byte("console.log(1)")},
		".secret":              {Data: []byte("not public")},
	}
}
