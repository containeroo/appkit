// Package httpauth bridges application-owned users and authenticators into net/http.
// The same authentication middleware can serve JSON APIs or redirecting browser pages.
package httpauth

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/url"

	"github.com/containeroo/appkit/auth"
)

// contextKey keeps user identity separate between distinct generic user types.
type contextKey[User any] struct{}

// WithUser attaches a resolved application user to the request context.
func WithUser[User any](r *http.Request, user User) *http.Request {
	return r.WithContext(context.WithValue(r.Context(), contextKey[User]{}, user))
}

// CurrentUser returns the current authenticated application user.
func CurrentUser[User any](r *http.Request) (User, bool) {
	user, ok := r.Context().Value(contextKey[User]{}).(User)
	return user, ok
}

// OnFailure selects the browser or API response for unsuccessful authentication.
// Implementations must not expose internal authentication errors to clients.
type OnFailure func(http.ResponseWriter, *http.Request, int)

// JSONFailure sends a small, consistent JSON error without redirecting clients.
func JSONFailure(w http.ResponseWriter, _ *http.Request, status int) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.Header().Set("Cache-Control", "no-store")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(map[string]string{"error": http.StatusText(status)})
}

// RedirectFailure redirects unauthenticated browser requests to a local login path.
// Forbidden and internal errors are never redirected.
func RedirectFailure(loginPath string) OnFailure {
	return func(w http.ResponseWriter, r *http.Request, status int) {
		if status != http.StatusUnauthorized {
			JSONFailure(w, r, status)
			return
		}
		if !isLocalPath(loginPath) {
			JSONFailure(w, r, http.StatusInternalServerError)
			return
		}
		destination := loginPath
		query := url.Values{"next": {r.URL.RequestURI()}}
		separator := "?"
		if len(destination) > 0 && destination[len(destination)-1] == '?' {
			separator = ""
		} else if containsQuestionMark(destination) {
			separator = "&"
		}
		http.Redirect(w, r, destination+separator+query.Encode(), http.StatusSeeOther)
	}
}

// containsQuestionMark reports whether a path includes an existing query.
func containsQuestionMark(path string) bool {
	for i := range path {
		if path[i] == '?' {
			return true
		}
	}
	return false
}

// isLocalPath accepts only absolute local paths without protocol-relative syntax.
func isLocalPath(path string) bool {
	return len(path) > 0 && path[0] == '/' && (len(path) == 1 || path[1] != '/') && !containsBackslash(path) && !containsControl(path)
}

// containsBackslash rejects ambiguous redirect paths across URL parsers.
func containsBackslash(path string) bool {
	for i := range path {
		if path[i] == '\\' {
			return true
		}
	}
	return false
}

// Require tries authenticators in order, passes the resolved user to the next handler,
// and distinguishes expected denial from unexpected authentication failure.
func Require[User any](onFailure OnFailure, authenticators ...auth.Authenticator[User]) func(http.Handler) http.Handler {
	if onFailure == nil {
		onFailure = JSONFailure
	}
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if len(authenticators) == 0 {
				onFailure(w, r, http.StatusInternalServerError)
				return
			}
			for _, authenticator := range authenticators {
				user, err := authenticator.Authenticate(r)
				if err == nil {
					next.ServeHTTP(w, WithUser(r, user))
					return
				}
				if errors.Is(err, auth.ErrForbidden) {
					onFailure(w, r, http.StatusForbidden)
					return
				}
				if errors.Is(err, auth.ErrInvalidCredentials) {
					onFailure(w, r, http.StatusUnauthorized)
					return
				}
				if !errors.Is(err, auth.ErrUnauthenticated) {
					onFailure(w, r, http.StatusInternalServerError)
					return
				}
			}
			onFailure(w, r, http.StatusUnauthorized)
		})
	}
}

// Authorize applies application-owned authorization policy to the resolved user.
func Authorize[User any](allowed func(*http.Request, User) bool, onFailure OnFailure) func(http.Handler) http.Handler {
	if onFailure == nil {
		onFailure = JSONFailure
	}
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			user, ok := CurrentUser[User](r)
			if !ok {
				onFailure(w, r, http.StatusUnauthorized)
				return
			}
			if allowed == nil || !allowed(r, user) {
				onFailure(w, r, http.StatusForbidden)
				return
			}
			next.ServeHTTP(w, r)
		})
	}
}

// containsControl reports whether a redirect target includes forbidden control characters.
func containsControl(path string) bool {
	for _, ch := range path {
		if ch < 0x20 || ch == 0x7f {
			return true
		}
	}
	return false
}
