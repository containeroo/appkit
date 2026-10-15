# appkit

`appkit` is a composable Go toolkit for HTTP applications. It keeps application-owned concerns—users, persistence, authorization, and business policy—outside the toolkit while providing well-defined boundaries for configuration, identity, sessions, logging, encryption, and HTTP serving.

## Guides

- [HTTP server](/appkit/guides/http-server/): application and management listeners.
- [Trusted proxy](/appkit/guides/trusted-proxy/): configure a safe reverse-proxy identity boundary.
- [OIDC](/appkit/guides/oidc/): verify OpenID Connect identities before creating application sessions.
- [Sessions](/appkit/guides/sessions/): persist hashed, opaque browser sessions.

## Packages

- [`auth`](https://pkg.go.dev/github.com/containeroo/appkit/auth) defines the application-independent authentication contracts.
- [`config`](https://pkg.go.dev/github.com/containeroo/appkit/config) registers reusable deployment flag groups.
- [`httpauth`](https://pkg.go.dev/github.com/containeroo/appkit/httpauth) supplies authentication and authorization middleware.
- [`httpserver`](https://pkg.go.dev/github.com/containeroo/appkit/httpserver) composes application and operational HTTP handlers.
- [`secrets`](https://pkg.go.dev/github.com/containeroo/appkit/secrets) encrypts application settings with a deployment-managed key.
