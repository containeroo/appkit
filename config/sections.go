// Package config registers optional, reusable deployment flag groups with tinyflags.
// It never parses application flags or owns the application's top-level configuration.
package config

import (
	"fmt"
	"strings"

	"github.com/containeroo/tinyflags"
)

const (
	// SectionServer groups public and management HTTP server settings.
	SectionServer = "Server"
	// SectionDatabase groups database connection and pool settings.
	SectionDatabase = "Database"
	// SectionAuthentication groups deployment-level authentication overrides.
	SectionAuthentication = "Authentication"
	// SectionProxy groups trusted-proxy header settings.
	SectionProxy = "Trusted proxy"
	// SectionOIDC groups OIDC deployment secrets and overrides.
	SectionOIDC = "OIDC"
	// SectionTelemetry groups OpenTelemetry export flags.
	SectionTelemetry = "OpenTelemetry"
	// SectionLogging groups logs, metrics, and request diagnostics.
	SectionLogging = "Logging and diagnostics"
)

// NewFlagSet configures the common CLI conventions while leaving parsing and
// section ordering to the application. The envName is the application name,
// not a full environment variable (for example "MOTUS").
func NewFlagSet(name, envName, version string) *tinyflags.FlagSet {
	flags := tinyflags.NewFlagSet(name, tinyflags.ContinueOnError)
	env := strings.ToUpper(strings.TrimSuffix(envName, "_"))
	flags.EnvPrefix(env + "_")
	flags.Version(version)
	flags.HideEnvs()
	flags.Note(fmt.Sprintf("\nFlags can also be set using %s__<FLAG_NAME> environment variables; for example %s__LISTEN_ADDRESS.", env, env))
	return flags
}
