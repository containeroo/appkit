package proxy

import (
	"fmt"
	"net"
	"net/http"
	"net/netip"
	"strings"
)

// TrustCIDRs checks the immediate TCP peer against nonempty trusted network ranges.
// It never trusts Forwarded or X-Forwarded-For. The ingress must strip and replace identity headers.
func TrustCIDRs(cidrs []string) (TrustFunc, error) {
	if len(cidrs) == 0 {
		return nil, fmt.Errorf("trusted-proxy authentication requires --trusted-proxy-cidrs")
	}
	prefixes := make([]netip.Prefix, 0, len(cidrs))
	for _, cidr := range cidrs {
		prefix, err := netip.ParsePrefix(strings.TrimSpace(cidr))
		if err != nil {
			return nil, fmt.Errorf("invalid trusted proxy CIDR %q: %w", cidr, err)
		}
		prefixes = append(prefixes, prefix.Masked())
	}
	return func(r *http.Request) bool {
		host, _, err := net.SplitHostPort(r.RemoteAddr)
		if err != nil {
			return false
		}
		address, err := netip.ParseAddr(host)
		if err != nil {
			return false
		}
		address = address.Unmap()
		for _, prefix := range prefixes {
			if prefix.Addr().Is4In6() && prefix.Bits() >= 96 {
				prefix = netip.PrefixFrom(prefix.Addr().Unmap(), prefix.Bits()-96)
			}
			if prefix.Contains(address) {
				return true
			}
		}
		return false
	}, nil
}
