package config

import (
	"github.com/containeroo/appkit/auth/proxy"
	"github.com/containeroo/tinyflags"
)

// Proxy contains parsed deployment flags; DeploymentOverrides identifies explicit values.
type Proxy struct {
	// TrustedCIDRs lists allowed immediate peers; a sole "*" explicitly trusts all peers.
	TrustedCIDRs []string
	// UsernameHeaders are the incoming username header names, in priority order.
	UsernameHeaders []string
	// EmailHeaders are the incoming email header names, in priority order.
	EmailHeaders []string
	// DisplayNameHeaders are the incoming display-name header names.
	DisplayNameHeaders []string
	// GroupHeaders are the group membership header names.
	GroupHeaders []string
	// AdminGroup is an optional application-specific group mapping input.
	AdminGroup string
}

// OIDC holds deployment-managed OIDC secrets and optional runtime overrides.
type OIDC struct {
	// Issuer is the optional issuer override.
	Issuer string
	// ClientID is the optional client ID override.
	ClientID string
	// ClientSecret is the OIDC client secret.
	ClientSecret string
	// SessionSecret protects temporary login state and may be reused by application-managed sessions.
	SessionSecret string
	// GroupClaim is the optional external groups claim.
	GroupClaim string
	// AdminGroup is an optional external administrator group mapping input.
	AdminGroup string
}

// RegisterProxy registers trusted-proxy identity header overrides.
func RegisterProxy(tf *tinyflags.FlagSet, cfg *Proxy) {
	defaults := proxy.Defaults()
	tf.StringSliceVar(&cfg.TrustedCIDRs, "trusted-proxy-cidrs", nil,
		"Trusted ingress peer CIDRs, or * to explicitly trust every peer (unsafe without ingress isolation)").
		Placeholder("CIDR").
		Section(SectionProxy).
		Value()

	tf.StringSliceVar(&cfg.UsernameHeaders, "trusted-username-headers",
		defaults.UsernameHeaders,
		"Trusted-proxy username headers").
		Placeholder("USERNAME-HEADER").
		Section(SectionProxy).
		Value()

	tf.StringSliceVar(&cfg.EmailHeaders, "trusted-email-headers",
		defaults.EmailHeaders,
		"Trusted-proxy email headers").
		Placeholder("EMAIL-HEADER").
		AllowEmpty().
		Section(SectionProxy).
		Value()
	tf.StringSliceVar(&cfg.DisplayNameHeaders, "trusted-display-name-headers",
		defaults.DisplayNameHeaders,
		"Trusted-proxy display name headers").
		Placeholder("DISPLAY-NAME-HEADER").
		Section(SectionProxy).
		AllowEmpty().
		Value()
	tf.StringSliceVar(&cfg.GroupHeaders, "trusted-group-headers",
		defaults.GroupHeaders,
		"Trusted-proxy group headers").
		Placeholder("GROUP-HEADER").
		AllowEmpty().
		Section(SectionProxy).
		Value()
	tf.StringVar(&cfg.AdminGroup, "trusted-admin-group", "", "External group granting administration access").
		Placeholder("GROUP").
		Section(SectionProxy).
		Value()
}

// DeploymentOverrides returns explicitly selected flag or environment settings.
// Values omitted from overrides are inherited from persisted settings and appkit defaults.
// A present key with an empty list intentionally clears that header list.
func (cfg Proxy) DeploymentOverrides(overrides tinyflags.Overrides) proxy.Layer {
	values := overrides.Values()
	var layer proxy.Layer
	if _, ok := values["trusted-proxy-cidrs"]; ok {
		layer.TrustedCIDRs = copyHeaders(cfg.TrustedCIDRs)
	}
	if _, ok := values["trusted-username-headers"]; ok {
		layer.UsernameHeaders = copyHeaders(cfg.UsernameHeaders)
	}
	if _, ok := values["trusted-email-headers"]; ok {
		layer.EmailHeaders = copyHeaders(cfg.EmailHeaders)
	}
	if _, ok := values["trusted-display-name-headers"]; ok {
		layer.DisplayNameHeaders = copyHeaders(cfg.DisplayNameHeaders)
	}
	if _, ok := values["trusted-group-headers"]; ok {
		layer.GroupHeaders = copyHeaders(cfg.GroupHeaders)
	}
	return layer
}

// copyHeaders returns a distinct slice pointer, including for explicitly empty lists.
func copyHeaders(headers []string) *[]string {
	cloned := append([]string{}, headers...)
	return &cloned
}

// RegisterOIDC registers OIDC deployment settings with secret masking.
func RegisterOIDC(tf *tinyflags.FlagSet, cfg *OIDC) {
	tf.StringVar(&cfg.Issuer, "oidc-issuer", "", "OpenID Connect issuer override").
		Section(SectionOIDC).
		Placeholder("URL").
		Value()

	tf.StringVar(&cfg.ClientID, "oidc-client-id", "", "OpenID Connect client ID override").
		Section(SectionOIDC).
		Value()

	tf.StringVar(&cfg.ClientSecret, "oidc-client-secret", "", "OpenID Connect client secret").
		Section(SectionOIDC).
		OverriddenValueMaskFn(tinyflags.MaskFirstLast).
		Value()

	tf.StringVar(&cfg.SessionSecret, "oidc-session-secret", "", "Secret for OIDC transient login state").
		Section(SectionOIDC).
		OverriddenValueMaskFn(tinyflags.MaskFirstLast).
		Validate(tinyflags.
			Optional(tinyflags.MinLength(32))).
		Value()

	tf.StringVar(&cfg.GroupClaim, "oidc-group-claim", "groups", "External group-membership claim name").
		Section(SectionOIDC).
		Value()

	tf.StringVar(&cfg.AdminGroup, "oidc-admin-group", "", "External group granting administration access").
		Section(SectionOIDC).
		Value()
}
