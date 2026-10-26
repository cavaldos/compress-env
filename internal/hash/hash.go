// Package hash implements the ENVENC01 container: a signature, a little-endian
// u32 length and the content XORed with a key derived from the password.
package hash

import (
	"encoding/binary"
	"errors"
	"fmt"
	"unicode/utf8"
)

const (
	signature  = "ENVENC0"
	version    = '1'
	headerSize = 12
)

var defaultKey = []byte{
	0x42, 0x13, 0x37, 0x89, 0xAB, 0xCD, 0xEF, 0x01,
	0x23, 0x45, 0x67, 0x89, 0xAB, 0xCD, 0xEF, 0x01,
}

// Encode makes content unreadable. An empty password selects the default key.
func Encode(content, password string) []byte {
	encoded := xor([]byte(content), generateKey(password))

	result := make([]byte, 0, headerSize+len(encoded))
	result = append(result, signature...)
	result = append(result, version)
	result = binary.LittleEndian.AppendUint32(result, uint32(len(encoded)))
	return append(result, encoded...)
}

// Decode turns the binary content back into readable text.
func Decode(data []byte, password string) (string, error) {
	if len(data) < headerSize || string(data[:7]) != signature {
		return "", errors.New("Invalid file format: not an encoded .env file")
	}
	if data[7] != version {
		return "", fmt.Errorf("Unsupported version: %c", data[7])
	}

	length := binary.LittleEndian.Uint32(data[8:headerSize])
	if uint64(len(data)-headerSize) < uint64(length) {
		return "", errors.New("File is truncated or corrupted")
	}

	decoded := xor(data[headerSize:headerSize+int(length)], generateKey(password))
	if !utf8.Valid(decoded) {
		return "", errors.New("Invalid UTF-8 sequence or incorrect password")
	}
	return string(decoded), nil
}

func xor(data, key []byte) []byte {
	out := make([]byte, len(data))
	for i, b := range data {
		out[i] = b ^ key[i%len(key)]
	}
	return out
}

func generateKey(password string) []byte {
	if password == "" {
		return defaultKey
	}

	key := make([]byte, 0, 16)
	key = binary.LittleEndian.AppendUint64(key, hashStr(password))
	key = binary.LittleEndian.AppendUint64(key, hashStr(password+"salt1"))
	return key
}
