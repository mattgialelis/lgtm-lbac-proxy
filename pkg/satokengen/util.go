package satokengen

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"

	"github.com/xdg-go/pbkdf2"
)

// EncodePassword encodes a password using PBKDF2.
func EncodePassword(password string, salt string) (string, error) {
	newPasswd := pbkdf2.Key([]byte(password), []byte(salt), 10000, 50, sha256.New)
	return hex.EncodeToString(newPasswd), nil
}

const alphanum = "0123456789ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz"

// GetRandomString generates a random alphanumeric string of the specified length,
// optionally using only specified characters
func GetRandomString(n int, alphabets ...byte) (string, error) {
	chars := alphanum
	if len(alphabets) > 0 {
		chars = string(alphabets)
	}
	cnt := len(chars)
	max := 255 / cnt * cnt

	bytes := make([]byte, n)

	randread := n * 5 / 4
	randbytes := make([]byte, randread)

	for i := 0; i < n; {
		if _, err := rand.Read(randbytes); err != nil {
			return "", err
		}

		for j := 0; i < n && j < randread; j++ {
			b := int(randbytes[j])
			if b >= max {
				continue
			}

			bytes[i] = chars[b%cnt]
			i++
		}
	}

	return string(bytes), nil
}
