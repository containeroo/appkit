package config

import (
	"testing"

	"github.com/containeroo/appkit/auth"
)

func TestAuthenticationModeOnlyAcceptsCanonicalValues(t *testing.T) {
	for _, mode := range []auth.Mode{auth.ModeNone, auth.ModeLocal, auth.ModeOIDC, auth.ModeProxy} {
		flags := NewFlagSet("app", "APP", "test")
		var config Authentication
		finalize := RegisterAuthentication(flags, &config, AuthenticationOptions{})
		if err := flags.Parse([]string{"--auth-mode", string(mode)}); err != nil {
			t.Fatalf("canonical mode %q rejected: %v", mode, err)
		}
		finalize()
		if config.ModeOverride == nil || *config.ModeOverride != mode {
			t.Fatalf("canonical mode %q lost after parsing: %#v", mode, config.ModeOverride)
		}
	}
	for _, mode := range []string{"trusted-proxy", "unsupported"} {
		flags := NewFlagSet("app", "APP", "test")
		var config Authentication
		RegisterAuthentication(flags, &config, AuthenticationOptions{})
		if err := flags.Parse([]string{"--auth-mode", mode}); err == nil {
			t.Fatalf("deprecated or unknown mode %q accepted", mode)
		}
	}
}
