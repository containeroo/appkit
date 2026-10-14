package oidc

import (
	"errors"
	"testing"

	"github.com/containeroo/appkit/auth"
)

func TestLocalPath(t *testing.T) {
	for _, path := range []string{"/", "/pages/home", "/training?current=1"} {
		if !localPath(path) {
			t.Fatalf("local path rejected: %q", path)
		}
	}
	for _, path := range []string{"", "https://evil.example", "//evil.example", "/\\evil.example", "/\r\nLocation: evil"} {
		if localPath(path) {
			t.Fatalf("unsafe path accepted: %q", path)
		}
	}
}
func TestRejectUnsignedState(t *testing.T) {
	flow := &Flow{stateSecret: []byte("state secret 32 bytes min long for test")}
	if _, err := flow.decodeState("payload.signature"); !errors.Is(err, auth.ErrInvalidCredentials) {
		t.Fatalf("unsigned state accepted: %v", err)
	}
}
