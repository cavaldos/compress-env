package tui

import (
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
)

func newTestPicker(t *testing.T, files ...string) picker {
	t.Helper()
	dir := t.TempDir()
	for _, f := range files {
		path := filepath.Join(dir, f)
		if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(path, []byte("A=1\n"), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	p, err := newPicker(dir)
	if err != nil {
		t.Fatal(err)
	}
	return p
}

func rowText(p picker) []string {
	var out []string
	for _, r := range p.rows {
		name := r.n.name
		if r.n.isDir {
			name += "/"
		}
		out = append(out, r.prefix+name)
	}
	return out
}

func TestTreeShowsEnvFilesExpanded(t *testing.T) {
	p := newTestPicker(t, ".env", "notes.txt", "a/.env", "a/b/.env.local", "a/b/readme.md", "c/data.txt")

	want := []string{"a/", "├ b/", "│ └ .env.local", "└ .env", ".env"}
	if got := rowText(p); !reflect.DeepEqual(got, want) {
		t.Errorf("rows = %q, want %q", got, want)
	}
}

func TestAllFilesViewOpensOnlySelectedDirs(t *testing.T) {
	p := newTestPicker(t, ".env", "notes.txt", "a/.env", "c/data.txt")
	p.toggleView()

	// a/ holds a selected file, so it is open; c/ is closed
	want := []string{"a/", "└ .env", "c/", ".env", "notes.txt"}
	if got := rowText(p); !reflect.DeepEqual(got, want) {
		t.Errorf("rows = %q, want %q", got, want)
	}

	p.cursor = 2 // c/
	p.right()
	if got := rowText(p); len(got) != 6 || !strings.HasSuffix(got[3], "data.txt") {
		t.Errorf("after opening c/: %q", got)
	}
	p.right() // already open: step into it
	p.left()  // on a file: jump to the parent
	if p.current().name != "c" {
		t.Errorf("left on a file should select the parent, got %q", p.current().name)
	}
	p.left() // collapse
	if len(p.rows) != 5 {
		t.Errorf("after collapsing: %q", rowText(p))
	}
}

func TestSelectDirectoryTogglesItsFiles(t *testing.T) {
	p := newTestPicker(t, "a/.env", "a/.env.local", "b/.env")
	p.selected = map[string]bool{} // start from nothing

	p.cursor = 0 // a/
	p.toggleSelect()
	if got := p.selection(); !reflect.DeepEqual(got, []string{filepath.Join("a", ".env"), filepath.Join("a", ".env.local")}) {
		t.Errorf("selection = %q", got)
	}
	if s, n := p.state(p.current()); s != 2 || n != 2 {
		t.Errorf("state = %d/%d", s, n)
	}

	p.toggleSelect()
	if got := p.selection(); len(got) != 0 {
		t.Errorf("second toggle should clear: %q", got)
	}
}

func TestOutcome(t *testing.T) {
	p := newTestPicker(t, ".env", "a/.env", "a/.env.local")

	// untouched: the default, nothing to save
	if files, save := p.outcome(); len(files) != 2 || save {
		t.Errorf("default outcome = %q, save=%v", files, save)
	}

	// a different choice is saved
	p.selected[filepath.Join("a", ".env.local")] = true
	if files, save := p.outcome(); len(files) != 3 || !save {
		t.Errorf("custom outcome = %q, save=%v", files, save)
	}

	// nothing selected falls back to the default, nothing to save
	p.toggleAll()
	p.toggleAll()
	p.selected = map[string]bool{}
	if files, save := p.outcome(); len(files) != 2 || save {
		t.Errorf("empty outcome = %q, save=%v", files, save)
	}
}
