# Secrets

The `secrets` package encrypts persisted application settings with AES-GCM and a deployment-managed key. It is intended for values such as OAuth client secrets or integration tokens that your application needs to store but should not keep in plaintext.

## Create the cipher

The key is optional when encryption is not configured, but when present it must be base64-encoded and exactly 32 bytes after decoding.

```go
cipher, err := secrets.New(os.Getenv("APP_ENCRYPTION_KEY"))
if err != nil {
	return fmt.Errorf("configure secret encryption: %w", err)
}
if !cipher.Configured() {
	return errors.New("secret encryption is required")
}
```

`config.RegisterAuthentication` can register and validate the same deployment key through its `--encryption-key` flag.

## Store ciphertext, not plaintext

```go
encrypted, err := cipher.Encrypt(clientSecret)
if err != nil {
	return err
}
// Persist encrypted in your application database.

plaintext, err := cipher.Decrypt(encrypted)
if err != nil {
	return err
}
```

Each encryption operation uses a fresh random nonce. Ciphertext is versioned so the package can evolve its stored representation. Do not log plaintext, encryption keys, or complete ciphertext values. Rotating a deployment key requires an application migration that decrypts each value with the old key and re-encrypts it with the new key.
