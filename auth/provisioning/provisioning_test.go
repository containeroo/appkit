package provisioning

import (
	"context"
	"errors"
	"testing"
)

func TestNormalize(t *testing.T) {
	identity, err := Normalize(Identity{Issuer: " issuer ", Subject: " subject ", Username: " alice ", Email: " alice@example.test ", DisplayName: " Alice "})
	if err != nil {
		t.Fatal(err)
	}
	if identity.Issuer != "issuer" || identity.Subject != "subject" || identity.Username != "alice" || identity.Email != "alice@example.test" || identity.DisplayName != "Alice" {
		t.Fatalf("identity: %#v", identity)
	}
}

func TestNormalizeRejectsIncompleteIdentity(t *testing.T) {
	_, err := Normalize(Identity{Issuer: "issuer", Subject: "subject"})
	if !errors.Is(err, ErrIncompleteIdentity) {
		t.Fatalf("error: %v", err)
	}
}

func TestResolverEnablesRegistrationForMissingIdentity(t *testing.T) {
	resolver := Resolver[string]{
		Registration: RegistrationPolicy{Load: func(context.Context) (bool, error) { return true, nil }},
		IsMissing:    func(error) bool { return true },
		Lookup:       func(context.Context, Identity) (string, error) { return "", errors.New("missing") },
		Resolve: func(_ context.Context, identity Identity, enabled bool) (string, error) {
			if !enabled {
				t.Fatal("registration disabled")
			}
			return identity.Username, nil
		},
	}
	user, err := resolver.Login(context.Background(), Identity{Issuer: "issuer", Subject: "subject", Username: "alice"})
	if err != nil || user != "alice" {
		t.Fatalf("user=%q err=%v", user, err)
	}
}

func TestResolverSkipsRegistrationForExistingIdentity(t *testing.T) {
	resolver := Resolver[string]{
		Registration: RegistrationPolicy{Load: func(context.Context) (bool, error) { t.Fatal("policy lookup"); return false, nil }},
		IsMissing:    func(error) bool { return false },
		Lookup:       func(context.Context, Identity) (string, error) { return "existing", nil },
		Resolve: func(_ context.Context, _ Identity, enabled bool) (string, error) {
			if enabled {
				t.Fatal("registration enabled")
			}
			return "existing", nil
		},
	}
	user, err := resolver.Login(context.Background(), Identity{Issuer: "issuer", Subject: "subject", Username: "alice"})
	if err != nil || user != "existing" {
		t.Fatalf("user=%q err=%v", user, err)
	}
}
