// Package oidc provides a reusable, storage-independent OpenID Connect authorization-code flow.
// Applications resolve the verified identity and issue their own application sessions.
package oidc

import (
	"context"
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/containeroo/appkit/auth"
	coreoidc "github.com/coreos/go-oidc/v3/oidc"
	"golang.org/x/oauth2"
)

const stateLifetime = 10 * time.Minute

// Config supplies deployment-managed OIDC credentials and browser callback settings.
type Config struct {
	// Issuer is the trusted provider issuer URL.
	Issuer string
	// ClientID identifies this client at the provider.
	ClientID string
	// ClientSecret authenticates this client at the token endpoint.
	ClientSecret string
	// StateSecret signs transient login state; must be at least 32 bytes.
	StateSecret string
	// CallbackURL is the externally reachable HTTPS (or local development HTTP) callback URL.
	CallbackURL string
	// CookieName distinguishes this flow from other applications on the same domain.
	CookieName string
	// CookiePath scopes temporary login state to the deployment, defaulting to root.
	CookiePath string
	// GroupClaim is an optional ID-token claim containing groups.
	GroupClaim string
	// Scopes adds provider-specific scopes, if needed, beyond openid/profile/email.
	Scopes []string
}

// Flow provides the provider handshake and verifies callback identity without accessing application persistence.
type Flow struct {
	// oauth is the configured authorization-code client.
	oauth *oauth2.Config
	// verifier validates tokens and their audience against the expected client.
	verifier *coreoidc.IDTokenVerifier
	// issuer namespaces verified external subjects.
	issuer string
	// stateSecret signs temporary login state cookies.
	stateSecret []byte
	// cookieName is the name of the temporary state cookie.
	cookieName string
	// cookiePath controls the browser cookie scope.
	cookiePath string
	// secure indicates whether the callback requires TLS.
	secure bool
	// groupClaim names the optional groups claim in an ID token.
	groupClaim string
}

// loginState binds the current browser to the OAuth state and PKCE verifier.
type loginState struct {
	// State is the random anti-CSRF value bound to the browser.
	State string `json:"s"`
	// Verifier is the PKCE verifier committed during authorization.
	Verifier string `json:"v"`
	// Next is the validated local destination after successful login.
	Next string `json:"n"`
	// Expires is the UNIX timestamp after which the handshake is rejected.
	Expires int64 `json:"e"`
}

// claims contains only optional, presentational identity claims.
type claims struct {
	// Username is an optional provider suggested username.
	Username string `json:"preferred_username"`
	// Email is the asserted email claim.
	Email string `json:"email"`
	// EmailVerified reports the provider's email verification claim.
	EmailVerified bool `json:"email_verified"`
	// Name is an optional display name.
	Name string `json:"name"`
}

// New discovers the OIDC issuer and builds an identity-verifying browser flow.
func New(ctx context.Context, cfg Config) (*Flow, error) {
	if cfg.Issuer == "" || cfg.ClientID == "" || cfg.CallbackURL == "" {
		return nil, errors.New("OIDC issuer, client ID, and callback URL are required")
	}
	if len(cfg.StateSecret) < 32 {
		return nil, errors.New("OIDC state secret must be at least 32 bytes")
	}
	callback, err := url.Parse(cfg.CallbackURL)
	if err != nil || callback.Host == "" || (callback.Scheme != "https" && callback.Scheme != "http") || callback.Fragment != "" || callback.RawQuery != "" || callback.User != nil {
		return nil, errors.New("invalid OIDC callback URL")
	}
	if callback.Scheme == "http" && callback.Hostname() != "localhost" && callback.Hostname() != "127.0.0.1" && callback.Hostname() != "::1" {
		return nil, errors.New("OIDC callback must use HTTPS outside localhost")
	}
	provider, err := coreoidc.NewProvider(ctx, cfg.Issuer)
	if err != nil {
		return nil, fmt.Errorf("discover OIDC provider: %w", err)
	}
	name := cfg.CookieName
	if name == "" {
		name = "appkit_oidc_state"
	}
	path := cfg.CookiePath
	if path == "" {
		path = "/"
	}
	if path[0] != '/' || strings.HasPrefix(path, "//") {
		return nil, errors.New("invalid OIDC cookie path")
	}
	scopes := append([]string{coreoidc.ScopeOpenID, "profile", "email"}, cfg.Scopes...)
	return &Flow{
		oauth: &oauth2.Config{ClientID: cfg.ClientID, ClientSecret: cfg.ClientSecret,
			RedirectURL: cfg.CallbackURL, Endpoint: provider.Endpoint(), Scopes: scopes},
		verifier: provider.Verifier(&coreoidc.Config{ClientID: cfg.ClientID}),
		issuer:   cfg.Issuer, stateSecret: []byte(cfg.StateSecret),
		cookieName: name, cookiePath: path, secure: callback.Scheme == "https", groupClaim: cfg.GroupClaim,
	}, nil
}

