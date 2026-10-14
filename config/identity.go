package config

import "github.com/containeroo/tinyflags"

// Proxy holds optional deployment header overrides for a trusted identity proxy.
type Proxy struct {
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

var trustedUsernameHeaders = []string{
	"X-Forwarded-User",
	"X-Auth-Request-User",
	"Remote-User",
}

var trustedEmailHeaders = []string{
	"X-Forwarded-Email",
	"X-Auth-Request-Email",
	"X-Authentik-Email",
}

var trustedDisplayNameHeaders = []string{
	"X-Forwarded-Name",
	"X-Auth-Request-Preferred-Username",
	"X-Authentik-Name",
}

var trustedGroupHeaders = []string{
	"X-Forwarded-Groups",
	"X-Auth-Request-Groups",
}

// RegisterProxy registers trusted-proxy identity header overrides.
func RegisterProxy(tf *tinyflags.FlagSet, cfg *Proxy) {
	tf.StringSliceVar(&cfg.UsernameHeaders, "trusted-username-headers",
		trustedUsernameHeaders,
		"Trusted-proxy username headers").
		Section(SectionProxy).
		Value()

	tf.StringSliceVar(&cfg.EmailHeaders, "trusted-email-headers",
		trustedEmailHeaders,
		"Trusted-proxy email headers").
		Section(SectionProxy).
		Value()
	tf.StringSliceVar(&cfg.DisplayNameHeaders, "trusted-display-name-headers",
		trustedDisplayNameHeaders,
		"Trusted-proxy display name headers").
		Section(SectionProxy).
		Value()
	tf.StringSliceVar(&cfg.GroupHeaders, "trusted-group-headers",
		trustedGroupHeaders,
		"Trusted-proxy group headers").
		Section(SectionProxy).
		Value()
	tf.StringVar(&cfg.AdminGroup, "trusted-admin-group", "", "External group granting administration access").
		Section(SectionProxy).
		Value()
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
