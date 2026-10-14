package local

import (
	"errors"
	"testing"

	"github.com/containeroo/appkit/auth"
)

func TestHashPassword(t *testing.T) {
	hash, err := HashPassword("correct horse battery staple")
	if err != nil {
		t.Fatal(err)
	}
	if hash == "correct horse battery staple" {
		t.Fatal("password not hashed")
	}
	if err := VerifyPassword(hash, "correct horse battery staple"); err != nil {
		t.Fatal(err)
	}
	if err := VerifyPassword(hash, "wrong"); !errors.Is(err, auth.ErrInvalidCredentials) {
		t.Fatal(err)
	}
}
