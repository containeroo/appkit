# Configuration

The `config` package registers reusable deployment flags with `tinyflags`; it deliberately does not parse your application's flags or own its top-level configuration. That boundary lets an application keep policy in its own database or configuration layer while accepting explicit operational overrides.

## Available groups

| Group | Purpose |
| --- | --- |
| `RegisterServer` | Application and management listener addresses plus an optional route prefix. |
| `RegisterLogging` | Text or JSON structured logging and debug level. |
| `RegisterLogFormat` | Only the shared text/JSON format flag for CLI commands without server diagnostics. |
| `RegisterAuthentication` | Canonical `none`, `local`, `oidc`, and `proxy` modes, recovery login, encryption key, and optional session lifetime overrides. |
| `RegisterProxy` | Trusted-proxy CIDRs and identity-header overrides. |
| `RegisterOIDC` | Provider credentials and callback deployment settings. |
| `RegisterPostgres` | Optional PostgreSQL connection settings. |
| `RegisterObservability` / `RegisterOpenTelemetry` | Metrics, diagnostics, and OTLP exporter settings. |

## Parse, then finalize

Some groups need a finalizer because an application must distinguish an absent value from an explicit false or empty override. Capture and invoke that finalizer only after parsing.

```go
var authentication config.Authentication
finishAuthentication := config.RegisterAuthentication(
	flags,
	&authentication,
	config.AuthenticationOptions{RegisterSessionLifetime: true},
)

if err := flags.Parse(os.Args[1:]); err != nil {
	return err
}
finishAuthentication()
```

For example, `ModeOverride == nil` means deployment did not select a mode; a non-nil value means it did. The same distinction applies to `AllowRegistrationOverride`, including an explicit `false`.

CLI commands that only need structured output can register the format without adding `--debug` or `--access-log`:

```go
var format logging.LogFormat
finishLogFormat := config.RegisterLogFormat(flags, &format)

if err := flags.Parse(os.Args[1:]); err != nil {
	return err
}
finishLogFormat()
```

## Secrets and logs

Treat deployment configuration as sensitive. The authentication group validates the base64 application encryption key and masks it in help and override output. Use JSON logs in production when they are sent to a log collector; use text logs for local development and direct terminal inspection.

Configuration is deployment input, not an authorization system. Do not expose management endpoints, database settings, or identity-header configuration through public application routes.

## Authentication mode values (v0.6.0)

The accepted modes are `none`, `local`, `oidc`, and `proxy`. The old
`trusted-proxy` mode value and Go constant were removed. Configuration with the
old value fails validation instead of silently switching modes. Applications
with persisted authentication policy must migrate stored mode values before
running the new version; AppKit intentionally does not own application SQL.

Trusted-proxy **header configuration** (including `--trusted-proxy-cidrs`)
remains supported; only the authentication mode alias was removed.
