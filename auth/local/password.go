// Package local provides password hashing and verification for application-owned user stores.
package local

import (
	"errors"

	"github.com/containeroo/appkit/auth"
	"golang.org/x/crypto/bcrypt"
)

// HashPassword hashes a password with bcrypt, using its default work factor.
// Applications must enforce their own password and account policies before calling it.
func HashPassword(password string) (string, error) {
	if password == "" {
		return "", errors.New("password is empty")
	}
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	return string(hash), err
}

// VerifyPassword compares a supplied password with a stored bcrypt hash.
func VerifyPassword(hash, password string) error {
	if bcrypt.CompareHashAndPassword([]byte(hash), []byte(password)) != nil {
		return auth.ErrInvalidCredentials
	}
	return nil
}
