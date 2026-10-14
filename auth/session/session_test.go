package session

import (
	"context"
	"errors"
	"net/http"
	"testing"
	"time"

	"github.com/containeroo/appkit/auth"
)

type memoryStore struct {
	stored map[string]string
	expiry map[string]time.Time
}

func (s *memoryStore) Create(_ context.Context, digest string, user string, expiry time.Time) error {
	if s.stored == nil {
		s.stored = make(map[string]string)
		s.expiry = make(map[string]time.Time)
	}
	s.stored[digest], s.expiry[digest] = user, expiry
	return nil
}
func (s *memoryStore) Lookup(_ context.Context, digest string) (string, error) {
	if user, ok := s.stored[digest]; ok && time.Now().Before(s.expiry[digest]) {
		return user, nil
	}
	return "", ErrNotFound
}
func (s *memoryStore) Revoke(_ context.Context, digest string) error {
	delete(s.stored, digest)
	return nil
}

func TestSessionLifecycle(t *testing.T) {
	store := new(memoryStore)
	manager := Manager[string]{Store: store, Lifetime: time.Hour, Purpose: "test-local"}
	token, err := manager.Start(context.Background(), "alice")
	if err != nil || len(token) < 40 {
		t.Fatalf("issue session: %q %v", token, err)
	}
	if _, found := store.stored[token]; found {
		t.Fatal("raw session token was persisted")
	}
	got, err := manager.Resolve(context.Background(), token)
	if err != nil || got != "alice" {
		t.Fatalf("lookup: %q %v", got, err)
	}
	if _, err := manager.Resolve(context.Background(), "wrong"); !errors.Is(err, auth.ErrUnauthenticated) {
		t.Fatalf("invalid token: %v", err)
	}
	if err := manager.End(context.Background(), token); err != nil {
		t.Fatal(err)
	}
	if _, err := manager.Resolve(context.Background(), token); !errors.Is(err, auth.ErrUnauthenticated) {
		t.Fatalf("revoked session: %v", err)
	}
}

func TestSessionPurposeIsolation(t *testing.T) {
	if TokenHash("local", "token") == TokenHash("bootstrap", "token") {
		t.Fatal("hashes must be purpose-separated")
	}
	if TokenHash("local", "token") == "token" {
		t.Fatal("hash must not be raw token")
	}
}

func TestSessionConfiguration(t *testing.T) {
	cases := []Manager[string]{{}, {Store: new(memoryStore), Lifetime: time.Hour}, {Store: new(memoryStore), Purpose: "test"}}
	for _, item := range cases {
		if err := item.Validate(); err == nil {
			t.Fatal("invalid session configuration was accepted")
		}
	}
}

func TestCookieSecuritySettings(t *testing.T) {
	cfg := CookieConfig{Name: "sid", Path: "/motus/", Secure: true, Lifetime: time.Hour}
	cookie := cfg.Cookie("token", false)
	if !cookie.HttpOnly || !cookie.Secure || cookie.Path != "/motus/" || cookie.SameSite != http.SameSiteLaxMode {
		t.Fatalf("unexpected cookie: %#v", cookie)
	}
	cleared := cfg.Cookie("token", true)
	if cleared.MaxAge != -1 || cleared.Value != "" {
		t.Fatalf("unexpected cleared cookie: %#v", cleared)
	}
}
