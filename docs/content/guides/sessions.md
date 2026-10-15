# Sessions

`auth/session` creates opaque, random browser tokens. Your database stores only a purpose-separated hash, never the raw bearer token.

```go
manager := session.Manager[User]{
	Store:    users,
	Lifetime: 12 * time.Hour,
	Purpose:  "browser-session",
}

token, err := manager.Start(ctx, user)
if err != nil {
	return err
}
cookie := (session.CookieConfig{
	Name:     "session",
	Path:     "/",
	Lifetime: 12 * time.Hour,
	Secure:   true,
}).Cookie(token, false)
http.SetCookie(w, cookie)
```

Implement `session.Store` so lookup excludes expired and revoked sessions as well as disabled accounts. When a request is unauthenticated, `session.Authenticator` returns `auth.ErrUnauthenticated`, allowing `httpauth.Require` to try another configured authenticator.
