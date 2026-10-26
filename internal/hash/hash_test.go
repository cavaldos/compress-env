package hash

import (
	"encoding/hex"
	"strings"
	"testing"
)

const content = "--- .env ---\nKEY=value\n\n\n"

// Golden files produced by the original Rust implementation.
var golden = map[string]string{
	"":           "454e56454e433031190000006f3e1aa985a8817703684aa4a186aa587f6556e5dea8e50b29",
	"mypassword": "454e56454e43303119000000b69184d6bd5b9c93f447b78f4292f65ca6cac89ae65bf8efde",
}

func TestEncodeMatchesRustOutput(t *testing.T) {
	for password, want := range golden {
		if got := hex.EncodeToString(Encode(content, password)); got != want {
			t.Errorf("Encode(password=%q) = %s, want %s", password, got, want)
		}
	}
}

func TestDecodeRustOutput(t *testing.T) {
	for password, enc := range golden {
		data, _ := hex.DecodeString(enc)
		got, err := Decode(data, password)
		if err != nil || got != content {
			t.Errorf("Decode(password=%q) = %q, %v; want %q", password, got, err, content)
		}
	}
}

func TestRoundTrip(t *testing.T) {
	in := strings.Repeat("TOKEN=tiếng Việt ✓\n", 50)
	for _, password := range []string{"", "p", "a much longer password ü"} {
		got, err := Decode(Encode(in, password), password)
		if err != nil || got != in {
			t.Errorf("round trip failed for password %q: %v", password, err)
		}
	}
}

func TestDecodeErrors(t *testing.T) {
	valid := Encode(content, "pw")

	cases := map[string][]byte{
		"too short":     []byte("ENVENC01"),
		"bad signature": append([]byte("XXXXXXX1"), valid[8:]...),
		"bad version":   append([]byte("ENVENC02"), valid[8:]...),
		"truncated":     valid[:len(valid)-1],
	}
	for name, data := range cases {
		if _, err := Decode(data, "pw"); err == nil {
			t.Errorf("%s: expected error", name)
		}
	}

	if _, err := Decode(valid, "wrong"); err == nil {
		t.Error("wrong password: expected error")
	}
}
