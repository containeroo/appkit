package web

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"testing/fstest"
)

func TestSPAIndexAndMissingAssets(t *testing.T) {
	handler := (SPA{Assets: fstest.MapFS{"index.html": {Data: []byte(`<base href="{{ .BaseHref }}"><meta name="route" content="{{ .RoutePrefix }}">`)}}, BasePath: "/motus"}).Handler()
	for _, tc := range []struct {
		path, accept string
		want         int
	}{
		{"/", "text/html", 200}, {"/training", "text/html", 200}, {"/assets/missing.js", "text/html", 404}, {"/api/missing", "application/json", 404},
	} {
		request := httptest.NewRequest(http.MethodGet, tc.path, nil)
		request.Header.Set("Accept", tc.accept)
		response := httptest.NewRecorder()
		handler.ServeHTTP(response, request)
		if response.Code != tc.want {
			t.Errorf("%s: got %d want %d", tc.path, response.Code, tc.want)
		}
		if tc.path == "/" && response.Body.String() != `<base href="/motus/"><meta name="route" content="/motus">` {
			t.Errorf("wrong template data: %s", response.Body.String())
		}
	}
}
