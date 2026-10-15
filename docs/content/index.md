# appkit

`appkit` is a composable Go toolkit for HTTP applications. It keeps application-owned concerns—users, persistence, authorization, and business policy—outside the toolkit while providing well-defined boundaries for configuration, identity, sessions, logging, encryption, and HTTP serving.

Start with the [quickstart](/appkit/guides/quickstart/) to assemble a small application, then select the authentication and deployment guides that match your environment. Appkit provides contracts and secure defaults; your application remains the authority for user records, persistence, and authorization rules.

## Start here

- [Quickstart](/appkit/guides/quickstart/): wire flags, logging, and independent application and management listeners.
- [Configuration](/appkit/guides/configuration/): register reusable deployment flag groups without handing control of application policy to the toolkit.
- [HTTP server](/appkit/guides/http-server/): mount application and operational routes safely.

## Identity and access

- [Local authentication](/appkit/guides/local-authentication/): authenticate passwords with an application-owned credential store.
- [Trusted proxy](/appkit/guides/trusted-proxy/): configure a safe reverse-proxy identity boundary.
- [OIDC](/appkit/guides/oidc/): verify OpenID Connect identities before creating application sessions.
- [Sessions](/appkit/guides/sessions/): persist hashed, opaque browser sessions.
- [Middleware and authorization](/appkit/guides/middleware/): attach a user to a request and enforce authentication or policy.

## Application services

- [Secrets](/appkit/guides/secrets/): encrypt deployment-managed application settings.
- [Single-page applications](/appkit/guides/spa/): serve an embedded SPA with browser-navigation fallback.

## Packages

- [`auth`](https://pkg.go.dev/github.com/containeroo/appkit/auth) defines the application-independent authentication contracts.
- [`config`](https://pkg.go.dev/github.com/containeroo/appkit/config) registers reusable deployment flag groups.
- [`httpauth`](https://pkg.go.dev/github.com/containeroo/appkit/httpauth) supplies authentication and authorization middleware.
- [`httpserver`](https://pkg.go.dev/github.com/containeroo/appkit/httpserver) composes application and operational HTTP handlers.
- [`secrets`](https://pkg.go.dev/github.com/containeroo/appkit/secrets) encrypts application settings with a deployment-managed key.
