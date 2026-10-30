package config

import (
	"os"
	"os/exec"
	"path/filepath"
	"reflect"
	"testing"
)

func write(t *testing.T, dir, name, content string) {
	t.Helper()
	path := filepath.Join(dir, name)
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}
}

func TestDefaultFilesInGitRepo(t *testing.T) {
	if _, err := exec.LookPath("git"); err != nil {
		t.Skip("git not installed")
	}
	dir := t.TempDir()
	if out, err := exec.Command("git", "init", "-q", dir).CombinedOutput(); err != nil {
		t.Fatalf("git init: %v %s", err, out)
	}

	write(t, dir, ".gitignore", "*.local\nsecrets.json\nconfig/*.pem\ndist/\nbin.dat\n.env.bin\ncompress-env.key\nbig.txt\n")
	write(t, dir, ".env", "A=1\n")                 // always chosen
	write(t, dir, "api/.env.local", "B=2\n")       // ignored
	write(t, dir, "secrets.json", "{}\n")          // ignored
	write(t, dir, "config/tls.pem", "-----\n")     // ignored
	write(t, dir, "config/app.yaml", "x: 1\n")     // tracked: not chosen
	write(t, dir, "dist/app.js", "console.log()")  // build output: not chosen
	write(t, dir, "dist/.env", "C=3\n")            // .env: chosen anyway
	write(t, dir, "dist/.env.production", "D=4\n") // env file in an ignored dir
	write(t, dir, "bin.dat", "\x00\x01\x02")       // binary: not chosen
	write(t, dir, ".env.bin", "encoded")           // output: not chosen
	write(t, dir, "compress-env.key", "AGE-SECRET-KEY-1")
	big := make([]byte, maxDefaultSize+1)
	for i := range big {
		big[i] = 'a'
	}
	write(t, dir, "big.txt", string(big)) // too large: not chosen

	candidates, err := Candidates(dir)
	if err != nil {
		t.Fatal(err)
	}
	ig := LoadIgnored(dir)
	if ig == nil {
		t.Fatal("LoadIgnored returned nil inside a git repository")
	}

	got := DefaultFiles(dir, candidates, ig)
	want := []string{
		".env",
		filepath.Join("api", ".env.local"),
		filepath.Join("config", "tls.pem"),
		filepath.Join("dist", ".env"),
		filepath.Join("dist", ".env.production"),
		"secrets.json",
	}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("DefaultFiles = %q\nwant %q", got, want)
	}

	if !ig.InDir(filepath.Join("dist", "")+string(filepath.Separator)) || ig.InDir("config"+string(filepath.Separator)) {
		t.Errorf("InDir: dist/ should be ignored as a whole, config/ should not")
	}
}

func TestDefaultFilesWithoutGit(t *testing.T) {
	dir := t.TempDir()
	write(t, dir, ".env", "A=1\n")
	write(t, dir, "a/.env.local", "B=2\n")
	write(t, dir, "secrets.json", "{}\n")

	candidates, err := Candidates(dir)
	if err != nil {
		t.Fatal(err)
	}
	got := DefaultFiles(dir, candidates, nil)
	if want := []string{".env"}; !reflect.DeepEqual(got, want) {
		t.Errorf("DefaultFiles = %q, want %q", got, want)
	}
}
