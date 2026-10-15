# appkit

A composable Go framework for HTTP applications, including configuration, authentication, sessions, logging, encryption, and web serving.

## Install

```sh
go get github.com/containeroo/appkit
```

## Example

Define your application and management routes, then serve them independently:

```go
app := http.NewServeMux()
app.HandleFunc("GET /", home)

management := http.NewServeMux()
management.HandleFunc("GET /healthz", health)
management.HandleFunc("GET /readyz", readiness)

serverConfig := httpserver.Config{
    Application: app,
    Management:  management,
    RoutePrefix: cfg.Server.RoutePrefix,
}
handlers := httpserver.New(serverConfig)
if err := httpserver.Run(
    ctx,
    cfg.Server.ListenAddress,
    cfg.Server.ManagementListenAddress,
    handlers,
    logger.With("component", "server"),
); err != nil {
    logger.Error("run server", "event", "server_run_failed", "error", err)
    return err
}
```

See [`examples/basic`](examples/basic/main.go) for a complete runnable example with configuration flags, logging, graceful shutdown, and both listeners.

```sh
go run ./examples/basic --route-prefix /app

curl http://127.0.0.1:8080/app/
curl -i http://127.0.0.1:8081/healthz
```

## Trusted-proxy configuration

Appkit owns the defaults, priority order, and validation of trusted-proxy headers. Applications should not copy lists of header names or reimplement the fallback logic. Configure the flag group with `config.RegisterProxy`, and resolve the effective settings when **proxy mode is active**:

```go
var deployment config.Proxy
config.RegisterProxy(flags, &deployment)
// Register your other flags, then call flags.Parse(args).

// Application persistence decides which settings are present. Nil means
// inherit; a pointer to an empty slice means intentionally clear the value.
var persisted proxy.Layer
if dbSettings.HasUsernameHeaders {
    persisted.UsernameHeaders = &dbSettings.UsernameHeaders
}
if dbSettings.HasEmailHeaders {
    persisted.EmailHeaders = &dbSettings.EmailHeaders
}

settings, err := proxy.Resolve(proxy.Options{
    Persisted: persisted,
    Overrides: deployment.DeploymentOverrides(flags.Overrides()),
})
if err != nil {
    return err
}

provider, err := proxy.New(settings, "my-auth-gateway", resolveUser)
if err != nil {
    return err
}
```

`resolveUser` is the application's `proxy.Resolver[User]`, which maps a verified `auth.ExternalIdentity` to its own user. Appkit does not own database persistence, user provisioning, or domain authorization. The example `HasUsernameHeaders`/`HasEmailHeaders` fields represent the application's own database presence markers, not appkit fields. Applications with a legacy empty-means-unset convention can simply omit that layer entry for empty stored lists.

The precedence is **appkit defaults → explicitly persisted settings → explicitly set CLI/environment flags**. `proxy.Layer` uses `nil` for an absent value and a pointer to an empty slice for an explicit override that removes inherited values. `config.Proxy.DeploymentOverrides` uses the flag origins to make that distinction. Resolution requires an effective username or email header and a valid trusted CIDR list. An email-only gateway (such as one injecting `X-Auth-Request-Email`) is supported: when no username header has a value, the email becomes the username and stable proxy subject.

No immediate peer is trusted by default. A sole `*` allows every valid peer, and mixing `*` with other CIDRs is rejected. **Wildcard trust does not authenticate requests by itself.** The ingress must strip and replace _all_ identity headers accepted by the effective configuration, and network policy must prevent access that bypasses that ingress. In particular, explicitly restrict accepted header names to those your gateway controls rather than relying on the entire default fallback list.

Use `httpauth.AuthorizeWithError` for authorization checks that can fail independently of access denial.
`web.SPA` renders the entry point as a Go template, while `httpkit/assets.Static` serves static files.
