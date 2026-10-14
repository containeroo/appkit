package config

import (
	"strings"
	"testing"
	"time"

	"github.com/containeroo/appkit/auth"
	"github.com/containeroo/appkit/logging"
)

func TestComposableFlagGroups(t *testing.T) {
	tf := NewFlagSet("example", "EXAMPLE", "test")
	tf.SectionOrder(SectionServer, SectionDatabase, SectionAuthentication, SectionOIDC, SectionLogging)
	var server Server
	var database Postgres
	var authentication Authentication
	var logs Logging
	finishServer := RegisterServer(tf, &server, ServerOptions{})
	RegisterPostgres(tf, &database, PostgresOptions{})
	finishAuth := RegisterAuthentication(tf, &authentication, AuthenticationOptions{RegisterSessionLifetime: true})
	finishLogs := RegisterLogging(tf, &logs)
	if err := tf.Parse([]string{"--database-url", "postgres://local/test", "--route-prefix", "/example",
		"--auth-mode", "proxy", "--allow-user-registration=false", "--session-lifetime", "24h", "--log-format", "text"}); err != nil {
		t.Fatal(err)
	}
	finishServer()
	finishAuth()
	finishLogs()
	if server.RoutePrefix != "/example" {
		t.Errorf("route prefix: %q", server.RoutePrefix)
	}
	if !strings.Contains(server.ListenAddress, ":8080") {
		t.Errorf("listen: %q", server.ListenAddress)
	}
	if database.URL != "postgres://local/test" {
		t.Errorf("db URL: %q", database.URL)
	}
	if authentication.ModeOverride == nil || *authentication.ModeOverride != auth.ModeProxy {
		t.Errorf("mode: %v", authentication.ModeOverride)
	}
	if authentication.AllowRegistrationOverride == nil || *authentication.AllowRegistrationOverride {
		t.Errorf("false override was lost: %v", authentication.AllowRegistrationOverride)
	}
	if authentication.SessionLifetime != 24*time.Hour {
		t.Errorf("duration: %s", authentication.SessionLifetime)
	}
	if logs.Format != logging.LogFormatText {
		t.Errorf("log format: %q", logs.Format)
	}
}

func TestUnsetOverridesRemainNil(t *testing.T) {
	tf := NewFlagSet("example", "EXAMPLE", "test")
	var cfg Authentication
	finalize := RegisterAuthentication(tf, &cfg, AuthenticationOptions{})
	if err := tf.Parse(nil); err != nil {
		t.Fatal(err)
	}
	finalize()
	if cfg.ModeOverride != nil || cfg.AllowRegistrationOverride != nil {
		t.Fatalf("unset flags became overrides: %#v", cfg)
	}
}
