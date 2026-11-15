package browser

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/containeroo/appkit/auth"
)

func TestBrowserPolicy(t *testing.T) {
	t.Run("changed issuer requires another sign-in", func(t *testing.T) {
		t.Parallel()
		p := Policy{Mode: auth.ModeOIDC, Issuer: "new", Recovery: true}
		if p.Accepts(Origin{Mode: auth.ModeOIDC, Issuer: "old"}) || !p.Accepts(Origin{Mode: auth.ModeOIDC, Issuer: "new"}) {
			t.Fatal("incorrect issuer boundary")
		}
	})
	t.Run("disabling recovery rejects local sessions", func(t *testing.T) {
		t.Parallel()
		p := Policy{Mode: auth.ModeProxy}
		if p.Accepts(Origin{Mode: auth.ModeLocal}) {
			t.Fatal("accepted local session")
		}
	})
	t.Run("missing recovery cookie falls through to proxy", func(t *testing.T) {
		t.Parallel()
		a := Authenticator[string]{Policy: Policy{Mode: auth.ModeProxy, Recovery: true},
			Session: func(*http.Request) (string, Origin, error) { return "", Origin{}, auth.ErrUnauthenticated },
			Primary: auth.AuthenticatorFunc[string](func(*http.Request) (string, error) { return "proxy account", nil })}
		u, err := a.Authenticate(httptest.NewRequest("GET", "/", nil))
		if err != nil || u != "proxy account" {
			t.Fatalf("got %q, %v", u, err)
		}
	})
	t.Run("session store failure stops proxy fallback", func(t *testing.T) {
		t.Parallel()
		failure := errors.New("database unavailable")
		a := Authenticator[string]{Policy: Policy{Mode: auth.ModeProxy, Recovery: true},
			Session: func(*http.Request) (string, Origin, error) { return "", Origin{}, failure },
			Primary: auth.AuthenticatorFunc[string](func(*http.Request) (string, error) { t.Fatal("unexpected fallback"); return "", nil })}
		_, err := a.Authenticate(httptest.NewRequest("GET", "/", nil))
		if !errors.Is(err, failure) {
			t.Fatal(err)
		}
	})
}
