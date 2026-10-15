package proxy

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"reflect"
	"strings"
	"testing"

	"github.com/containeroo/appkit/auth"
)

// TestResolveDefaultHeaders checks that unset persisted header values inherit appkit defaults.
func TestResolveDefaultHeaders(t *testing.T) {
	cidrs := []string{"127.0.0.1/32"}
	cfg, err := Resolve(Options{Overrides: Layer{TrustedCIDRs: &cidrs}})
	if err != nil {
		t.Fatal(err)
	}
	defaults := Defaults()
	if !reflect.DeepEqual(cfg.UsernameHeaders, defaults.UsernameHeaders) || !reflect.DeepEqual(cfg.EmailHeaders, defaults.EmailHeaders) {
		t.Fatalf("defaults not inherited: %#v", cfg)
	}
	if !reflect.DeepEqual(cfg.TrustedCIDRs, cidrs) {
		t.Fatalf("trusted networks missing: %#v", cfg.TrustedCIDRs)
	}
}

// TestResolvePrecedence checks that persisted settings and overrides are applied in order.
func TestResolvePrecedence(t *testing.T) {
	storedUsername := []string{"X-Stored-User"}
	storedEmail := []string{"X-Stored-Email"}
	deploymentUsername := []string{"X-Envoy-User"}
	cidrs := []string{"*"}
	cfg, err := Resolve(Options{
		Persisted: Layer{UsernameHeaders: &storedUsername, EmailHeaders: &storedEmail},
		Overrides: Layer{UsernameHeaders: &deploymentUsername, TrustedCIDRs: &cidrs},
	})
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(cfg.UsernameHeaders, deploymentUsername) {
		t.Fatalf("deployment override lost: %q", cfg.UsernameHeaders)
	}
	if !reflect.DeepEqual(cfg.EmailHeaders, storedEmail) {
		t.Fatalf("persisted setting lost: %q", cfg.EmailHeaders)
	}
	if !reflect.DeepEqual(cfg.GroupHeaders, Defaults().GroupHeaders) {
		t.Fatalf("default group headers lost: %q", cfg.GroupHeaders)
	}
	deploymentUsername[0] = "mutated"
	storedEmail[0] = "mutated"
	if cfg.UsernameHeaders[0] != "X-Envoy-User" || cfg.EmailHeaders[0] != "X-Stored-Email" {
		t.Fatal("resolved configuration aliases its input slices")
	}
}

// TestResolveExplicitlyEmptyHeaders ensures an empty layer is different from an absent layer.
func TestResolveExplicitlyEmptyHeaders(t *testing.T) {
	empty := []string{}
	cidrs := []string{"*"}
	cfg, err := Resolve(Options{Overrides: Layer{
		TrustedCIDRs: &cidrs, UsernameHeaders: &empty,
	}})
	if err != nil {
		t.Fatal(err)
	}
	if len(cfg.UsernameHeaders) != 0 || len(cfg.EmailHeaders) == 0 {
		t.Fatalf("explicitly empty username list not preserved: %#v", cfg)
	}
	if _, err := Resolve(Options{Overrides: Layer{
		TrustedCIDRs: &cidrs, UsernameHeaders: &empty, EmailHeaders: &empty,
	}}); err == nil || !strings.Contains(err.Error(), "username or email") {
		t.Fatalf("empty identity sources accepted: %v", err)
	}
}

// TestResolveFailsClosed checks the trusted-peer allowlist is never inferred from defaults.
func TestResolveFailsClosed(t *testing.T) {
	if _, err := Resolve(Options{}); err == nil || !strings.Contains(err.Error(), "trusted-proxy-cidrs") {
		t.Fatalf("missing peer allowlist accepted: %v", err)
	}
	invalid := []string{"*", "127.0.0.1/32"}
	if _, err := Resolve(Options{Overrides: Layer{TrustedCIDRs: &invalid}}); err == nil {
		t.Fatal("mixed wildcard trusted-peer configuration accepted")
	}
	empty := []string{}
	stored := []string{"127.0.0.1/32"}
	if _, err := Resolve(Options{Persisted: Layer{TrustedCIDRs: &stored}, Overrides: Layer{TrustedCIDRs: &empty}}); err == nil {
		t.Fatal("explicitly cleared peer allowlist accepted")
	}
}

// TestResolveRejectsBadHeaders ensures invalid HTTP field names are rejected at startup.
func TestResolveRejectsBadHeaders(t *testing.T) {
	cidrs := []string{"*"}
	for _, bad := range []string{"", "X Auth", "X\r\nInjected", "X:Name", " X-Foo"} {
		badHeaders := []string{bad}
		_, err := Resolve(Options{Overrides: Layer{TrustedCIDRs: &cidrs, EmailHeaders: &badHeaders}})
		if err == nil || !strings.Contains(err.Error(), "invalid trusted proxy email header") {
			t.Errorf("bad header %q accepted: %v", bad, err)
		}
	}
}

// TestProxyNewWithEmailOnlyEnvoy checks the entire resolver-to-authenticator path.
func TestProxyNewWithEmailOnlyEnvoy(t *testing.T) {
	cidrs := []string{"10.0.0.0/8"}
	empty := []string{}
	email := []string{"X-Auth-Request-Email"}
	cfg, err := Resolve(Options{Overrides: Layer{
		TrustedCIDRs: &cidrs, UsernameHeaders: &empty, EmailHeaders: &email,
	}})
	if err != nil {
		t.Fatal(err)
	}
	provider, err := New(cfg, "envoy", func(_ context.Context, identity auth.ExternalIdentity) (string, error) {
		if identity.Username != "alice@example.test" || identity.Subject != "alice@example.test" || identity.Email != "alice@example.test" {
			t.Errorf("unexpected email-only identity: %#v", identity)
		}
		return identity.Username, nil
	})
	if err != nil {
		t.Fatal(err)
	}
	r := httptest.NewRequest(http.MethodGet, "/", nil)
	r.RemoteAddr = "10.0.2.4:40000"
	r.Header.Set("X-Auth-Request-Email", "alice@example.test")
	user, err := provider.Authenticate(r)
	if err != nil || user != "alice@example.test" {
		t.Fatalf("email-only authentication: %q, %v", user, err)
	}
	r.RemoteAddr = "192.0.2.10:40000"
	if _, err := provider.Authenticate(r); !errors.Is(err, auth.ErrUnauthenticated) {
		t.Fatalf("untrusted peer accepted: %v", err)
	}
}

// TestDefaultsIndependent ensures callers cannot mutate another application's default header lists.
func TestDefaultsIndependent(t *testing.T) {
	first := Defaults()
	first.EmailHeaders[1] = "X-Tampered"
	second := Defaults()
	if second.EmailHeaders[1] != "X-Auth-Request-Email" {
		t.Fatalf("default list was mutated: %#v", second)
	}
}

// TestNewRejectsMissingIdentityResolver checks the application resolver and issuer stay mandatory.
func TestNewRejectsMissingIdentityResolver(t *testing.T) {
	cidrs := []string{"*"}
	cfg, err := Resolve(Options{Overrides: Layer{TrustedCIDRs: &cidrs}})
	if err != nil {
		t.Fatal(err)
	}
	if _, err := New[string](cfg, "envoy", nil); err == nil {
		t.Fatal("nil user resolver accepted")
	}
	if _, err := New(cfg, "", func(context.Context, auth.ExternalIdentity) (string, error) { return "", nil }); err == nil {
		t.Fatal("empty issuer accepted")
	}
}
