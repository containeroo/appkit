package provisioning

import (
	"context"
	"errors"
	"strings"

	"github.com/containeroo/appkit/auth"
)

// AccessPolicy selects external registration and required local privileges.
type AccessPolicy struct {
	// AllowRegistration permits atomic creation of an account for an unknown identity.
	AllowRegistration bool
	// RequireAdministrator restricts access to existing local administrator privileges.
	RequireAdministrator bool
}

// Accounts adapts an application's enabled-account store to a resolved access policy.
type Accounts[U any] struct {
	// Resolve must look up the verified issuer and subject, exclude disabled users,
	// and only create a new account and its identity link when registration is explicitly allowed.
	Resolve func(context.Context, Identity, bool) (U, error)
	// IsAdministrator reads the resolved local user's privileges.
	IsAdministrator func(U) bool
}

// Login resolves the external identity and enforces the required local role.
func (a Accounts[U]) Login(ctx context.Context, p AccessPolicy, id Identity) (U, error) {
	var zero U
	// Issuers and subjects are opaque identifiers; preserve their exact values.
	if strings.TrimSpace(id.Issuer) == "" || strings.TrimSpace(id.Subject) == "" {
		return zero, ErrIncompleteIdentity
	}
	if a.Resolve == nil || (p.RequireAdministrator && a.IsAdministrator == nil) {
		return zero, errors.New("incomplete external account adapter")
	}
	user, err := a.Resolve(ctx, id, p.AllowRegistration)
	if err != nil {
		return zero, err
	}
	if p.RequireAdministrator && !a.IsAdministrator(user) {
		return zero, auth.ErrForbidden
	}
	return user, nil
}
