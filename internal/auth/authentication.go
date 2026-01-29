package auth

import (
	"github.com/alexedwards/argon2id"
)

func HashPassword(password string) (string, error) {

	params := &argon2id.Params{
		Memory:      64 * 1024, // 64 MB
		Iterations:  3,
		Parallelism: 2,
		SaltLength:  16,
		KeyLength:   32,
	}

	hash, err := argon2id.CreateHash(password, params)

	if err != nil {
		return "", err
	}

	return hash, nil

}

func CheckPasswordHash(password, hash string) (bool, error) {

	match, err := argon2id.ComparePasswordAndHash(password, hash)

	if err != nil {
		return false, err
	}

	return match, nil
}
