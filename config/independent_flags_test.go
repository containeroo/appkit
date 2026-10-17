package config

import (
	"fmt"
	"strings"
	"testing"

	"github.com/containeroo/tinyflags"
)

// TestIndependentFlags checks that each registration exposes only its own flag.
func TestIndependentFlags(t *testing.T) {
	tests := []struct {
		name     string
		register func(*tinyflags.FlagSet)
	}{
		{"route-prefix", func(f *tinyflags.FlagSet) { RegisterRoutePrefix(f, new(Server)) }},
		{"listen-address", func(f *tinyflags.FlagSet) { RegisterListenAddress(f, new(Server), nil) }},
		{"management-listen-address", func(f *tinyflags.FlagSet) { RegisterManagementListenAddress(f, new(Server), nil) }},
		{"public-url", func(f *tinyflags.FlagSet) { RegisterPublicURL(f, new(Server), "") }},
		{"database-url", func(f *tinyflags.FlagSet) { RegisterDatabaseURL(f, new(Postgres), PostgresOptions{}) }},
		{"database-max-conns", func(f *tinyflags.FlagSet) { RegisterDatabaseMaxConns(f, new(Postgres)) }},
		{"database-min-idle-conns", func(f *tinyflags.FlagSet) { RegisterDatabaseMinIdleConns(f, new(Postgres)) }},
		{"debug", func(f *tinyflags.FlagSet) { RegisterDebug(f, new(Logging)) }},
		{"access-log", func(f *tinyflags.FlagSet) { RegisterAccessLog(f, new(Logging)) }},
		{"disable-metrics", func(f *tinyflags.FlagSet) { RegisterDisableMetrics(f, new(Observability)) }},
		{"performance-diagnostics", func(f *tinyflags.FlagSet) { RegisterPerformanceDiagnostics(f, new(Observability)) }},
		{"otel-service-name", func(f *tinyflags.FlagSet) { RegisterOTelServiceName(f, new(OpenTelemetry), "service") }},
		{"otel-exporter-otlp-endpoint", func(f *tinyflags.FlagSet) { RegisterOTelEndpoint(f, new(OpenTelemetry)) }},
		{"otel-exporter-otlp-insecure", func(f *tinyflags.FlagSet) { RegisterOTelInsecure(f, new(OpenTelemetry)) }},
		{"trusted-proxy-cidrs", func(f *tinyflags.FlagSet) { RegisterTrustedProxyCIDRs(f, new(Proxy)) }},
		{"trusted-username-headers", func(f *tinyflags.FlagSet) { RegisterTrustedUsernameHeaders(f, new(Proxy)) }},
		{"trusted-email-headers", func(f *tinyflags.FlagSet) { RegisterTrustedEmailHeaders(f, new(Proxy)) }},
		{"trusted-display-name-headers", func(f *tinyflags.FlagSet) { RegisterTrustedDisplayNameHeaders(f, new(Proxy)) }},
		{"trusted-group-headers", func(f *tinyflags.FlagSet) { RegisterTrustedGroupHeaders(f, new(Proxy)) }},
		{"trusted-admin-group", func(f *tinyflags.FlagSet) { RegisterTrustedAdminGroup(f, new(Proxy)) }},
		{"oidc-issuer", func(f *tinyflags.FlagSet) { RegisterOIDCIssuer(f, new(OIDC)) }},
		{"oidc-client-id", func(f *tinyflags.FlagSet) { RegisterOIDCClientID(f, new(OIDC)) }},
		{"oidc-client-secret", func(f *tinyflags.FlagSet) { RegisterOIDCClientSecret(f, new(OIDC)) }},
		{"oidc-session-secret", func(f *tinyflags.FlagSet) { RegisterOIDCSessionSecret(f, new(OIDC)) }},
		{"oidc-group-claim", func(f *tinyflags.FlagSet) { RegisterOIDCGroupClaim(f, new(OIDC)) }},
		{"oidc-admin-group", func(f *tinyflags.FlagSet) { RegisterOIDCAdminGroup(f, new(OIDC)) }},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			flags := NewFlagSet("test", "TEST", "test")
			tc.register(flags)
			err := flags.Parse([]string{"--help"})
			if !tinyflags.IsHelpRequested(err) {
				t.Fatalf("expected help: %v", err)
			}
			help := err.Error()
			if !strings.Contains(help, "--"+tc.name) {
				t.Fatalf("flag missing from help: %s", help)
			}
			for _, other := range tests {
				if other.name != tc.name && strings.Contains(help, "--"+other.name+" ") {
					t.Fatalf("unexpected flag %q in help: %s", other.name, help)
				}
			}
		})
	}
}

