# Middleware and authorization

`httpauth` connects appkit authenticators to `net/http`. It records the resolved application user in the request context and leaves authorization policy to your application.

## Require authentication

Pass a failure strategy and one or more authenticators to `Require`.

```go
protected := httpauth.Require(
	httpauth.RedirectFailure("/login"),
	session.Authenticator[User]{
		Manager:    sessions,
		CookieName: "session",
	},
)(dashboard)
```

Authenticators are tried in order only when they return `auth.ErrUnauthenticated`. A malformed or unavailable identity backend should return its own error, stopping the chain instead of silently falling back to another mechanism.

Use `httpauth.JSONFailure` for APIs. `RedirectFailure` accepts only a local login path and preserves the original local request destination in a `next` parameter; forbidden and internal-error responses are not redirected.

## Read the current user

```go
func dashboard(w http.ResponseWriter, r *http.Request) {
	user, ok := httpauth.CurrentUser[User](r)
	if !ok {
		http.Error(w, "missing authenticated user", http.StatusInternalServerError)
		return
	}
	_ = user
}
```

## Apply authorization

`Authorize` handles a simple allow-or-deny function. For policies backed by a database or another service, use `AuthorizeWithError` to distinguish a denial from an infrastructure failure.

```go
adminOnly := httpauth.AuthorizeWithError(
	func(r *http.Request, user User) (bool, error) {
		return permissions.CanAdmin(r.Context(), user.ID)
	},
	httpauth.JSONFailure,
	func(w http.ResponseWriter, _ *http.Request, err error) {
		logger.Error("authorize request", "error", err)
		http.Error(w, "authorization unavailable", http.StatusServiceUnavailable)
	},
)
```

Never use a proxy group or display-name claim as application authorization without translating it into application-owned policy.
