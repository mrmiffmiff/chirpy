package auth

import (
	"fmt"

	"github.com/alexedwards/argon2id"
)

func HashPassword(password string) (string, error) {
	hashed, err := argon2id.CreateHash(password, argon2id.DefaultParams)
	if err != nil {
		return "", fmt.Errorf("Problem hashing password: %w\n", err)
	}
	return hashed, nil
}

func CheckPasswordHash(password, hash string) (bool, error) {
	match, err := argon2id.ComparePasswordAndHash(password, hash)
	if err != nil {
		return false, fmt.Errorf("Problem comparing password and hash: %w\n", err)
	}
	return match, nil
}
