# Trusted proxy

Use proxy authentication only when an upstream proxy strips client-supplied identity headers, sets its own verified headers, and is the request's immediate peer. Appkit does not trust any peer by default.

```go
var deployment config.Proxy
config.RegisterProxy(flags, &deployment)
// Register remaining flags, then parse them.

settings, err := proxy.Resolve(proxy.Options{
	Persisted: persistedProxySettings,
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

`resolveUser` maps a verified `auth.ExternalIdentity` to the application's user. Appkit does not own user provisioning, storage, or authorization.

Resolution order is **appkit defaults → explicitly persisted settings → explicitly set CLI or environment flags**. A `proxy.Layer` field is `nil` when absent; a pointer to an empty slice deliberately clears an inherited list. Restrict header names to those your gateway controls.

A sole `*` trusted CIDR permits every valid immediate peer; mixing `*` with other CIDRs is rejected. Wildcard trust does not make forwarded headers safe—network policy must still prevent bypassing the ingress.
