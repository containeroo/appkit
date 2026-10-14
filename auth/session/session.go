// Package session implements opaque, purpose-bound, hashed browser sessions.
// It is storage-agnostic: applications implement Store using their own database.
package session

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/containeroo/appkit/auth"
)

// ErrNotFound indicates an absent or invalidated session.
var ErrNotFound = errors.New("session not found")

// Store persists only hashes of bearer session tokens, never raw tokens.
// Lookup must exclude expired and revoked sessions and disabled accounts.
type Store[User any] interface {
	// Create atomically persists a session hash and its expiry.
	Create(context.Context, string, User, time.Time) error
	// Lookup resolves a live session hash to an application user.
	Lookup(context.Context, string) (User, error)
	// Revoke permanently invalidates a session hash.
	Revoke(context.Context, string) error
}

// Manager issues, resolves, and revokes sessions for one application.
type Manager[User any] struct {
	// Store is the application-supplied durable session store.
	Store Store[User]
	// Lifetime is the maximum age of newly issued sessions.
	Lifetime time.Duration
	// Purpose isolates hashes from other types of token.
	Purpose string
}

// Validate ensures that session configuration cannot issue unsafe sessions.
func (m Manager[User]) Validate() error {
	if m.Store == nil {
		return errors.New("session store is required")
	}
	if m.Lifetime <= 0 {
		return errors.New("session lifetime must be positive")
	}
	if m.Purpose == "" {
		return errors.New("session purpose is required")
	}
	return nil
}

// TokenHash derives the purpose-separated identifier stored in the database.
func TokenHash(purpose, token string) string {
	digest := sha256.Sum256([]byte(purpose + "\x00" + token))
	return hex.EncodeToString(digest[:])
}

// Start creates a cryptographically random token and persists only its hash.
func (m Manager[User]) Start(ctx context.Context, user User) (string, error) {
	if err := m.Validate(); err != nil {
		return "", err
	}
	bytes := make([]byte, 32)
	if _, err := rand.Read(bytes); err != nil {
		return "", fmt.Errorf("generate session token: %w", err)
	}
	token := base64.RawURLEncoding.EncodeToString(bytes)
	if err := m.Store.Create(ctx, TokenHash(m.Purpose, token), user, time.Now().Add(m.Lifetime)); err != nil {
		return "", fmt.Errorf("persist session: %w", err)
	}
	return token, nil
}

// Resolve authenticates the session token using application-owned persistence.
func (m Manager[User]) Resolve(ctx context.Context, token string) (User, error) {
	var zero User
	if err := m.Validate(); err != nil {
		return zero, err
	}
	if token == "" {
		return zero, auth.ErrUnauthenticated
	}
	user, err := m.Store.Lookup(ctx, TokenHash(m.Purpose, token))
	if errors.Is(err, ErrNotFound) {
		return zero, auth.ErrUnauthenticated
	}
	return user, err
}

// End revokes one session without storing or logging the raw token.
func (m Manager[User]) End(ctx context.Context, token string) error {
	if err := m.Validate(); err != nil {
		return err
	}
	if token == "" {
		return nil
	}
	return m.Store.Revoke(ctx, TokenHash(m.Purpose, token))
}

// CookieConfig defines an HTTP-only cookie scoped to the application's deployment path.
type CookieConfig struct {
	// Name is the session cookie name, unique within the deployment.
	Name string
	// Path is the cookie path, typically the mounted application prefix or root.
	Path string
	// Lifetime is the browser cookie lifetime.
	Lifetime time.Duration
	// Secure must be true for HTTPS deployments.
	Secure bool
	// SameSite is the cookie cross-site policy.
	SameSite http.SameSite
}

// Cookie returns an HTTP-only cookie using configured security settings.
func (c CookieConfig) Cookie(token string, expire bool) *http.Cookie {
	path := c.Path
	if path == "" {
		path = "/"
	}
	sameSite := c.SameSite
	if sameSite == 0 || sameSite == http.SameSiteDefaultMode {
		sameSite = http.SameSiteLaxMode
	}
	cookie := &http.Cookie{Name: c.Name, Value: token, Path: path, Secure: c.Secure,
		HttpOnly: true, SameSite: sameSite}
	if expire {
		cookie.Value = ""
		cookie.MaxAge = -1
	} else {
		cookie.MaxAge = int(c.Lifetime.Seconds())
	}
	return cookie
}

// Authenticator resolves users from a named session cookie.
type Authenticator[User any] struct {
	// Manager provides secure session lookup.
	Manager Manager[User]
	// CookieName is the cookie carrying the opaque session token.
	CookieName string
}

// Authenticate implements the shared auth.Authenticator contract.
func (a Authenticator[User]) Authenticate(request *http.Request) (User, error) {
	cookie, err := request.Cookie(a.CookieName)
	if err != nil {
		var zero User
		return zero, auth.ErrUnauthenticated
	}
	return a.Manager.Resolve(request.Context(), cookie.Value)
}
