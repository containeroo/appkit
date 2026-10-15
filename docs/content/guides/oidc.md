# OIDC

`auth/oidc` performs an authorization-code flow with PKCE, verifies the ID token, and returns an `auth.ExternalIdentity`. Your application resolves that identity and creates its own session.

```go
flow, err := oidc.New(ctx, oidc.Config{
	Issuer:       "https://id.example.com",
	ClientID:     "appkit-example",
	ClientSecret: clientSecret,
	StateSecret:  stateSecret, // at least 32 bytes
	CallbackURL:  "https://app.example.com/auth/callback",
})
if err != nil {
	return err
}

// Login handler: flow.Begin(w, r, "/")
// Callback handler: identity, next, err := flow.Complete(w, r)
```

Use HTTPS for the callback in deployed environments. HTTP callbacks are accepted only for `localhost`, `127.0.0.1`, and `::1` development targets. The requested post-login destination is restricted to a local path.
