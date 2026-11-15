package local

import (
	"context"
	"crypto/sha256"
	"errors"
	"github.com/containeroo/appkit/auth"
	"testing"
)

func TestAuthorizedSetup(t *testing.T) {
	digest := sha256.Sum256([]byte("bootstrap"))
	called := false
	setup := Setup[string]{TokenDigest: &digest, Validate: func(string, string) error { return nil }, Create: func(_ context.Context, name, hash string) (string, error) {
		called = true
		return name, VerifyPassword(hash, "password")
	}}
	if _, err := setup.Bootstrap(t.Context(), "wrong", "admin", "password"); !errors.Is(err, auth.ErrForbidden) || called {
		t.Fatal("unauthorized setup", err)
	}
	if u, err := setup.Bootstrap(t.Context(), "bootstrap", "admin", "password"); err != nil || u != "admin" || !called {
		t.Fatal("setup failed", err)
	}
}
