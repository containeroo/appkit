# appkit

![appkit](docs/content/assets/appkit-wordmark.svg)

A composable Go toolkit for HTTP applications: configuration, authentication, sessions, logging, encryption, and serving.

Full guides and runnable examples live in the [documentation site](https://containeroo.github.io/appkit/).

## Install

```sh
go get github.com/containeroo/appkit
```

### Browser authentication and first-run setup

Pass resolved settings to `auth/browser.Policy`. The browser authenticator checks
the stored session method, the current OIDC issuer and optional local recovery.
Supply a primary request authenticator for proxy or explicitly anonymous access.
Only a missing or inapplicable session permits fallback. Storage errors propagate.

Use `auth/provisioning.Accounts` to map a verified issuer and subject to an enabled
local user. Its access policy controls registration and the administrator requirement.
The store is responsible for atomic identity linking and returning current privileges.

Use `auth/local.Setup` for the first administrator. It accepts a credential validator,
an optional bootstrap-token digest and an atomic create callback. The callback must
reject repeated setup and persist any initial external link in the same transaction.
Issue a session after successful setup if the deployment permits local sign-in.
