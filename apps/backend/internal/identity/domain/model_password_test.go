package domain

import (
	"encoding/base64"
	"errors"
	"fmt"
	"strings"
	"testing"
	"time"

	"golang.org/x/crypto/argon2"
)

func TestHashPassword_RoundTrip(t *testing.T) {
	hash, err := HashPassword("rahasia-ini-uji1")
	if err != nil {
		t.Fatalf("hash: %v", err)
	}
	if !strings.HasPrefix(hash, "$argon2id$v=19$") {
		t.Fatalf("hash must be a PHC argon2id string, got %s", hash)
	}
	if strings.Contains(hash, "rahasia-ini-uji1") {
		t.Fatal("hash must not contain the password")
	}
	if err := VerifyPassword(hash, "rahasia-ini-uji1"); err != nil {
		t.Fatalf("verify correct password: %v", err)
	}
	if err := VerifyPassword(hash, "salah-total"); !errors.Is(err, ErrPasswordMismatch) {
		t.Fatalf("verify wrong password: want ErrPasswordMismatch, got %v", err)
	}
}

func TestHashPassword_UniqueSaltPerHash(t *testing.T) {
	first, err := HashPassword("sama-di-dua-hash")
	if err != nil {
		t.Fatalf("hash: %v", err)
	}
	second, err := HashPassword("sama-di-dua-hash")
	if err != nil {
		t.Fatalf("hash: %v", err)
	}
	if first == second {
		t.Fatal("two hashes of the same password must differ: the salt must be random")
	}
}

func TestVerifyPassword_RejectsMalformedHash(t *testing.T) {
	tests := []struct {
		name string
		hash string
	}{
		{"empty", ""},
		{"not argon2id", "$argon2i$v=19$m=65536,t=3,p=2$c2FsdA$aGFzaA"},
		{"wrong version", "$argon2id$v=18$m=65536,t=3,p=2$c2FsdA$aGFzaA"},
		{"bad salt base64", "$argon2id$v=19$m=65536,t=3,p=2$***$aGFzaA"},
		{"too few parts", "$argon2id$v=19$m=65536,t=3,p=2"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := VerifyPassword(tt.hash, "apa-saja"); err == nil {
				t.Fatal("a malformed hash must not verify")
			}
			if err := VerifyPassword(tt.hash, "apa-saja"); errors.Is(err, ErrPasswordMismatch) {
				t.Fatal("a malformed hash must be a format error, not a mismatch")
			}
		})
	}
}

func TestVerifyPassword_ReadsParametersFromHash(t *testing.T) {
	// A hash produced with different parameters must still verify, because the
	// parameters travel inside the string.
	custom := harnessArgonHash(t)
	if err := VerifyPassword(custom, "kata-sandi-lama"); err != nil {
		t.Fatalf("hash with older parameters must verify: %v", err)
	}
}

// harnessArgonHash builds a PHC argon2id string with non-default parameters, so
// a test can prove VerifyPassword reads the parameters from the hash rather
// than using the constants.
func harnessArgonHash(t *testing.T) string {
	t.Helper()
	salt := []byte("0123456789abcdef")
	key := argon2.IDKey([]byte("kata-sandi-lama"), salt, 2, 32*1024, 1, 16)
	return fmt.Sprintf("$argon2id$v=%d$m=%d,t=%d,p=%d$%s$%s",
		argon2.Version, 32*1024, 2, 1,
		base64.RawStdEncoding.EncodeToString(salt),
		base64.RawStdEncoding.EncodeToString(key),
	)
}

func TestBurnPasswordCheck_CostsAboutAsMuchAsARealVerify(t *testing.T) {
	hash, err := HashPassword("rahasia-ini-uji1")
	if err != nil {
		t.Fatalf("hash: %v", err)
	}
	BurnPasswordCheck("warm-up") // builds the dummy hash outside the timing

	start := time.Now()
	_ = VerifyPassword(hash, "salah")
	verify := time.Since(start)

	start = time.Now()
	BurnPasswordCheck("salah")
	burn := time.Since(start)

	// An unknown username must not answer visibly faster than a known one.
	if burn < verify/2 {
		t.Fatalf("burn took %v but a real verify took %v: login timing would reveal usernames", burn, verify)
	}
}
