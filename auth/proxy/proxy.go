// Package proxy accepts externally asserted identities only from explicitly trusted requests.
// Trusted network placement alone is insufficient if the upstream forwards spoofed headers.
package proxy

import (
	"context"
	"errors"
	"net/http"
	"strings"

	"github.com/containeroo/appkit/auth"
)

// TrustFunc returns true only when a request came through a verified ingress.
// The ingress must strip client-supplied identity headers and set its own.
type TrustFunc func(*http.Request) bool

// Resolver maps a verified external identity into an application-owned user.
type Resolver[User any] func(context.Context, auth.ExternalIdentity) (User, error)

// Authenticator enforces an ingress trust boundary before reading identity headers.
type Authenticator[User any] struct {
	// Trust validates the immediate upstream. Nil denies all requests.
	Trust TrustFunc
	// Resolve maps the external identity into a local user.
	Resolve Resolver[User]
	// UsernameHeaders are tried in order until a nonempty value is found.
	UsernameHeaders []string
	// EmailHeaders are optional email-identity header names.
	EmailHeaders []string
	// DisplayNameHeaders are optional human-readable name headers.
	DisplayNameHeaders []string
	// GroupHeaders contain optional comma-separated group names.
	GroupHeaders []string
	// Issuer is a stable deployment-specific namespace for asserted users.
	Issuer string
}

// Validate checks that ingress trust and identity resolution are explicitly configured.
func (a Authenticator[User]) Validate() error {
	if a.Trust == nil || a.Resolve == nil {
		return errors.New("trusted proxy requires trust check and identity resolver")
	}
	if a.Issuer == "" || (len(a.UsernameHeaders) == 0 && len(a.EmailHeaders) == 0) {
		return errors.New("trusted proxy requires issuer and username or email headers")
	}
	return nil
}

// Authenticate converts trusted headers into a verified external identity.
func (a Authenticator[User]) Authenticate(r *http.Request) (User, error) {
	var zero User
	if err := a.Validate(); err != nil {
		return zero, err
	}
	if !a.Trust(r) {
		return zero, auth.ErrUnauthenticated
	}
	username := firstHeader(r, a.UsernameHeaders)
	email := firstHeader(r, a.EmailHeaders)
	if username == "" {
		username = email
	}
	if username == "" {
		return zero, auth.ErrUnauthenticated
	}
	identity := auth.ExternalIdentity{
		Issuer: a.Issuer, Subject: username, Username: username,
		Email: email, DisplayName: firstHeader(r, a.DisplayNameHeaders),
		Groups: groups(firstHeader(r, a.GroupHeaders)),
	}
	return a.Resolve(r.Context(), identity)
}

// firstHeader returns the first nonempty value from configured header names.
func firstHeader(request *http.Request, headers []string) string {
	for _, name := range headers {
		if value := strings.TrimSpace(request.Header.Get(name)); value != "" {
			return value
		}
	}
	return ""
}

// groups parses comma-separated proxy-provided group memberships.
func groups(value string) []string {
	var result []string
	for _, group := range strings.Split(value, ",") {
		if name := strings.TrimSpace(group); name != "" {
			result = append(result, name)
		}
	}
	return result
}
