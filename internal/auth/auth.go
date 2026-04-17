package auth

import (
	"runtime"

	"github.com/alexedwards/argon2id"
)

func HashPassword(password string) (string, error) {
	params := &argon2id.Params{
		Memory:      128 * 1024,
		Iterations:  4,
		Parallelism: uint8(runtime.NumCPU()),
		SaltLength:  16,
		KeyLength:   32,
	}

	hashedPass, err := argon2id.CreateHash(password, params)
	if err != nil {
		return "", err
	}

	return hashedPass, nil
}

func CheckPasswordHash(password, hash string) (bool, error) {
	isValid, err := argon2id.ComparePasswordAndHash(password, hash)
	if err != nil {
		return false, err
	}

	return isValid, nil
}
