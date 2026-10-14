// Package provisioning coordinates external identities with application-owned user storage.
package provisioning

import (
	"context"
	"errors"
	"strings"
)

// ErrIncompleteIdentity indicates that a provider omitted an identity required for provisioning.
var ErrIncompleteIdentity = errors.New("external identity requires issuer, subject, and username")

// Identity is a verified external identity supplied by an authentication boundary.
type Identity struct {
	Issuer      string
	Subject     string
	Username    string
	Email       string
	DisplayName string
}

// Normalize trims identity fields and rejects identities that cannot be linked safely.
func Normalize(identity Identity) (Identity, error) {
	identity.Issuer = strings.TrimSpace(identity.Issuer)
	identity.Subject = strings.TrimSpace(identity.Subject)
	identity.Username = strings.TrimSpace(identity.Username)
	identity.Email = strings.TrimSpace(identity.Email)
	identity.DisplayName = strings.TrimSpace(identity.DisplayName)
	if identity.Issuer == "" || identity.Subject == "" || identity.Username == "" {
		return Identity{}, ErrIncompleteIdentity
	}
	return identity, nil
}

// RegistrationPolicy resolves a deployment override before persisted registration settings.
type RegistrationPolicy struct {
	Override *bool
	Load     func(context.Context) (bool, error)
}

// Enabled reports whether unknown external identities may create users.
func (p RegistrationPolicy) Enabled(ctx context.Context) (bool, error) {
	if p.Override != nil {
		return *p.Override, nil
	}
	if p.Load == nil {
		return false, nil
	}
	return p.Load(ctx)
}

// Resolver coordinates identity lookup and atomic user provisioning.
type Resolver[User any] struct {
	Registration RegistrationPolicy
	IsMissing    func(error) bool
	Lookup       func(context.Context, Identity) (User, error)
	Resolve      func(context.Context, Identity, bool) (User, error)
}

// Login normalizes identity data and only enables registration when no identity exists.
func (r Resolver[User]) Login(ctx context.Context, identity Identity) (User, error) {
	var zero User
	identity, err := Normalize(identity)
	if err != nil {
		return zero, err
	}
	if r.Lookup == nil || r.Resolve == nil {
		return zero, errors.New("external identity resolver is not configured")
	}
	registrationEnabled := false
	if _, err := r.Lookup(ctx, identity); err != nil {
		if r.IsMissing == nil || !r.IsMissing(err) {
			return zero, err
		}
		registrationEnabled, err = r.Registration.Enabled(ctx)
		if err != nil {
			return zero, err
		}
	}
	return r.Resolve(ctx, identity, registrationEnabled)
}
