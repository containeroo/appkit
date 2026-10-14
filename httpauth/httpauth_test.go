package httpauth

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/containeroo/appkit/auth"
)

type user struct {
	Name  string
	Admin bool
}

func TestRequireAndAuthorize(t *testing.T) {
	identity := auth.AuthenticatorFunc[user](func(*http.Request) (user, error) { return user{Name: "alice", Admin: true}, nil })
	handler := Require(JSONFailure, identity)(Authorize[user](func(_ *http.Request, u user) bool { return u.Admin }, JSONFailure)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		actual, ok := CurrentUser[user](r)
		if !ok || actual.Name != "alice" {
			t.Errorf("unexpected user: %#v", actual)
		}
		w.WriteHeader(http.StatusNoContent)
	})))
	recorder := httptest.NewRecorder()
	handler.ServeHTTP(recorder, httptest.NewRequest("GET", "/api/me", nil))
	if recorder.Code != http.StatusNoContent {
		t.Fatalf("status: %d", recorder.Code)
	}
}

func TestCurrentUserFromContext(t *testing.T) {
	request := WithUser(httptest.NewRequest(http.MethodGet, "/", nil), user{Name: "alice"})

	actual, ok := CurrentUserFromContext[user](request.Context())
	if !ok {
		t.Fatal("user missing from context")
	}
	if actual.Name != "alice" {
		t.Fatalf("user name: %q", actual.Name)
	}
}

func TestAuthenticateErrors(t *testing.T) {
	cases := []struct {
		failure error
		want    int
	}{
		{auth.ErrUnauthenticated, 401}, {auth.ErrInvalidCredentials, 401}, {auth.ErrForbidden, 403}, {errors.New("DB unavailable"), 500},
	}
	for _, tc := range cases {
		identity := auth.AuthenticatorFunc[user](func(*http.Request) (user, error) { return user{}, tc.failure })
		response := httptest.NewRecorder()
		Require(JSONFailure, identity)(http.HandlerFunc(func(http.ResponseWriter, *http.Request) { t.Error("must not run") })).ServeHTTP(response, httptest.NewRequest("GET", "/api/me", nil))
		if response.Code != tc.want {
			t.Fatalf("failure %s produced %d, want %d", tc.failure, response.Code, tc.want)
		}
		if tc.want == 500 && response.Body.String() == "DB unavailable" {
			t.Fatal("internal error leaked")
		}
	}
}

func TestAuthenticateFallsBackOnlyOnMissingCredentials(t *testing.T) {
	first := auth.AuthenticatorFunc[user](func(*http.Request) (user, error) { return user{}, auth.ErrUnauthenticated })
	second := auth.AuthenticatorFunc[user](func(*http.Request) (user, error) { return user{Name: "fallback"}, nil })
	response := httptest.NewRecorder()
	Require(JSONFailure, first, second)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		u, _ := CurrentUser[user](r)
		if u.Name != "fallback" {
			t.Errorf("unexpected user %q", u.Name)
		}
		w.WriteHeader(204)
	})).ServeHTTP(response, httptest.NewRequest("GET", "/api/me", nil))
	if response.Code != 204 {
		t.Fatal(response.Code)
	}
}

func TestRedirectFailure(t *testing.T) {
	response := httptest.NewRecorder()
	RedirectFailure("/auth/login")(response, httptest.NewRequest("GET", "/reports/last?view=full", nil), 401)
	if response.Code != 303 || response.Header().Get("Location") != "/auth/login?next=%2Freports%2Flast%3Fview%3Dfull" {
		t.Fatalf("unexpected redirect: %d %q", response.Code, response.Header().Get("Location"))
	}
	response = httptest.NewRecorder()
	RedirectFailure("//attacker.example")(response, httptest.NewRequest("GET", "/", nil), 401)
	if response.Code != 500 {
		t.Fatal("unsafe redirect accepted")
	}
}

func TestAuthorizeWithError(t *testing.T) {
	for _, tc := range []struct {
		user    bool
		allowed bool
		failure error
		want    int
	}{
		{true, true, nil, 204}, {true, false, nil, 403}, {true, false, errors.New("database unavailable"), 500}, {false, true, nil, 401},
	} {
		request := httptest.NewRequest(http.MethodGet, "/admin", nil)
		if tc.user {
			request = WithUser(request, user{Name: "alice"})
		}
		handler := AuthorizeWithError[user](func(*http.Request, user) (bool, error) { return tc.allowed, tc.failure }, JSONFailure, nil)(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) { w.WriteHeader(204) }))
		response := httptest.NewRecorder()
		handler.ServeHTTP(response, request)
		if response.Code != tc.want {
			t.Errorf("allowed=%t failure=%v: got %d want %d", tc.allowed, tc.failure, response.Code, tc.want)
		}
	}
}
