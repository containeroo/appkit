// Package browser composes browser authentication using resolved application settings.
package browser

import (
	"errors"
	"net/http"

	"github.com/containeroo/appkit/auth"
)

// Origin identifies the method and provider that issued a persisted browser session.
type Origin struct {
	// Mode is local or OIDC.
	Mode auth.Mode
	// Issuer identifies the verified OIDC provider.
	Issuer string
}

// Policy determines which session origins and request authenticators are active.
type Policy struct {
	// Mode selects the primary authentication method.
	Mode auth.Mode
	// Issuer identifies the active OIDC provider.
	Issuer string
	// Recovery permits local password sessions alongside the primary method.
	Recovery bool
}

// LocalEnabled reports whether a local session may be used.
func (p Policy) LocalEnabled() bool { return p.Mode == auth.ModeLocal || p.Recovery }

// Accepts reports whether the current configuration trusts a session's origin.
func (p Policy) Accepts(origin Origin) bool {
	switch origin.Mode {
	case auth.ModeLocal:
		return p.LocalEnabled()
	case auth.ModeOIDC:
		return p.Mode == auth.ModeOIDC && p.Issuer != "" && origin.Issuer == p.Issuer
	default:
		return false
	}
}

// Authenticator checks browser sessions and dispatches to the configured request provider.
type Authenticator[U any] struct {
	// Policy contains the effective application settings.
	Policy Policy
	// Session reads a live account and its authenticated origin from session storage.
	Session func(*http.Request) (U, Origin, error)
	// Primary verifies requests in proxy or explicitly unauthenticated mode.
	Primary auth.Authenticator[U]
}

// Authenticate stops on operational errors and falls back only when no trusted session exists.
func (a Authenticator[U]) Authenticate(r *http.Request) (U, error) {
	var zero U
	switch a.Policy.Mode {
	case auth.ModeLocal, auth.ModeOIDC, auth.ModeProxy, auth.ModeNone:
	default:
		return zero, auth.ErrUnauthenticated
	}
	if a.Policy.LocalEnabled() || a.Policy.Mode == auth.ModeOIDC {
		if a.Session == nil {
			return zero, errors.New("browser session resolver is required")
		}
		user, origin, err := a.Session(r)
		if err == nil && a.Policy.Accepts(origin) {
			return user, nil
		}
		if err != nil && !errors.Is(err, auth.ErrUnauthenticated) {
			return zero, err
		}
	}
	if (a.Policy.Mode == auth.ModeProxy || a.Policy.Mode == auth.ModeNone) && a.Primary != nil {
		return a.Primary.Authenticate(r)
	}
	return zero, auth.ErrUnauthenticated
}
