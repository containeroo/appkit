package proxy

import (
	"net/http"
	"net/http/httptest"
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
