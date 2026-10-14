package proxy

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/containeroo/appkit/auth"
)

func TestProxyTrustBoundary(t *testing.T) {
	trusted := false
	provider := Authenticator[string]{
		Issuer: "internal-proxy", UsernameHeaders: []string{"X-Forwarded-User"},
		GroupHeaders: []string{"X-Forwarded-Groups"},
		Trust:        func(*http.Request) bool { return trusted },
		Resolve: func(_ context.Context, identity auth.ExternalIdentity) (string, error) {
			if identity.Subject != "alice" || len(identity.Groups) != 2 {
				t.Errorf("unexpected identity: %#v", identity)
			}
			return identity.Username, nil
		},
	}
	request := httptest.NewRequest("GET", "/", nil)
	request.Header.Set("X-Forwarded-User", "alice")
	request.Header.Set("X-Forwarded-Groups", "staff, operations")
	if _, err := provider.Authenticate(request); !errors.Is(err, auth.ErrUnauthenticated) {
		t.Fatalf("untrusted request accepted: %v", err)
	}
	trusted = true
	user, err := provider.Authenticate(request)
	if err != nil || user != "alice" {
		t.Fatalf("valid identity: %q %v", user, err)
	}
}
func TestProxyRequiresExplicitTrust(t *testing.T) {
	provider := Authenticator[string]{Issuer: "proxy", UsernameHeaders: []string{"X-Forwarded-User"}}
	if err := provider.Validate(); err == nil {
		t.Fatal("missing trust boundary accepted")
	}
}
