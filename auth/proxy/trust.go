package proxy

import (
	"errors"
	"fmt"
	"net"
	"net/http"
	"net/netip"
	"strings"
)

// IsWildcardCIDRs reports whether all immediate TCP peers are explicitly trusted.
// The wildcard must be the sole list entry; other combinations are invalid.
func IsWildcardCIDRs(cidrs []string) bool {
	return len(cidrs) == 1 && strings.TrimSpace(cidrs[0]) == "*"
}

// TrustCIDRs checks the immediate TCP peer against configured networks, or all
// valid peers when the sole entry is "*". It never trusts forwarded IP headers.
// The ingress must strip and replace any client-supplied identity headers.
func TrustCIDRs(cidrs []string) (TrustFunc, error) {
	if len(cidrs) == 0 {
		return nil, errors.New("trusted-proxy authentication requires --trusted-proxy-cidrs")
	}
	if IsWildcardCIDRs(cidrs) {
		return func(r *http.Request) bool {
			_, ok := immediatePeer(r)
			return ok
		}, nil
	}

	prefixes := make([]netip.Prefix, 0, len(cidrs))
	for _, cidr := range cidrs {
		value := strings.TrimSpace(cidr)
		if value == "*" {
			return nil, errors.New("trusted proxy wildcard * must be specified alone")
		}
		prefix, err := netip.ParsePrefix(value)
		if err != nil {
			return nil, fmt.Errorf("invalid trusted proxy CIDR %q: %w", cidr, err)
		}
		prefixes = append(prefixes, prefix.Masked())
	}
	return func(r *http.Request) bool {
		address, ok := immediatePeer(r)
		if !ok {
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

// immediatePeer parses the real transport peer, not a forwarded header.
func immediatePeer(r *http.Request) (netip.Addr, bool) {
	if r == nil {
		return netip.Addr{}, false
	}
	host, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		return netip.Addr{}, false
	}
	address, err := netip.ParseAddr(host)
	if err != nil {
		return netip.Addr{}, false
	}
	return address, true
}
