// Package auth defines application-independent authentication contracts and errors.
// An application owns its user model, database, account provisioning and authorization policy.
package auth

import (
	"errors"
	"net/http"
)

var (
	// ErrUnauthenticated indicates that the request carries no usable credentials.
	ErrUnauthenticated = errors.New("authentication required")
	// ErrInvalidCredentials indicates credentials were presented but rejected.
	ErrInvalidCredentials = errors.New("invalid credentials")
	// ErrForbidden indicates an authenticated identity is not permitted to proceed.
	ErrForbidden = errors.New("access denied")
)

// Mode identifies an optional deployment authentication override.
// Stored application authentication policy remains owned by its application.
type Mode string

const (
	// ModeNone disables enforcement (only for explicitly trusted environments).
	ModeNone Mode = "none"
	// ModeLocal uses application local credentials and server-side sessions.
	ModeLocal Mode = "local"
	// ModeTrustedProxy supports deployments using the older trusted-proxy value.
	ModeTrustedProxy Mode = "trusted-proxy"
	// ModeProxy uses identities asserted by an independently trusted reverse proxy.
	ModeProxy Mode = "proxy"
	// ModeOIDC uses an OpenID Connect provider.
	ModeOIDC Mode = "oidc"
)

// Authenticator resolves the application's own user type from an HTTP request.
// Returning ErrUnauthenticated allows another authenticator to be tried.
type Authenticator[User any] interface {
	// Authenticate returns the application user represented by the request.
	// It returns ErrUnauthenticated when no usable credentials are present, which
	// allows middleware to try a later authenticator; other errors stop the chain.
	Authenticate(*http.Request) (User, error)
}

// AuthenticatorFunc adapts a function into an Authenticator.
type AuthenticatorFunc[User any] func(*http.Request) (User, error)

// Authenticate implements Authenticator for a function.
func (fn AuthenticatorFunc[User]) Authenticate(request *http.Request) (User, error) {
	return fn(request)
}

// ExternalIdentity contains verified attributes from an external identity provider.
// The issuer and subject, not email or display name, form the stable identity key.
type ExternalIdentity struct {
	// Issuer identifies the provider that asserted the subject.
	Issuer string
	// Subject is the stable, provider-issued identity key.
	Subject string
	// Username is a suggested application username.
	Username string
	// Email is an optional email claim, not automatically a verified address.
	Email string
	// EmailVerified indicates that the provider explicitly verified the address.
	EmailVerified bool
	// DisplayName is an optional human-readable name.
	DisplayName string
	// Groups are external group values, not application permissions.
	Groups []string
}
