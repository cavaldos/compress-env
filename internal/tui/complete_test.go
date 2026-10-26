package tui

import (
	"os"
	"path/filepath"
	"reflect"
	"testing"

	"filippo.io/age"
	"github.com/cavaldos/compress-env/internal/hash"
	"github.com/cavaldos/compress-env/internal/pubkey"
	"github.com/charmbracelet/bubbles/textinput"
)

func writeFile(t *testing.T, path string, data []byte) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, data, 0o644); err != nil {
		t.Fatal(err)
	}
}

func completionFixture(t *testing.T) string {
	dir := t.TempDir()
	pw := hash.Encode("--- .env ---\nA=1\n\n", "pw")
	age, err := pubkey.Encrypt("x", mustRecipients(t))
	if err != nil {
		t.Fatal(err)
	}
	writeFile(t, filepath.Join(dir, ".env.bin"), pw)
	writeFile(t, filepath.Join(dir, "prod.bin"), pw)
	writeFile(t, filepath.Join(dir, "keys.bin"), age)
	writeFile(t, filepath.Join(dir, "notes.txt"), []byte("hello"))
	writeFile(t, filepath.Join(dir, "backups", "old.bin"), pw)
	writeFile(t, filepath.Join(dir, ".hidden", "x.bin"), pw)
	writeFile(t, filepath.Join(dir, "node_modules", "x.bin"), pw)
	return dir
}

func mustRecipients(t *testing.T) []age.Recipient {
	pub, _, err := pubkey.GenerateKeyPair()
	if err != nil {
		t.Fatal(err)
	}
	r, err := pubkey.ParseRecipients(pub)
	if err != nil {
		t.Fatal(err)
	}
	return r
}

func TestCompletions(t *testing.T) {
	dir := completionFixture(t)
	sep := string(filepath.Separator)

	tests := []struct {
		value string
		want  []string
	}{
		// only restorable files and visible directories
		{"", []string{"backups" + sep, ".env.bin", "keys.bin", "prod.bin"}},
		{"p", []string{"prod.bin"}},
		{".", []string{".hidden" + sep, ".env.bin"}},
		{"backups" + sep, []string{"backups" + sep + "old.bin"}},
		{"nothing", nil},
	}
	for _, tt := range tests {
		if got := completions(dir, tt.value); !reflect.DeepEqual(got, tt.want) {
			t.Errorf("completions(%q) = %q, want %q", tt.value, got, tt.want)
		}
	}
}

func newCompletionModel(t *testing.T) model {
	ti := textinput.New()
	m := model{defaultFile: ".env.bin", dir: completionFixture(t), input: ti}
	m, _ = m.ask(screenRestoreFile, "", textinput.EchoNormal)
	return m
}

func TestTabExtendsThenCycles(t *testing.T) {
	m := newCompletionModel(t)
	sep := string(filepath.Separator)

	// empty input: nothing in common, so Tab inserts the first match
	m.complete(false)
	if got := m.input.Value(); got != "backups"+sep {
		t.Fatalf("first tab = %q", got)
	}

	// Tab on a fresh prefix completes to the only match
	m.input.SetValue("pr")
	m.refreshMatches()
	m.complete(false)
	if got := m.input.Value(); got != "prod.bin" {
		t.Errorf("single match = %q", got)
	}

	// a prefix shared by several matches is extended first, then cycled
	m.input.SetValue("")
	m.refreshMatches()
	m.complete(false)
	m.complete(false)
	if got := m.input.Value(); got != ".env.bin" {
		t.Errorf("second tab = %q", got)
	}
	m.complete(true)
	if got := m.input.Value(); got != "backups"+sep {
		t.Errorf("shift+tab = %q", got)
	}
}

func TestTabIntoDirectoryListsItsFiles(t *testing.T) {
	m := newCompletionModel(t)
	sep := string(filepath.Separator)

	m.input.SetValue("back")
	m.refreshMatches()
	m.complete(false)
	if got := m.input.Value(); got != "backups"+sep {
		t.Fatalf("value = %q", got)
	}
	if want := []string{"backups" + sep + "old.bin"}; !reflect.DeepEqual(m.matches, want) {
		t.Errorf("matches = %q, want %q", m.matches, want)
	}
}
