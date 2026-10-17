package config

import (
	"encoding/base64"
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/containeroo/appkit/auth"
	"github.com/containeroo/tinyflags"
)

// TestIndependentAuthenticationFlags checks selected flags and their finalizers.
func TestIndependentAuthenticationFlags(t *testing.T) {
	flags := NewFlagSet("example", "EXAMPLE", "test")
	var cfg Authentication
	finishMode := RegisterAuthMode(flags, &cfg)
	finishRegistration := RegisterUserRegistration(flags, &cfg)
	RegisterLocalLogin(flags, &cfg)
	if err := flags.Parse([]string{"--auth-mode=proxy", "--allow-user-registration=false", "--local-login"}); err != nil {
		t.Fatal(err)
	}
	finishMode()
	finishRegistration()
	if cfg.ModeOverride == nil || *cfg.ModeOverride != auth.ModeProxy {
		t.Fatalf("mode override: %v", cfg.ModeOverride)
	}
	if cfg.AllowRegistrationOverride == nil || *cfg.AllowRegistrationOverride || !cfg.LocalLogin {
		t.Fatalf("registration or local login: %+v", cfg)
	}
	if err := flags.Parse([]string{"--encryption-key=invalid"}); err == nil {
		t.Fatal("unregistered encryption key was accepted")
	}
}

// TestIndependentAuthenticationDefaults verifies missing overrides remain unset.
func TestIndependentAuthenticationDefaults(t *testing.T) {
	flags := NewFlagSet("example", "EXAMPLE", "test")
	var cfg Authentication
	finishMode := RegisterAuthMode(flags, &cfg)
	finishRegistration := RegisterUserRegistration(flags, &cfg)
	if err := flags.Parse(nil); err != nil {
		t.Fatal(err)
	}
	finishMode()
	finishRegistration()
	if cfg.ModeOverride != nil || cfg.AllowRegistrationOverride != nil {
		t.Fatalf("unset flags became overrides: %+v", cfg)
	}
}

// TestIndependentAuthenticationEnvironment checks individual environment flags.
func TestIndependentAuthenticationEnvironment(t *testing.T) {
	t.Setenv("EXAMPLE__AUTH_MODE", "oidc")
	t.Setenv("EXAMPLE__ALLOW_USER_REGISTRATION", "true")
	t.Setenv("EXAMPLE__LOCAL_LOGIN", "true")
	flags := NewFlagSet("example", "EXAMPLE", "test")
	var cfg Authentication
	finishMode := RegisterAuthMode(flags, &cfg)
	finishRegistration := RegisterUserRegistration(flags, &cfg)
	RegisterLocalLogin(flags, &cfg)
	if err := flags.Parse(nil); err != nil {
		t.Fatal(err)
	}
	finishMode()
	finishRegistration()
	if cfg.ModeOverride == nil || *cfg.ModeOverride != auth.ModeOIDC ||
		cfg.AllowRegistrationOverride == nil || !*cfg.AllowRegistrationOverride || !cfg.LocalLogin {
		t.Fatalf("environment flags: %+v", cfg)
	}
}

// TestIndependentEncryptionKey checks validation and secret masking.
func TestIndependentEncryptionKey(t *testing.T) {
	key := base64.StdEncoding.EncodeToString([]byte("01234567890123456789012345678901"))
	flags := NewFlagSet("example", "EXAMPLE", "test")
	var cfg Authentication
	RegisterEncryptionKey(flags, &cfg)
	if err := flags.Parse([]string{"--encryption-key", key}); err != nil {
		t.Fatal(err)
	}
	if cfg.EncryptionKey != key || strings.Contains(fmt.Sprint(flags.Overrides().Values()), key) {
		t.Fatal("encryption key was lost or exposed")
	}
	invalid := NewFlagSet("example", "EXAMPLE", "test")
	RegisterEncryptionKey(invalid, &cfg)
	if err := invalid.Parse([]string{"--encryption-key=invalid"}); err == nil {
		t.Fatal("invalid encryption key was accepted")
	}
}

// TestIndependentSessionLifetime checks defaults and validation.
func TestIndependentSessionLifetime(t *testing.T) {
	for _, tc := range []struct {
		name            string
		defaultLifetime time.Duration
		args            []string
		want            time.Duration
		valid           bool
	}{
		{"default", 0, nil, 12 * time.Hour, true},
		{"custom default", 4 * time.Hour, nil, 4 * time.Hour, true},
		{"explicit", 0, []string{"--session-lifetime=2h"}, 2 * time.Hour, true},
		{"zero", 0, []string{"--session-lifetime=0s"}, 0, false},
	} {
		t.Run(tc.name, func(t *testing.T) {
			flags := NewFlagSet("example", "EXAMPLE", "test")
			var cfg Authentication
			RegisterSessionLifetime(flags, &cfg, tc.defaultLifetime)
			err := flags.Parse(tc.args)
			if (err == nil) != tc.valid {
				t.Fatalf("unexpected parse result: %v", err)
			}
			if tc.valid && cfg.SessionLifetime != tc.want {
				t.Fatalf("session lifetime: %s", cfg.SessionLifetime)
			}
		})
	}
}

// TestIndependentHelpShowsOnlySelectedFlags checks composable help output.
func TestIndependentHelpShowsOnlySelectedFlags(t *testing.T) {
	flags := NewFlagSet("example", "EXAMPLE", "test")
	var cfg Authentication
	RegisterLocalLogin(flags, &cfg)
	err := flags.Parse([]string{"--help"})
	if !tinyflags.IsHelpRequested(err) {
		t.Fatalf("expected help, got %v", err)
	}
	if !strings.Contains(err.Error(), "--local-login") || strings.Contains(err.Error(), "--encryption-key") {
		t.Fatalf("unexpected help: %s", err)
	}
}
