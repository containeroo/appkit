package proxy

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestTrustCIDRs(t *testing.T) {
	if _, err := TrustCIDRs(nil); err == nil {
		t.Fatal("empty allowlist accepted")
	}
	if _, err := TrustCIDRs([]string{"invalid"}); err == nil {
		t.Fatal("bad CIDR accepted")
	}
	trust, err := TrustCIDRs([]string{"127.0.0.1/32", "2001:db8::/32"})
	if err != nil {
		t.Fatal(err)
	}
	for _, tc := range []struct {
		remote  string
		allowed bool
	}{
		{"127.0.0.1:1234", true}, {"[::ffff:127.0.0.1]:1234", true},
		{"[2001:db8::1]:1234", true}, {"192.0.2.1:1234", false},
		{"127.0.0.1", false}, {"bad:address", false},
	} {
		request := httptest.NewRequest(http.MethodGet, "/", nil)
		request.RemoteAddr = tc.remote
		request.Header.Set("X-Forwarded-For", "127.0.0.1")
		if actual := trust(request); actual != tc.allowed {
			t.Errorf("peer %q allowed=%t, want=%t", tc.remote, actual, tc.allowed)
		}
	}
}

func TestTrustCIDRsWildcard(t *testing.T) {
	for _, cidrs := range [][]string{{"*"}, {" * "}} {
		if !IsWildcardCIDRs(cidrs) {
			t.Fatalf("valid wildcard not recognized: %q", cidrs)
		}
		trust, err := TrustCIDRs(cidrs)
		if err != nil {
			t.Fatal(err)
		}
		for _, remote := range []string{"127.0.0.1:12", "198.51.100.4:65535", "[2001:db8::1]:123", "[::ffff:127.0.0.1]:90"} {
			request := httptest.NewRequest(http.MethodGet, "/", nil)
			request.RemoteAddr = remote
			if !trust(request) {
				t.Errorf("wildcard rejected valid peer %q", remote)
			}
		}
		for _, remote := range []string{"", "127.0.0.1", "bad:address", "not-an-ip:80"} {
			request := httptest.NewRequest(http.MethodGet, "/", nil)
			request.RemoteAddr = remote
			request.Header.Set("X-Forwarded-For", "127.0.0.1")
			if trust(request) {
				t.Errorf("wildcard accepted malformed peer %q", remote)
			}
		}
		if trust(nil) {
			t.Error("wildcard accepted nil request")
		}
	}
}

func TestTrustCIDRsRejectsMixedWildcard(t *testing.T) {
	for _, cidrs := range [][]string{
		{"*", "192.0.2.0/24"},
		{"192.0.2.0/24", "*"},
		{"*", "*"},
		{"  *  ", "2001:db8::/32"},
	} {
		if IsWildcardCIDRs(cidrs) {
			t.Fatalf("mixed wildcard considered valid: %q", cidrs)
		}
		if _, err := TrustCIDRs(cidrs); err == nil || !strings.Contains(err.Error(), "alone") {
			t.Errorf("mixed wildcard %q: expected useful error, got %v", cidrs, err)
		}
	}
	if IsWildcardCIDRs(nil) || IsWildcardCIDRs([]string{"0.0.0.0/0"}) {
		t.Error("non-wildcard list misidentified")
	}
}
