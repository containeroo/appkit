package provisioning

import (
	"context"
	"errors"
	"github.com/containeroo/appkit/auth"
	"testing"
)

func TestAdministratorPolicy(t *testing.T) {
	id := Identity{Issuer: "issuer", Subject: " subject "}
	accounts := Accounts[string]{Resolve: func(_ context.Context, key Identity, allow bool) (string, error) {
		if allow || key.Subject != id.Subject {
			t.Fatal("unexpected provisioning or subject change")
		}
		return "member", nil
	}, IsAdministrator: func(string) bool { return false }}
	_, err := accounts.Login(t.Context(), AccessPolicy{RequireAdministrator: true}, id)
	if !errors.Is(err, auth.ErrForbidden) {
		t.Fatal(err)
	}
}