// Begin writes a signed state cookie and redirects the browser to the provider.
// The application chooses the post-login next route; unsafe destinations become root.
func (f *Flow) Begin(w http.ResponseWriter, r *http.Request, next string) error {
	bytes := make([]byte, 32)
	if _, err := rand.Read(bytes); err != nil {
		return fmt.Errorf("generate OIDC state: %w", err)
	}
	state := base64.RawURLEncoding.EncodeToString(bytes)
	verifier := oauth2.GenerateVerifier()
	if !localPath(next) {
		next = "/"
	}
	saved := loginState{State: state, Verifier: verifier, Next: next, Expires: time.Now().Add(stateLifetime).Unix()}
	payload, err := json.Marshal(saved)
	if err != nil {
		return err
	}
	encoded := base64.RawURLEncoding.EncodeToString(payload)
	mac := hmac.New(sha256.New, f.stateSecret)
	_, _ = mac.Write([]byte(encoded))
	cookieValue := encoded + "." + base64.RawURLEncoding.EncodeToString(mac.Sum(nil))
	http.SetCookie(w, f.stateCookie(cookieValue, int(stateLifetime.Seconds())))
	http.Redirect(w, r, f.oauth.AuthCodeURL(state, oauth2.S256ChallengeOption(verifier), coreoidc.Nonce(state)), http.StatusFound)
	return nil
}

// Complete validates the browser state, exchanges the code, and verifies the signed ID token.
// The caller must resolve the identity into a local account and create a session itself.
func (f *Flow) Complete(w http.ResponseWriter, r *http.Request) (auth.ExternalIdentity, string, error) {
	var empty auth.ExternalIdentity
	defer http.SetCookie(w, f.stateCookie("", -1))
	cookie, err := r.Cookie(f.cookieName)
	if err != nil {
		return empty, "", auth.ErrInvalidCredentials
	}
	saved, err := f.decodeState(cookie.Value)
	if err != nil || saved.Expires <= time.Now().Unix() || saved.State == "" || saved.Verifier == "" || saved.State != r.URL.Query().Get("state") {
		return empty, "", auth.ErrInvalidCredentials
	}
	if r.URL.Query().Get("error") != "" || r.URL.Query().Get("code") == "" {
		return empty, "", auth.ErrInvalidCredentials
	}
	token, err := f.oauth.Exchange(r.Context(), r.URL.Query().Get("code"), oauth2.VerifierOption(saved.Verifier))
	if err != nil {
		return empty, "", fmt.Errorf("OIDC code exchange: %w", err)
	}
	raw, ok := token.Extra("id_token").(string)
	if !ok || raw == "" {
		return empty, "", auth.ErrInvalidCredentials
	}
	verified, err := f.verifier.Verify(r.Context(), raw)
	if err != nil || verified.Issuer != f.issuer || verified.Subject == "" || verified.Nonce != saved.State {
		return empty, "", auth.ErrInvalidCredentials
	}
	var profile claims
	if err := verified.Claims(&profile); err != nil {
		return empty, "", auth.ErrInvalidCredentials
	}
	identity := auth.ExternalIdentity{
		Issuer: verified.Issuer, Subject: verified.Subject, Username: profile.Username,
		Email: profile.Email, EmailVerified: profile.EmailVerified, DisplayName: profile.Name,
	}
	if f.groupClaim != "" {
		var fields map[string]json.RawMessage
		if err := verified.Claims(&fields); err != nil {
			return empty, "", auth.ErrInvalidCredentials
		}
		if value, ok := fields[f.groupClaim]; ok {
			if err := json.Unmarshal(value, &identity.Groups); err != nil {
				var group string
				if err := json.Unmarshal(value, &group); err != nil {
					return empty, "", auth.ErrInvalidCredentials
				}
				identity.Groups = []string{group}
			}
		}
	}
	if !localPath(saved.Next) {
		return empty, "", auth.ErrInvalidCredentials
	}
	return identity, saved.Next, nil
}

// stateCookie creates one secure HTTP-only transient login-state cookie.
func (f *Flow) stateCookie(value string, age int) *http.Cookie {
	return &http.Cookie{Name: f.cookieName, Value: value, Path: f.cookiePath,
		HttpOnly: true, Secure: f.secure, SameSite: http.SameSiteLaxMode, MaxAge: age}
}

// decodeState verifies the cookie MAC before decoding the login state.
func (f *Flow) decodeState(value string) (loginState, error) {
	var state loginState
	if len(value) > 4096 {
		return state, auth.ErrInvalidCredentials
	}
	parts := strings.Split(value, ".")
	if len(parts) != 2 {
		return state, auth.ErrInvalidCredentials
	}
	provided, err := base64.RawURLEncoding.DecodeString(parts[1])
	if err != nil {
		return state, auth.ErrInvalidCredentials
	}
	mac := hmac.New(sha256.New, f.stateSecret)
	_, _ = mac.Write([]byte(parts[0]))
	if !hmac.Equal(provided, mac.Sum(nil)) {
		return state, auth.ErrInvalidCredentials
	}
	payload, err := base64.RawURLEncoding.DecodeString(parts[0])
	if err != nil {
		return state, auth.ErrInvalidCredentials
	}
	if err := json.Unmarshal(payload, &state); err != nil {
		return state, auth.ErrInvalidCredentials
	}
	return state, nil
}

// localPath accepts only local URLs for post-login redirects.
func localPath(value string) bool {
	return strings.HasPrefix(value, "/") && !strings.HasPrefix(value, "//") && !strings.ContainsAny(value, "\\\r\n")
}
