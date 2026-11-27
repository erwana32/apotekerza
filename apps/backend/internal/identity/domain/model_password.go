package domain

import (
	"crypto/rand"
	"crypto/subtle"
	"encoding/base64"
	"errors"
	"fmt"
	"strings"

	"golang.org/x/crypto/argon2"
)

// Argon2id parameters. The hash string carries them, so raising them later
// stays compatible with existing hashes (they keep verifying with their own
// recorded parameters). OWASP's 2024 minimum for argon2id is 19 MiB, t=2, p=1;
// these are deliberately above it for a single-pharmacy server.
const (
	argonMemory  = 64 * 1024 // 64 MiB
	argonTime    = 3         // iterations
	argonThreads = 2         // parallelism
	argonKeyLen  = 32        // derived key length
	argonSaltLen = 16        // salt length
)

// ErrPasswordMismatch is returned when a password does not match the hash.
var ErrPasswordMismatch = errors.New("identity: password does not match")

// HashPassword returns an argon2id hash in the PHC string format
// $argon2id$v=19$m=<mem>,t=<time>,p=<threads>$<salt>$<hash>.
func HashPassword(password string) (string, error) {
	salt := make([]byte, argonSaltLen)
	if _, err := rand.Read(salt); err != nil {
		return "", fmt.Errorf("identity: generate salt: %w", err)
	}
	key := argon2.IDKey([]byte(password), salt, argonTime, argonMemory, argonThreads, argonKeyLen)
	return fmt.Sprintf("$argon2id$v=%d$m=%d,t=%d,p=%d$%s$%s",
		argon2.Version, argonMemory, argonTime, argonThreads,
		base64.RawStdEncoding.EncodeToString(salt),
		base64.RawStdEncoding.EncodeToString(key),
	), nil
}

// VerifyPassword reports whether password matches encodedHash. It reads the
// parameters from the hash, so a hash created with older settings still works.
func VerifyPassword(encodedHash, password string) error {
	params, salt, key, err := parseArgon2id(encodedHash)
	if err != nil {
		return err
	}
	other := argon2.IDKey([]byte(password), salt, params.time, params.memory, params.threads, uint32(len(key)))
	if subtle.ConstantTimeCompare(key, other) != 1 {
		return ErrPasswordMismatch
	}
	return nil
}

type argonParams struct {
	memory  uint32
	time    uint32
	threads uint8
}

// parseArgon2id splits a PHC argon2id string into its parameters, salt and key.
func parseArgon2id(encodedHash string) (argonParams, []byte, []byte, error) {
	var p argonParams
	parts := strings.Split(encodedHash, "$")
	// "", "argon2id", "v=19", "m=..,t=..,p=..", salt, hash
	if len(parts) != 6 || parts[1] != "argon2id" {
		return p, nil, nil, fmt.Errorf("identity: unsupported password hash format")
	}
	var version int
	if _, err := fmt.Sscanf(parts[2], "v=%d", &version); err != nil {
		return p, nil, nil, fmt.Errorf("identity: parse hash version: %w", err)
	}
	if version != argon2.Version {
		return p, nil, nil, fmt.Errorf("identity: unsupported argon2 version %d", version)
	}
	var mem, t uint32
	var threads uint8
	if _, err := fmt.Sscanf(parts[3], "m=%d,t=%d,p=%d", &mem, &t, &threads); err != nil {
		return p, nil, nil, fmt.Errorf("identity: parse hash parameters: %w", err)
	}
	salt, err := base64.RawStdEncoding.DecodeString(parts[4])
	if err != nil {
		return p, nil, nil, fmt.Errorf("identity: decode salt: %w", err)
	}
	key, err := base64.RawStdEncoding.DecodeString(parts[5])
	if err != nil {
		return p, nil, nil, fmt.Errorf("identity: decode key: %w", err)
	}
	return argonParams{memory: mem, time: t, threads: threads}, salt, key, nil
}
