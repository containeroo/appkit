package web

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"testing/fstest"
)

func TestSPA(t *testing.T) {
	handler := SPA{BasePath: "/motus/", Assets: fstest.MapFS{
		"index.html":    {Data: []byte(`<base href="{{.BaseHref}}">`)},
		"assets/app.js": {Data: []byte(`console.log("hello")`)},
	}}.Handler()
	tests := []struct {
		path, accept string
		status       int
		contains     string
	}{
		{"/", "", 200, `href="/motus/"`},
		{"/training", "text/html", 200, `href="/motus/"`},
		{"/assets/app.js", "", 200, `console.log`},
		{"/assets/missing.js", "text/html", 404, ""},
		{"/missing-route", "application/json", 404, ""},
	}
	for _, tc := range tests {
		record := httptest.NewRecorder()
		request := httptest.NewRequest("GET", tc.path, nil)
		if tc.accept != "" {
			request.Header.Set("Accept", tc.accept)
		}
		handler.ServeHTTP(record, request)
		if record.Code != tc.status {
			t.Errorf("%s: got %d, want %d", tc.path, record.Code, tc.status)
		}
		if tc.contains != "" && !contains(record.Body.String(), tc.contains) {
			t.Errorf("%s: missing %q", tc.path, tc.contains)
		}
	}
	response := httptest.NewRecorder()
	handler.ServeHTTP(response, httptest.NewRequest(http.MethodPost, "/", nil))
	if response.Code != 405 {
		t.Error("POST should be rejected")
	}
}
func contains(text, needle string) bool {
	for i := 0; i+len(needle) <= len(text); i++ {
		if text[i:i+len(needle)] == needle {
			return true
		}
	}
	return false
}
