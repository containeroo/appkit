package secrets

import (
	"encoding/base64"
	"errors"
	"strings"
	"testing"
)

func TestCipherRoundTripAndVersionCompatibility(t *testing.T) {
	key := base64.StdEncoding.EncodeToString([]byte("0123456789abcdef0123456789abcdef"))
	cipher, err := New(key)
	if err != nil {
		t.Fatal(err)
	}
	first, err := cipher.Encrypt("private token")
	if err != nil {
		t.Fatal(err)
	}
	second, err := cipher.Encrypt("private token")
	if err != nil {
		t.Fatal(err)
	}
	if !strings.HasPrefix(first, "v1:") || first == second {
		t.Fatal("bad version or nonce reuse")
	}
	plain, err := cipher.Decrypt(first)
	if err != nil || plain != "private token" {
		t.Fatalf("roundtrip: %q %v", plain, err)
	}
	if _, err := cipher.Decrypt("v2:wrong"); err == nil {
		t.Fatal("unsupported format accepted")
	}
}
func TestCipherUnconfiguredAndInvalidKey(t *testing.T) {
	cipher, err := New("")
	if err != nil {
		t.Fatal(err)
	}
	if _, err := cipher.Encrypt("secret"); !errors.Is(err, ErrNotConfigured) {
		t.Fatal(err)
	}
	if _, err := New("not-base64"); err == nil {
		t.Fatal("invalid key accepted")
	}
}

func TestParseKey(t *testing.T) {
	value := base64.StdEncoding.EncodeToString([]byte("0123456789abcdef0123456789abcdef"))

	key, err := ParseKey(value)
	if err != nil {
		t.Fatal(err)
	}
	if len(key) != 32 {
		t.Fatalf("key length: %d", len(key))
	}
}
