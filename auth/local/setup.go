package local

import (
	"context"
	"crypto/sha256"
	"crypto/subtle"
	"errors"

	"github.com/containeroo/appkit/auth"
)

// Setup provisions a deployment's first administrator.
type Setup[U any] struct {
	// TokenDigest requires a matching non-empty bootstrap token when configured.
	TokenDigest *[sha256.Size]byte
	// Validate enforces the application's username and password policy.
	Validate func(string, string) error
	// Create atomically creates the first administrator and any deployment-configured identity link.
	Create func(context.Context, string, string) (U, error)
}

// Bootstrap authorizes setup and hashes the validated password before calling the store.
func (s Setup[U]) Bootstrap(ctx context.Context, token, name, password string) (U, error) {
	var zero U
	if s.TokenDigest != nil {
		candidate := sha256.Sum256([]byte(token))
		if token == "" || subtle.ConstantTimeCompare(candidate[:], s.TokenDigest[:]) != 1 {
			return zero, auth.ErrForbidden
		}
	}
	if s.Validate == nil || s.Create == nil {
		return zero, errors.New("incomplete administrator setup adapter")
	}
	if err := s.Validate(name, password); err != nil {
		return zero, err
	}
	hash, err := HashPassword(password)
	if err != nil {
		return zero, err
	}
	return s.Create(ctx, name, hash)
}
