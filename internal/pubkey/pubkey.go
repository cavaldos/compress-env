// Package pubkey encrypts the aggregated content for one or more public keys
// so that only the holders of the matching private keys can restore it. It is
// a thin wrapper over age (X25519 keys: "age1..." / "AGE-SECRET-KEY-1...").
package pubkey

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"filippo.io/age"
)

const (
	magic           = "age-encryption.org/v1\n"
	publicKeyPrefix = "age1"
	privateKeyStart = "AGE-SECRET-KEY-"
)

// IsEncrypted reports whether data was produced by Encrypt.
func IsEncrypted(data []byte) bool {
	return bytes.HasPrefix(data, []byte(magic))
}

// GenerateKeyPair returns a new public and private key.
func GenerateKeyPair() (public, private string, err error) {
	id, err := age.GenerateX25519Identity()
	if err != nil {
		return "", "", err
	}
	return id.Recipient().String(), id.String(), nil
}

// ParseRecipients accepts a public key or the path to a file containing one
// (or several, one per line).
func ParseRecipients(input string) ([]age.Recipient, error) {
	text, err := keyText(input, publicKeyPrefix)
	if err != nil {
		return nil, err
	}
	recipients, err := age.ParseRecipients(strings.NewReader(text))
	if err != nil {
		return nil, fmt.Errorf("invalid public key: %w", err)
	}
	return recipients, nil
}

// ParseIdentities accepts a private key or the path to a file containing one.
func ParseIdentities(input string) ([]age.Identity, error) {
	text, err := keyText(input, privateKeyStart)
	if err != nil {
		return nil, err
	}
	identities, err := age.ParseIdentities(strings.NewReader(text))
	if err != nil {
		return nil, fmt.Errorf("invalid private key: %w", err)
	}
	return identities, nil
}

// Encrypt encrypts content for the recipients.
func Encrypt(content string, recipients []age.Recipient) ([]byte, error) {
	var buf bytes.Buffer
	w, err := age.Encrypt(&buf, recipients...)
	if err != nil {
		return nil, err
	}
	if _, err := io.WriteString(w, content); err != nil {
		return nil, err
	}
	if err := w.Close(); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

// Decrypt decrypts data produced by Encrypt.
func Decrypt(data []byte, identities []age.Identity) (string, error) {
	r, err := age.Decrypt(bytes.NewReader(data), identities...)
	if err != nil {
		var noMatch *age.NoIdentityMatchError
		if errors.As(err, &noMatch) {
			return "", errors.New("this private key does not match the public key the file was encrypted with")
		}
		return "", err
	}
	plain, err := io.ReadAll(r)
	if err != nil {
		return "", err
	}
	return string(plain), nil
}

// keyText returns input itself when it already looks like a key, and the
// contents of the file it points to otherwise.
func keyText(input, keyPrefix string) (string, error) {
	input = strings.TrimSpace(input)
	if input == "" {
		return "", errors.New("no key given")
	}
	if strings.HasPrefix(input, keyPrefix) {
		return input, nil
	}

	path := input
	if path == "~" || strings.HasPrefix(path, "~/") {
		if home, err := os.UserHomeDir(); err == nil {
			path = filepath.Join(home, path[1:])
		}
	}
	data, err := os.ReadFile(path)
	if err != nil {
		return "", fmt.Errorf("not a key and cannot read it as a file: %w", err)
	}
	return string(data), nil
}
