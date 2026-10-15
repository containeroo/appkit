# Local authentication

The `auth/local` package separates password verification from application persistence. Your application implements `CredentialStore`, which returns a user and a bcrypt password hash for an account; the appkit service verifies the password and starts a session.

## Implement the credential store

Return `auth.ErrInvalidCredentials` when the account is unknown or disabled. Do not reveal whether a username exists.

```go
type credentialStore struct {
	db *sql.DB
}

func (s credentialStore) LookupCredential(
	ctx context.Context,
	username string,
) (User, string, error) {
	// Load the application user and bcrypt hash from your database.
	// Return auth.ErrInvalidCredentials for absent or disabled accounts.
}
```

## Sign in and issue a session

```go
service := local.Service[User]{
	Credentials: credentialStore{db: db},
	Sessions:    sessions,
}

user, token, err := service.SignIn(r.Context(), form.Username, form.Password)
if err != nil {
	return err
}
http.SetCookie(w, cookieConfig.Cookie(token, false))
```

`SignIn` trims the username, maps missing or rejected credentials to `auth.ErrInvalidCredentials`, verifies the bcrypt hash, and returns a newly generated opaque session token. Store the token only in the browser cookie; session persistence receives a hash, as described in the [sessions guide](/appkit/guides/sessions/).

## Password lifecycle

Create new hashes with `local.HashPassword`. Apply your own rules—minimum length, breach screening, MFA requirements, account lockout, and password-reset flow—before storing or accepting a password. Appkit intentionally does not prescribe those domain policies.
