package proxy

import (
	"errors"
	"fmt"
	"strings"
)

// Config contains the effective identity headers and ingress trust boundary.
// Obtain it through Resolve rather than replicating framework defaults in an application.
type Config struct {
	// TrustedCIDRs contains immediate-peer networks, or a sole * for explicitly unrestricted peers.
	TrustedCIDRs []string
	// UsernameHeaders are checked in order for an asserted username.
	UsernameHeaders []string
	// EmailHeaders are checked in order for an asserted email address.
	EmailHeaders []string
	// DisplayNameHeaders are checked in order for an optional display name.
	DisplayNameHeaders []string
	// GroupHeaders are checked in order for optional group membership.
	GroupHeaders []string
}

// Layer describes explicitly present settings; nil means inherit, while a pointer
// to an empty slice intentionally clears an inherited setting.
type Layer struct {
	// TrustedCIDRs overrides the immediate-peer allowlist when non-nil.
	TrustedCIDRs *[]string
	// UsernameHeaders overrides username header priority when non-nil.
	UsernameHeaders *[]string
	// EmailHeaders overrides email header priority when non-nil.
	EmailHeaders *[]string
	// DisplayNameHeaders overrides display-name header priority when non-nil.
	DisplayNameHeaders *[]string
	// GroupHeaders overrides group header priority when non-nil.
	GroupHeaders *[]string
}

// Options defines persisted settings and higher-priority deployment overrides.
type Options struct {
	// Persisted contains application-managed settings, distinct from absent values.
	Persisted Layer
	// Overrides contains explicit CLI or environment settings and takes precedence.
	Overrides Layer
}

// Defaults returns independent copies of the standard identity header lists.
// No network peers are trusted by default.
func Defaults() Config {
	return Config{
		UsernameHeaders:    []string{"X-Forwarded-User", "X-Auth-Request-User", "Remote-User"},
		EmailHeaders:       []string{"X-Forwarded-Email", "X-Auth-Request-Email", "X-Authentik-Email"},
		DisplayNameHeaders: []string{"X-Forwarded-Name", "X-Auth-Request-Preferred-Username", "X-Authentik-Name"},
		GroupHeaders:       []string{"X-Forwarded-Groups", "X-Auth-Request-Groups"},
	}
}

// Resolve applies defaults, persisted settings and explicit deployment overrides
// in that order, then validates the effective proxy authentication configuration.
func Resolve(options Options) (Config, error) {
	cfg := Defaults()
	cfg.apply(options.Persisted)
	cfg.apply(options.Overrides)
	if err := cfg.Validate(); err != nil {
		return Config{}, err
	}
	return cfg, nil
}

// Validate checks trusted-peer networks and HTTP identity header configuration.
func (cfg Config) Validate() error {
	if _, err := TrustCIDRs(cfg.TrustedCIDRs); err != nil {
		return err
	}
	if len(cfg.UsernameHeaders) == 0 && len(cfg.EmailHeaders) == 0 {
		return errors.New("trusted-proxy authentication requires a username or email header")
	}
	for _, group := range []struct {
		name    string
		headers []string
	}{
		{"username", cfg.UsernameHeaders},
		{"email", cfg.EmailHeaders},
		{"display name", cfg.DisplayNameHeaders},
		{"group", cfg.GroupHeaders},
	} {
		for _, name := range group.headers {
			if !validHeaderName(name) {
				return fmt.Errorf("invalid trusted proxy %s header %q", group.name, name)
			}
		}
	}
	return nil
}

// New builds an authenticator from a resolved configuration and an application-owned resolver.
// It does not choose an issuer or user provisioning policy on the application's behalf.
func New[User any](cfg Config, issuer string, resolve Resolver[User]) (Authenticator[User], error) {
	if err := cfg.Validate(); err != nil {
		return Authenticator[User]{}, err
	}
	trust, err := TrustCIDRs(cfg.TrustedCIDRs)
	if err != nil {
		return Authenticator[User]{}, err
	}
	provider := Authenticator[User]{
		Issuer:             issuer,
		Resolve:            resolve,
		Trust:              trust,
		UsernameHeaders:    append([]string(nil), cfg.UsernameHeaders...),
		EmailHeaders:       append([]string(nil), cfg.EmailHeaders...),
		DisplayNameHeaders: append([]string(nil), cfg.DisplayNameHeaders...),
		GroupHeaders:       append([]string(nil), cfg.GroupHeaders...),
	}
	if err := provider.Validate(); err != nil {
		return Authenticator[User]{}, err
	}
	return provider, nil
}

// apply replaces only fields explicitly present in the supplied layer.
func (cfg *Config) apply(layer Layer) {
	if layer.TrustedCIDRs != nil {
		cfg.TrustedCIDRs = append([]string{}, (*layer.TrustedCIDRs)...)
	}
	if layer.UsernameHeaders != nil {
		cfg.UsernameHeaders = append([]string{}, (*layer.UsernameHeaders)...)
	}
	if layer.EmailHeaders != nil {
		cfg.EmailHeaders = append([]string{}, (*layer.EmailHeaders)...)
	}
	if layer.DisplayNameHeaders != nil {
		cfg.DisplayNameHeaders = append([]string{}, (*layer.DisplayNameHeaders)...)
	}
	if layer.GroupHeaders != nil {
		cfg.GroupHeaders = append([]string{}, (*layer.GroupHeaders)...)
	}
}

// validHeaderName rejects whitespace, separators, and control characters in header names.
func validHeaderName(name string) bool {
	if name == "" || strings.TrimSpace(name) != name {
		return false
	}
	for i := 0; i < len(name); i++ {
		c := name[i]
		if (c >= 'A' && c <= 'Z') || (c >= 'a' && c <= 'z') || (c >= '0' && c <= '9') {
			continue
		}
		if !strings.ContainsRune("!#$%&'*+-.^_`|~", rune(c)) {
			return false
		}
	}
	return true
}
