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
// Call the returned finalizer after parsing flags to preserve the distinction
// between an absent override and an explicit false value. It does not select an
// application's persisted authentication policy.
func RegisterAuthentication(tf *tinyflags.FlagSet, cfg *Authentication, opts AuthenticationOptions) func() {
	finalizeMode := RegisterAuthMode(tf, cfg)
	finalizeRegistration := RegisterUserRegistration(tf, cfg)
	RegisterLocalLogin(tf, cfg)
	RegisterEncryptionKey(tf, cfg)
	if opts.RegisterSessionLifetime {
		RegisterSessionLifetime(tf, cfg, opts.DefaultSessionLifetime)
	}
	return func() {
		finalizeMode()
		finalizeRegistration()
	}
}

// RegisterAuthMode registers the optional authentication mode override.
// Call the returned finalizer after parsing.
func RegisterAuthMode(tf *tinyflags.FlagSet, cfg *Authentication) func() {
	mode := tinyflags.Enum(tf, "auth-mode", auth.ModeNone,
		"Emergency override for persisted authentication mode",
		auth.ModeNone, auth.ModeLocal, auth.ModeProxy, auth.ModeOIDC).
		Section(SectionAuthentication).
		Placeholder("MODE")

	return func() {
		if mode.Changed() {
			selected := *mode.Value()
			cfg.ModeOverride = &selected
		}
	}
}

// RegisterUserRegistration registers the optional registration override.
// Call the returned finalizer after parsing.
func RegisterUserRegistration(tf *tinyflags.FlagSet, cfg *Authentication) func() {
	registrationDefault := false
	registration := tf.BoolVar(&registrationDefault, "allow-user-registration", false,
		"Deployment override for registration of external identities").
		Section(SectionAuthentication).
		Strict()

	return func() {
		if registration.Changed() {
			selected := *registration.Value()
			cfg.AllowRegistrationOverride = &selected
		}
	}
}

// RegisterLocalLogin registers the local recovery login flag.
func RegisterLocalLogin(tf *tinyflags.FlagSet, cfg *Authentication) {
	tf.BoolVar(&cfg.LocalLogin, "local-login", false,
		"Enable local recovery login alongside the selected authentication mode").
		Section(SectionAuthentication).
		Value()
}

// RegisterEncryptionKey registers the application encryption key flag.
func RegisterEncryptionKey(tf *tinyflags.FlagSet, cfg *Authentication) {
	tf.StringVar(&cfg.EncryptionKey, "encryption-key", "",
		"Base64-encoded 32-byte application encryption key").
		Section(SectionAuthentication).
		OverriddenValueMaskFn(tinyflags.MaskFirstLast).
		Validate(secrets.ValidateKey).
		Value()
}

// RegisterSessionLifetime registers a positive session lifetime with a default.
// A non-positive default selects 12 hours.
func RegisterSessionLifetime(tf *tinyflags.FlagSet, cfg *Authentication, defaultLifetime time.Duration) {
	if defaultLifetime <= 0 {
		defaultLifetime = 12 * time.Hour
	}
	tf.DurationVar(&cfg.SessionLifetime, "session-lifetime", defaultLifetime,
		"Authentication session lifetime").Section(SectionAuthentication).
		Validate(func(value time.Duration) error {
			if value <= 0 {
				return errors.New("session lifetime must be positive")
			}
			return nil
		}).
		Value()
}
