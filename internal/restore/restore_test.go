package restore

import (
	"os"
	"path/filepath"
	"testing"
)

func TestApplyKeepsBlankLines(t *testing.T) {
	dir := t.TempDir()
	// The layout written by compress: a header, the content, then a blank line.
	decoded := "--- a.json ---\n{\n\n  \"x\": 1\n}\n\n\n" +
		"--- sub/.env ---\nA=1\n\nB=2\n\n"

	restored, err := Apply(dir, decoded)
	if err != nil {
		t.Fatal(err)
	}
	if len(restored) != 2 {
		t.Fatalf("restored %d files, want 2: %q", len(restored), restored)
	}

	for name, want := range map[string]string{
		"a.json":                     "{\n\n  \"x\": 1\n}\n",
		filepath.Join("sub", ".env"): "A=1\n\nB=2\n",
	} {
		got, err := os.ReadFile(filepath.Join(dir, name))
		if err != nil {
			t.Fatal(err)
		}
		if string(got) != want {
			t.Errorf("%s = %q, want %q", name, got, want)
		}
	}
}

func TestApplySkipsEmptyFiles(t *testing.T) {
	dir := t.TempDir()
	restored, err := Apply(dir, "--- empty ---\n\n\n--- b ---\nB=1\n\n")
	if err != nil {
		t.Fatal(err)
	}
	if len(restored) != 1 {
		t.Errorf("restored = %q, want only b", restored)
	}
}