// TestIndependentServerAddresses checks listener finalizers and prefix normalization.
func TestIndependentServerAddresses(t *testing.T) {
	flags := NewFlagSet("test", "TEST", "test")
	var cfg Server
	RegisterRoutePrefix(flags, &cfg)
	finishListen := RegisterListenAddress(flags, &cfg, nil)
	finishManagement := RegisterManagementListenAddress(flags, &cfg, nil)
	if err := flags.Parse([]string{"--route-prefix=/garage/", "--listen-address=127.0.0.1:9000", "--management-listen-address=127.0.0.1:9001"}); err != nil {
		t.Fatal(err)
	}
	finishListen()
	finishManagement()
	if cfg.RoutePrefix != "/garage" || cfg.ListenAddress != "127.0.0.1:9000" || cfg.ManagementListenAddress != "127.0.0.1:9001" {
		t.Fatalf("unexpected server configuration: %+v", cfg)
	}
}

// TestIndependentFlagValidation checks validators on separately registered flags.
func TestIndependentFlagValidation(t *testing.T) {
	tests := []struct {
		name     string
		register func(*tinyflags.FlagSet)
		args     []string
	}{
		{"negative connections", func(f *tinyflags.FlagSet) { RegisterDatabaseMaxConns(f, new(Postgres)) }, []string{"--database-max-conns=-1"}},
		{"invalid prefix", func(f *tinyflags.FlagSet) { RegisterRoutePrefix(f, new(Server)) }, []string{"--route-prefix=//invalid"}},
		{"empty service name", func(f *tinyflags.FlagSet) { RegisterOTelServiceName(f, new(OpenTelemetry), "service") }, []string{"--otel-service-name="}},
		{"short OIDC secret", func(f *tinyflags.FlagSet) { RegisterOIDCSessionSecret(f, new(OIDC)) }, []string{"--oidc-session-secret=short"}},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			flags := NewFlagSet("test", "TEST", "test")
			tc.register(flags)
			if err := flags.Parse(tc.args); err == nil {
				t.Fatal("invalid value accepted")
			}
		})
	}
}

// TestIndependentFlagsCompose checks parsing, origins and secret masking.
func TestIndependentFlagsCompose(t *testing.T) {
	t.Setenv("TEST__TRUSTED_EMAIL_HEADERS", "X-User-Email")
	flags := NewFlagSet("test", "TEST", "test")
	var database Postgres
	var identity OIDC
	var proxy Proxy
	var telemetry OpenTelemetry
	RegisterDatabaseURL(flags, &database, PostgresOptions{})
	RegisterOIDCClientSecret(flags, &identity)
	RegisterTrustedEmailHeaders(flags, &proxy)
	RegisterOTelEndpoint(flags, &telemetry)
	RegisterOTelInsecure(flags, &telemetry)
	err := flags.Parse([]string{
		"--database-url=postgres://user:password@localhost/database",
		"--oidc-client-secret=very-private-secret",
		"--otel-exporter-otlp-endpoint=localhost:4317",
		"--otel-exporter-otlp-insecure",
	})
	if err != nil {
		t.Fatal(err)
	}
	if len(proxy.EmailHeaders) != 1 || proxy.EmailHeaders[0] != "X-User-Email" || !telemetry.Insecure {
		t.Fatalf("unexpected configuration: %+v %+v", proxy, telemetry)
	}
	layer := proxy.DeploymentOverrides(flags.Overrides())
	if layer.EmailHeaders == nil || len(*layer.EmailHeaders) != 1 || (*layer.EmailHeaders)[0] != "X-User-Email" {
		t.Fatalf("environment override not retained: %+v", layer)
	}
	values := fmt.Sprint(flags.Overrides().Values())
	if strings.Contains(values, "password") || strings.Contains(values, "very-private-secret") {
		t.Fatal("sensitive overrides were not masked")
	}
}
