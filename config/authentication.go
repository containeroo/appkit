package config

import (
	"errors"
	"time"

	"github.com/containeroo/appkit/auth"
	"github.com/containeroo/appkit/secrets"
	"github.com/containeroo/tinyflags"
)

// Authentication holds deployment-level recovery flags, not persisted auth policy.
type Authentication struct {
	// ModeOverride overrides database-managed auth policy only when non-nil.
	ModeOverride *auth.Mode
	// AllowRegistrationOverride overrides registration policy only when non-nil.
	AllowRegistrationOverride *bool
	// LocalLogin allows an application-managed recovery login alongside its selected mode.
	LocalLogin bool
	// EncryptionKey is a deployment key for application secrets.
	EncryptionKey string
	// SessionLifetime is the optional session duration selected by the application.
	SessionLifetime time.Duration
}

// AuthenticationOptions enables selected optional flags without imposing session policy.
type AuthenticationOptions struct {
	// RegisterSessionLifetime adds --session-lifetime only when enabled.
	RegisterSessionLifetime bool
	// DefaultSessionLifetime defaults to 12 hours if no value is supplied.
	DefaultSessionLifetime time.Duration
}

// RegisterAuthentication registers common authentication override flags.
// The nil-versus-explicit-false distinction is preserved for registration.
func RegisterAuthentication(tf *tinyflags.FlagSet, cfg *Authentication, opts AuthenticationOptions) func() {
	mode := tinyflags.Enum(tf, "auth-mode", auth.ModeNone,
		"Emergency override for persisted authentication mode",
		auth.ModeNone, auth.ModeLocal, auth.ModeProxy, auth.ModeTrustedProxy, auth.ModeOIDC).
		Section(SectionAuthentication).
		Placeholder("MODE")

	registrationDefault := false
	registration := tf.BoolVar(&registrationDefault, "allow-user-registration", false,
		"Deployment override for registration of external identities").
		Section(SectionAuthentication).
		Strict()

	tf.BoolVar(&cfg.LocalLogin, "local-login", false,
		"Enable local recovery login alongside the selected authentication mode").
		Section(SectionAuthentication).
		Value()

	tf.StringVar(&cfg.EncryptionKey, "encryption-key", "",
		"Base64-encoded 32-byte application encryption key").
		Section(SectionAuthentication).
		OverriddenValueMaskFn(tinyflags.MaskFirstLast).
		Validate(secrets.ValidateKey).
		Value()

	if opts.RegisterSessionLifetime {
		lifetime := opts.DefaultSessionLifetime
		if lifetime <= 0 {
			lifetime = 12 * time.Hour
		}
		tf.DurationVar(&cfg.SessionLifetime, "session-lifetime", lifetime,
			"Authentication session lifetime").Section(SectionAuthentication).
			Validate(func(value time.Duration) error {
				if value <= 0 {
					return errors.New("session lifetime must be positive")
				}
				return nil
			}).
			Value()
	}
	return func() {
		if mode.Changed() {
			selected := *mode.Value()
			cfg.ModeOverride = &selected
		}
		if registration.Changed() {
			selected := *registration.Value()
			cfg.AllowRegistrationOverride = &selected
		}
	}
}
