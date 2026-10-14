package local

import (
	"context"
	"errors"
	"strings"

	"github.com/containeroo/appkit/auth"
	"github.com/containeroo/appkit/auth/session"
)

// CredentialStore loads an application-owned user and a bcrypt password hash.
// Unknown or disabled accounts must return auth.ErrInvalidCredentials.
type CredentialStore[User any] interface {
	// LookupCredential fetches a user's current bcrypt password hash.
	LookupCredential(context.Context, string) (User, string, error)
}

// Service verifies local passwords and issues hashed server-side sessions.
type Service[User any] struct {
	// Credentials provides application-owned user lookup and account policy.
	Credentials CredentialStore[User]
	// Sessions provides durable, revocable session persistence.
	Sessions session.Manager[User]
}

// SignIn verifies credentials and creates a new opaque session.
// The caller writes the session cookie using session.CookieConfig.
func (s Service[User]) SignIn(ctx context.Context, username, password string) (User, string, error) {
	var zero User
	if s.Credentials == nil {
		return zero, "", errors.New("credential store is required")
	}
	user, hash, err := s.Credentials.LookupCredential(ctx, strings.TrimSpace(username))
	if errors.Is(err, auth.ErrUnauthenticated) || errors.Is(err, auth.ErrInvalidCredentials) {
		return zero, "", auth.ErrInvalidCredentials
	}
	if err != nil {
		return zero, "", err
	}
	if err := VerifyPassword(hash, password); err != nil {
		return zero, "", err
	}
	token, err := s.Sessions.Start(ctx, user)
	if err != nil {
		return zero, "", err
	}
	return user, token, nil
}
